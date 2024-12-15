package Coordinator

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/Mocker"
	"BHLayer2Node/handler"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// Coordinator，用于协调各部分的运行，作为代码的核心进程运行
// 所有的合成节点作为其follower
// 不断向合成节点发送heartbeatRequest
// heartbeatRequest中包括所有节点的slot分布，用于让合成节点感知到自己的进度，从而修改slot大小
// heartbeatResponse包括合成节点自己的状态，用于状态监控（Monitor）
// 可能包括节点完成的新slot，如果slot字段不为空，则取出交由Tracker处理
// 每隔一段时间，发起SlotVote，从tracker中打包一部分slot（来自各个节点）出来，发送给所有节点进行投票，以bitmap的形式传输投票结果
// 最后获得k票以上的slot们可以被打包，然后将slot和其投票结果元数据等作为交易传递给chainupper准备上链

// Coordinator
// 2024-11-25 先完成schedule
type Coordinator struct {
	pendingSchedules chan paradigm.TaskSchedule    // 等待被发送的调度任务
	unprocessedTasks chan paradigm.UnprocessedTask // 待处理任务
	scheduledTasks   chan paradigm.TaskSchedule    // 已经完成调度的任务
	nodeAddresses    map[int]Config.BHNodeAddress  // 节点地址映射，节点ID -> 地址
  epochHeartbeat   chan *pb.HeartbeatRequest     // 由taskManager构造，大小设置为1, 每个epoch构造一次，epoch只在taskManager中计数即可
	commitSlot       chan paradigm.CommitSlotItem  // 交给taskManager更新
  
	//mockerNodes      []*Mocker.MockerExecutionNode
	mu sync.Mutex // 保护共享数据
}

func (c *Coordinator) Start() {
	// 处理调度,向节点发送调度信息
	processSchedule := func() {
		for schedule := range c.pendingSchedules {
			mapSchedule := make(map[string]int32)
			for _, item := range schedule.Schedules {
				mapSchedule[strconv.Itoa(item.NID)] = item.Size
			}

			// 调用 sendSchedule，这里暂时是统一发给有节点 todo 有必要只给分配的节点?
			c.sendSchedule(schedule.Sign, schedule.Slot, schedule.Size, schedule.Model, schedule.Params, mapSchedule)
		}
	}
	//处理投票

	// 处理心跳
	processHeartbeat := func() {
		for heartbeat := range c.epochHeartbeat {
			c.sendHeartbeat(heartbeat) // 发送心跳
		}
	}
	// 启动协程处理调度任务
	go processSchedule()
	// 启动协程处理心跳
	go processHeartbeat()

	// todo 这里还有一个接收节点的commitSlot的
	// 这里收到了节点commitSlot后，通过channel发送给taskManager(commitSlot)
}

// sendSchedule 向所有节点发送某个sign的调度计划
func (c *Coordinator) sendSchedule(sign string, slot int, size int32, model string, params map[string]interface{}, schedule map[string]int32) {
	// 将 params 转换为 *struct pb.Struct
	convertedParams, err := structpb.NewStruct(params)
	if err != nil {
		LogWriter.Log("ERROR", fmt.Sprintf("Failed to convert params: %v", err))
		panic(err)
	}

	request := pb.ScheduleRequest{
		Sign:     sign,
		Slot:     strconv.Itoa(slot),
		Size:     size,
		Schedule: schedule,
		Model:    model,
		Params:   convertedParams,
	}

	var wg sync.WaitGroup
	successChannel := make(chan paradigm.ScheduleItem, len(c.nodeAddresses)) // 用于统计成功的任务大小

	// 遍历所有节点
	for nodeID, address := range c.nodeAddresses {
		wg.Add(1) // 增加 WaitGroup 计数器
		go func(nodeID int, address string, request *pb.ScheduleRequest) {
			defer wg.Done() // 减少 WaitGroup 计数器

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to connect to node %d at %s: %v", nodeID, address, err))
				return
			}
			defer conn.Close()

			client := pb.NewCoordinatorClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 发送调度请求
			resp, err := client.Schedule(ctx, request)
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
			assignedSize := schedule[resp.NodeId]
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
	remainingSize := size - acceptedSize

	// 输出统计结果
	LogWriter.Log("COORDINATOR", fmt.Sprintf("Schedule '%s' has %d size remaining unaccepted", sign, remainingSize))
	LogWriter.Log("COORDINATOR", fmt.Sprintf("Schedule '%s' total accepted size: %d", sign, acceptedSize))
	// 然后这里把数据放到scheduler重新来
	//newSlot := slot
	if remainingSize == size {
		// 如果所有节点都不接受，直接重新调度
		c.unprocessedTasks <- paradigm.UnprocessedTask{
			Sign:   sign,
			Slot:   slot,
			Size:   size,
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
		c.scheduledTasks <- paradigm.TaskSchedule{
			Sign:      sign,
			Slot:      slot,
			Size:      size,
			Model:     model,
			Params:    params,
			Schedules: acceptSchedules,
		}

	}

}

func (c *Coordinator) sendHeartbeat(heartbeat *pb.HeartbeatRequest) {
	var wg sync.WaitGroup
	disconnected := make(chan int, len(c.ips))                               // 用于说明节点失联
	response := make(chan *pb.HeartbeatResponse, len(c.ips))                 // 用于给voteHandler传递
	voteHandler := handler.NewVoteHandler(heartbeat, c.commitSlot, response) // 处理投票结果
	go voteHandler.Process()                                                 // 准备处理投票
	// 遍历所有节点
	for nodeID, address := range c.ips {
		wg.Add(1) // 增加 WaitGroup 计数器
		go func(nodeID int, address string, heartbeat *pb.HeartbeatRequest) {
			defer wg.Done() // 减少 WaitGroup 计数器

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to connect to node %d at %s: %v", nodeID, address, err))
				disconnected <- nodeID
				return
			}
			defer conn.Close()

			client := pb.NewCoordinatorClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 发送心跳
			resp, err := client.Heartbeat(ctx, heartbeat)
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to send schedule to node %d: %v", nodeID, err))
				//rejectedChannel <- 0 // 默认统计为未接受
				disconnected <- nodeID // 暂定为当作失联
				return
			}
			response <- resp // 交给voteHandler处理

		}(nodeID, address, heartbeat)
	}

	// 等待所有节点处理完成
	wg.Wait()
	close(response)     // 关闭response，voteHandler开始处理投票结果
	close(disconnected) // 关闭disconnected, 后续monitor结束处理 todo

}

func NewCoordinator(config *Config.BHLayer2NodeConfig, pendingSchedules chan paradigm.TaskSchedule, unprocessedTasks chan paradigm.UnprocessedTask, scheduledTasks chan paradigm.TaskSchedule, commitSlot chan paradigm.CommitSlotItem, epochHeartbeat chan *pb.HeartbeatRequest) *Coordinator {
	// 加载配置中的节点IP
	c := Coordinator{
		pendingSchedules: pendingSchedules,
		unprocessedTasks: unprocessedTasks,
		scheduledTasks:   scheduledTasks,
		nodeAddresses:    config.BHNodeAddressMap,
		//mockerNodes:      make([]*Mocker.MockerExecutionNode, 0),
		commitSlot:       commitSlot,
		epochHeartbeat:   epochHeartbeat,
		mu:               sync.Mutex{},
	}
	//for i, ip := range c.nodeAddresses {
	//	c.mockerNodes = append(c.mockerNodes, Mocker.NewMockerExecutionNode(i, ip, commitSlot))
	//}
	return &c
}
