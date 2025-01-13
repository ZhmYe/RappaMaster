package Coordinator

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/Network/Grpc"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"sync"
)

// Coordinator，用于协调各部分的运行，作为代码的核心进程运行
// 所有的合成节点作为其follower
// 不断向合成节点发送heartbeatRequest
// heartbeatRequest中包括所有节点的slot分布，用于让合成节点感知到自己的进度，从而修改slot大小
// heartbeatResponse包括合成节点自己的状态，用于状态监控（Monitor）
// 可能包括节点完成的新slot，如果slot字段不为空，则取出交由Tracker处理
// 每隔一段时间，发起SlotVote，从tracker中打包一部分slot（来自各个节点）出来，发送给所有节点进行投票，以bitmap的形式传输投票结果
// 最后获得k票以上的slot们可以被打包，然后将slot和其投票结果元数据等作为交易传递给chainupper准备上链

type Coordinator struct {
	pendingSchedules      chan paradigm.TaskSchedule      // 等待被发送的调度任务
	unprocessedTasks      chan paradigm.UnprocessedTask   // 待处理任务
	scheduledTasks        chan paradigm.TaskSchedule      // 已经完成调度的任务
	maxEpochDelay         int                             //说明多少时间后需要传递一个zkp证明以及多久后开始投票
	connManager           *Grpc.NodeGrpcManager           //用于管理GRPC客户端连接
	serverPort            int                             // coordinator对节点暴露的grpc端口
	epochHeartbeat        chan *pb.HeartbeatRequest       // 由taskManager构造，大小设置为1, 每个epoch构造一次，epoch只在taskManager中计数即可
	commitSlot            chan paradigm.CommitSlotItem    // 交给taskManager更新
	recoverRequestChannel chan paradigm.RecoverConnection // CollectInstance确认收集哪些slot以后发给coordinator进行分发
	//mockerNodes []*Mocker.MockerExecutionNode
	mu sync.Mutex // 保护共享数据

	pb.UnimplementedRappaMasterServer
}

func (c *Coordinator) Start() {
	// 处理调度,向节点发送调度信息
	processSchedule := func() {
		for schedule := range c.pendingSchedules {
			mapSchedule := make(map[int]int32)
			for _, item := range schedule.Schedules {
				mapSchedule[item.NID] = item.Size
			}

			// 调用 sendSchedule，这里暂时是统一发给有节点 todo 有必要只给分配的节点?
			c.sendSchedule(schedule.Sign, schedule.Slot, schedule.Size, schedule.Model, schedule.Params, mapSchedule)
		}
	}

	// 处理心跳
	processHeartbeat := func() {
		for heartbeat := range c.epochHeartbeat {
			c.sendHeartbeat(heartbeat) // 发送心跳
		}
	}

	// 处理commitslot
	processCommitSlot := func() {
		// 监听指定端口
		lis, err := net.Listen("tcp", ":"+strconv.Itoa(c.serverPort))
		if err != nil {
			LogWriter.Log("ERROR", fmt.Sprintf("Failed to listen: %v", err))
		}
		server := grpc.NewServer()
		pb.RegisterRappaMasterServer(server, c)
		LogWriter.Log("DEBUG", fmt.Sprintf("Coordinator gRPC server is running on :%d", c.serverPort))
		if err := server.Serve(lis); err != nil {
			LogWriter.Log("ERROR", fmt.Sprintf("Failed to serve: %v", err))
		}

	}

	processCollect := func() {
		for recoverRequest := range c.recoverRequestChannel {
			// todo @YZM 这里通过grpc发给节点
			// 假装发一下，然后返回一下response
			//index := 0
			//responseChannel := recoverRequest.ResponseChannel
			//LogWriter.Log("DEBUG", "Send Recover Request to nodes...")
			//for _, hash := range recoverRequest.Hashs {
			//	response := paradigm.RecoverResponse{
			//		SlotHash: hash,
			//		Data:     []byte(fmt.Sprintf("%d", index)),
			//	}
			//	index++
			//	responseChannel <- response
			//}
			c.sendCollect(recoverRequest)
			LogWriter.Log("DEBUG", "Receive All Recover Responses from nodes...")
			//close(responseChannel)
		}
	}
	// 启动协程处理调度任务
	go processSchedule()
	// 启动协程处理心跳
	go processHeartbeat()
	// 这里收到了节点commitSlot后，通过channel发送给taskManager(commitSlot)
	go processCommitSlot()

	// 收到了来自collector的收集任务，发起collect任务
	go processCollect()

	//for _, node := range c.mockerNodes {
	//	go node.Start()
	//}
}

func NewCoordinator(config *Config.BHLayer2NodeConfig, channel *paradigm.RappaChannel) *Coordinator {
	// 加载配置中的节点IP
	c := Coordinator{
		pendingSchedules: channel.PendingSchedule,
		unprocessedTasks: channel.UnprocessedTasks,
		scheduledTasks:   channel.ScheduledTasks,
		maxEpochDelay:    config.MaxEpochDelay,
		connManager:      Grpc.NewNodeGrpcManager(config.BHNodeAddressMap),
		serverPort:       config.GrpcPort,
		//mockerNodes:      make([]*Mocker.MockerExecutionNode, 0),
		commitSlot:            channel.CommitSlots,
		epochHeartbeat:        channel.EpochHeartbeat,
		recoverRequestChannel: channel.SlotCollectChannel,
		mu:                    sync.Mutex{},
	}
	//for i, address := range c.nodeAddresses {
	//	c.mockerNodes = append(c.mockerNodes, Mocker.NewMockerExecutionNode(i, address.GetAddrStr(), commitSlot))
	//}
	return &c
}
