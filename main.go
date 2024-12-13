package main

import (
	"BHLayer2Node/ChainUpper"
	"BHLayer2Node/Config"
	"BHLayer2Node/Coordinator"
	"BHLayer2Node/Event"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/Schedule"
	"BHLayer2Node/Task"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"
)

func main() {
	config := Config.LoadBHLayer2NodeConfig("config.json")

	// 初始化 channels
	initTasks := make(chan paradigm.UnprocessedTask, config.MaxUnprocessedTaskPoolSize)
	unprocessedTasks := make(chan paradigm.UnprocessedTask, config.MaxUnprocessedTaskPoolSize)
	//pendingRequestPool := make(chan paradigm.UnprocessedTask, config.MaxHttpRequestPoolSize)
	pendingSchedule := make(chan paradigm.TaskSchedule, config.MaxPendingSchedulePoolSize)
	scheduledTasks := make(chan paradigm.TaskSchedule, config.MaxScheduledTasksPoolSize)
	commitSlots := make(chan paradigm.CommitSlotItem, config.MaxCommitSlotItemPoolSize)

	//slotToVotes := make(chan paradigm.CommitSlotItem, config.MaxCommitSlotItemPoolSize)
	pendingTransactions := make(chan paradigm.Transaction, config.MaxCommitSlotItemPoolSize) // todo
	epochEvent := make(chan bool, 1)

	// 初始化各个组件
	//grpcEngine := Grpc.NewFakeGrpcEngine(pendingSlotPool, pendingSlotRecord)
	//grpcEngine.Setup(*config)
	httpEngine := HTTP.NewFakeHttpEngine(unprocessedTasks, initTasks)
	httpEngine.Setup(*config)

	event := Event.NewEvent(epochEvent)
	coordinator := Coordinator.NewCoordinator(config, pendingSchedule, unprocessedTasks, scheduledTasks, commitSlots)

	taskManager := Task.NewTaskManager(*config, scheduledTasks, commitSlots, unprocessedTasks, epochEvent, initTasks, pendingTransactions)

	chainUpper, err := ChainUpper.NewChainUpper(pendingTransactions, config)
	if err != nil {
		LogWriter.Log("ERROR", fmt.Sprintf("Failed to initialize ChainUpper: %v", err))
	}

	// 初始化 Scheduler
	scheduler := Schedule.NewScheduler(unprocessedTasks, pendingSchedule)
	// 配置 Scheduler
	scheduler.Setup(config)
	//scheduler.SetGrpc(grpcEngine)
	scheduler.SetTaskManager(taskManager)

	// 启动各个组件
	//go grpcEngine.Start()
	go httpEngine.Start()
	go taskManager.Start()
	go coordinator.Start()
	//定时，如果大于10s,EpochEvent队列里放置一个true
	go event.Start()
	//上链
	go chainUpper.Start()
	// 启动 Scheduler
	if err := scheduler.Start(); err != nil {
		LogWriter.Log("ERROR", fmt.Sprintf("Failed to start scheduler: %v", err))
		return
	}

	//// 模拟前端 HTTP 请求
	//go func() {
	//	for i := 0; i < 5; i++ {
	//		request := paradigm.UnprocessedTask{
	//			Sign:   fmt.Sprintf("Task-%d", i+1),
	//			Size:   1000,
	//			Model:  "TestModel",
	//			Params: map[string]interface{}{"param1": "value1"},
	//		}
	//		pendingRequestPool <- request
	//		LogWriter.Log("DEBUG", fmt.Sprintf("Submitted UnprocessedTask: %+v", request))
	//		time.Sleep(2 * time.Second)
	//	}
	//}()

	// 主程序保持运行，等待任务完成
	LogWriter.Log("INFO", "Main program is running...")
	//time.Sleep(200 * time.Second)
	timeStart := time.Now()
	for {
		if time.Since(timeStart) >= 200*time.Second {
			break
		}
	}
}
