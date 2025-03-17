package Coordinator

import (
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
func (c *Coordinator) sendSchedule(schedule paradigm.SynthTaskSchedule) {
	nodeAddresses := c.connManager.GetNodeAddresses()
	// 将 params 转换为 *struct pb.Struct
	convertedParams, err := structpb.NewStruct(schedule.Params)
	if err != nil {
		paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to convert Schedule params: %v", err))
		//paradigm.Log("ERROR", fmt.Sprintf("Failed to convert params: %v", err))
		//panic(e.Error())
		return
	}

	var wg sync.WaitGroup
	successChannel := make(chan *paradigm.Slot, len(nodeAddresses)) // 用于统计成功的任务大小
	rejectChannel := make(chan [2]interface{}, len(nodeAddresses))  // 用于统计失败的任务
	wg.Add(len(nodeAddresses))                                      // 增加 WaitGroup 计数器
	// 遍历所有节点
	for nID, index := range schedule.NodeIDMap {
		//for nodeID, address := range nodeAddresses {
		// TODO @YZM 这里暂时先这样写了，就是给每个slot一个标识
		//computeScheduleHash := func(nodeID int) paradigm.SlotHash {
		//	return fmt.Sprintf("%s_%d_%d", sign, slot, nodeID)

		address := nodeAddresses[nID]
		slot := schedule.Slots[index]
		request := pb.ScheduleRequest{
			Sign:   schedule.TaskID,
			Slot:   schedule.ScheduleID,
			Size:   slot.ScheduleSize,
			NodeID: int32(nID),
			Model:  paradigm.ModelTypeToString(schedule.Model),
			Params: convertedParams,
			Hash:   slot.SlotID,
		}
		//tmp := slot
		go func(nodeID int, address string, request *pb.ScheduleRequest) {
			defer wg.Done() // 减少 WaitGroup 计数器

			// 建立grpc连接
			conn, err := c.connManager.GetConn(nodeID)
			if err != nil {
				e := paradigm.Error(paradigm.ExecutorError, fmt.Sprintf("Failed to connect to node %d at %s: %v", nodeID, address, err))
				slot.SetError(e.Error())
				rejectChannel <- [2]interface{}{nodeID, e.Error()}
				return
			}
			client := pb.NewRappaExecutorClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 发送调度请求
			resp, err := client.Schedule(ctx, request, grpc.WaitForReady(true))
			if err != nil {
				e := paradigm.Error(paradigm.ExecutorError, fmt.Sprintf("Failed to send schedule to node %d: %v", nodeID, err))
				//slot.Status = paradigm.Failed
				//slot.SetError(errorMessage)
				rejectChannel <- [2]interface{}{nodeID, e.Error()}
				return
			}

			// 校验任务标识
			if resp.Sign != schedule.TaskID {
				e := paradigm.Error(paradigm.ExecutorError, fmt.Sprintf("Epoch Sign does not match: %s != %s", resp.Sign, schedule.TaskID))
				//slot.SetError(errorMessage)
				rejectChannel <- [2]interface{}{nodeID, e.Error()}
				return
			}
			nID, _ := strconv.Atoi(resp.NodeId)
			if nID != nodeID {
				e := paradigm.Error(paradigm.ExecutorError, fmt.Sprintf("NodeID does not match: %s != %d", resp.NodeId, nodeID))
				//slot.SetError(errorMessage)
				rejectChannel <- [2]interface{}{nodeID, e.Error()}
				return
			}

			// 根据节点反馈更新统计
			//assignedSize := request.Size
			//nID, _ := strconv.Atoi(resp.NodeId)
			if resp.Accept {
				//paradigm.Log("COORDINATOR", fmt.Sprintf("Node %s accepted schedule: %v", resp.NodeId, resp.Sign))
				successChannel <- slot

			} else {
				//errorMessage := fmt.Sprintf("Node %s rejected schedule: %v, reason: %s", resp.NodeId, resp.Sign, resp.ErrorMessage)
				e := paradigm.Error(paradigm.ExecutorError, fmt.Sprintf("Node %s rejected schedule: %v, reason: %s", resp.NodeId, resp.Sign, resp.ErrorMessage))
				//slot.SetError(errorMessage)
				rejectChannel <- [2]interface{}{nodeID, e.Error()}
				//rejectedChannel <- assignedSize
			}
		}(nID, address.GetAddrStr(), &request)
	}

	// 等待所有节点处理完成
	wg.Wait()
	close(successChannel)
	close(rejectChannel)
	//
	// 统计结果
	acceptedSize := int32(0)
	////acceptSchedules := make([]*paradigm.Slot, 0)
	for item := range successChannel {
		acceptedSize += item.ScheduleSize
		//	//acceptSchedules = append(acceptSchedules, item)
	}
	rejectNumber := 0
	for item := range rejectChannel {
		nID, errorMessage := item[0].(int), item[1].(string)
		rejectNumber++
		//schedule.Slots[nID]
		index := schedule.NodeIDMap[nID]
		schedule.Slots[index].SetError(errorMessage) // 更新失败的slot
	}
	remainingSize := schedule.Size - acceptedSize
	if remainingSize < 0 {
		remainingSize = 0
	}
	//输出统计结果
	paradigm.Print("COORDINATOR", fmt.Sprintf("Schedule '%s' has %d size remaining unaccepted, total accept size: %d", schedule.TaskID, remainingSize, acceptedSize))
	//paradigm.Print("COORDINATOR", fmt.Sprintf("Schedule '%s' total accepted size: %d", schedule.TaskID, acceptedSize))
	// 然后这里把数据放到scheduler重新来
	//newSlot := slot
	if remainingSize == schedule.Size {
		// 如果所有节点都不接受，直接重新调度
		c.channel.UnprocessedTasks <- paradigm.UnprocessedTask{
			TaskID: schedule.TaskID,
			Size:   schedule.Size,
			Model:  schedule.Model,
			Params: schedule.Params,
		}
		paradigm.Print("WARNING", fmt.Sprintf("No node accept schedules, restart the task %s scheduling...", schedule.TaskID))
	} else {
		// 如果有节点接受，那么如果节点有反馈，那么在反馈处更新unprocessedTask
		// 如果没有反馈，那么有额外处理 todo
		// 认为这是一个合法的slot
		paradigm.Print("COORDINATOR", fmt.Sprintf("Successfully schedule the task %s", schedule.TaskID))
		// 这是最后真正的schedule,由tracker获取
		//schedule.Print()
		//for nodeID, slot := range schedule.Slots {
		//	fmt.Println(slot.SlotID)
		//	fmt.Println(slot.ScheduleSize)
		//	fmt.Println(slot.Status)
		//	fmt.Println(nodeID)
		//}
		c.channel.ScheduledTasks <- schedule
		c.channel.OracleSchedules <- &schedule

	}

}
