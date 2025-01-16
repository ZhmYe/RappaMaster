package Coordinator

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"strconv"
	"sync"
	"time"
)

// sendSchedule 向所有节点发送某个sign的调度计划
func (c *Coordinator) sendSchedule(sign string, slot int32, totalSize int32, model paradigm.SupportModelType, params map[string]interface{}, schedule map[int]int32) {
	nodeAddresses := c.connManager.GetNodeAddresses()
	// 将 params 转换为 *struct pb.Struct
	convertedParams, err := structpb.NewStruct(params)
	if err != nil {
		LogWriter.Log("ERROR", fmt.Sprintf("Failed to convert params: %v", err))
		panic(err)
	}

	var wg sync.WaitGroup
	successChannel := make(chan paradigm.ScheduleItem, len(nodeAddresses)) // 用于统计成功的任务大小
	wg.Add(len(nodeAddresses))                                             // 增加 WaitGroup 计数器
	// 遍历所有节点
	for nodeID, address := range nodeAddresses {
		// TODO @YZM 这里暂时先这样写了，就是给每个slot一个标识
		computeScheduleHash := func(nodeID int) paradigm.SlotHash {
			return fmt.Sprintf("%s_%d_%d", sign, slot, nodeID)
		}
		request := pb.ScheduleRequest{
			Sign:   sign,
			Slot:   slot,
			Size:   schedule[nodeID],
			NodeID: int32(nodeID),
			Model:  paradigm.ModelTypeToString(model),
			Params: convertedParams,
			Hash:   computeScheduleHash(nodeID),
		}
		go func(nodeID int, address string, request *pb.ScheduleRequest) {
			defer wg.Done() // 减少 WaitGroup 计数器

			// 建立grpc连接
			conn, err := c.connManager.GetConn(nodeID)
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to connect to node %d at %s: %v", nodeID, address, err))
				return
			}
			client := pb.NewRappaExecutorClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 发送调度请求
			resp, err := client.Schedule(ctx, request, grpc.WaitForReady(true))
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to send schedule to node %d: %v", nodeID, err))
				//rejectedChannel <- 0 // 默认统计为未接受
				return
			}

			// 校验任务标识
			if resp.Sign != sign {
				LogWriter.Log("ERROR", fmt.Sprintf("Task Sign does not match: %s != %s", resp.Sign, sign))
				//rejectedChannel <- 0 // 默认统计为未接受
				return
			}

			// 根据节点反馈更新统计
			assignedSize := schedule[nodeID]
			nID, _ := strconv.Atoi(resp.NodeId)
			if resp.Accept {
				LogWriter.Log("COORDINATOR", fmt.Sprintf("Node %s accepted schedule: %v", resp.NodeId, resp.Sign))
				successChannel <- paradigm.ScheduleItem{
					Size: assignedSize,
					NID:  nID,
				}
			} else {
				LogWriter.Log("ERROR", fmt.Sprintf("Node %s rejected schedule: %v, reason: %s", resp.NodeId, resp.Sign, resp.ErrorMessage))
				//rejectedChannel <- assignedSize
			}
		}(nodeID, address.GetAddrStr(), &request)
	}

	// 等待所有节点处理完成
	wg.Wait()
	close(successChannel)
	//close(rejectedChannel)

	// 统计结果
	acceptedSize := int32(0)
	acceptSchedules := make([]paradigm.ScheduleItem, 0)
	for item := range successChannel {
		acceptedSize += item.Size
		acceptSchedules = append(acceptSchedules, item)
	}
	remainingSize := totalSize - acceptedSize
	if remainingSize < 0 {
		remainingSize = 0
	}
	// 输出统计结果
	LogWriter.Log("COORDINATOR", fmt.Sprintf("Schedule '%s' has %d size remaining unaccepted", sign, remainingSize))
	LogWriter.Log("COORDINATOR", fmt.Sprintf("Schedule '%s' total accepted size: %d", sign, acceptedSize))
	// 然后这里把数据放到scheduler重新来
	//newSlot := slot
	if remainingSize == totalSize {
		// 如果所有节点都不接受，直接重新调度
		c.channel.UnprocessedTasks <- paradigm.UnprocessedTask{
			Sign:   sign,
			Slot:   slot,
			Size:   totalSize,
			Model:  model,
			Params: params,
		}
		LogWriter.Log("WARNING", fmt.Sprintf("No node accept schedules, restart the task %s slot %d scheduling...", sign, slot))
	} else {
		// 如果有节点接受，那么如果节点有反馈，那么在反馈处更新unprocessedTask
		// 如果没有反馈，那么有额外处理 todo
		// 认为这是一个合法的slot
		LogWriter.Log("COORDINATOR", fmt.Sprintf("Successfully schedule the task %s slot %d, Waiting for result...", sign, slot))
		// 这是最后真正的schedule,由tracker获取
		c.channel.ScheduledTasks <- paradigm.TaskSchedule{
			Sign:      sign,
			Slot:      slot,
			Size:      totalSize,
			Model:     model,
			Params:    params,
			Schedules: acceptSchedules,
		}

	}

}
