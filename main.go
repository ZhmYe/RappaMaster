package main

import (
	"BHCoordinator/Config"
	"BHCoordinator/Coordinator"
	"BHCoordinator/Event"
	"BHCoordinator/LogWriter"
	"BHCoordinator/Network/HTTP"
	"BHCoordinator/Schedule"
	"BHCoordinator/Task"
	"BHCoordinator/paradigm"
	"fmt"
	"time"
)

func main() {
	config := Config.LoadBHCoordinatorConfig("")

	// 初始化 channels
	unprocessedTasks := make(chan paradigm.UnprocessedTask, config.MaxUnprocessedTaskPoolSize)
	//pendingRequestPool := make(chan paradigm.UnprocessedTask, config.MaxHttpRequestPoolSize)
	pendingSchedule := make(chan paradigm.TaskSchedule, config.MaxPendingSchedulePoolSize)
	scheduledTasks := make(chan paradigm.TaskSchedule, config.MaxScheduledTasksPoolSize)
	commitSlots := make(chan paradigm.CommitSlotItem, config.MaxCommitSlotItemPoolSize)
	epochEvent := make(chan bool, 1)
	// 初始化各个组件
	//grpcEngine := Grpc.NewFakeGrpcEngine(pendingSlotPool, pendingSlotRecord)
	//grpcEngine.Setup(*config)
	httpEngine := HTTP.NewFakeHttpEngine(unprocessedTasks)
	httpEngine.Setup(*config)

	event := Event.NewEvent(epochEvent)
	coordinator := Coordinator.NewCoordinator(pendingSchedule, unprocessedTasks, scheduledTasks)

	taskManager := Task.NewTaskManager(*config, scheduledTasks, commitSlots, unprocessedTasks, epochEvent)

	// 初始化 Scheduler
	scheduler := Schedule.NewScheduler(unprocessedTasks, pendingSchedule)
	scheduler.SetTaskManager(taskManager)
	// 配置 Scheduler
	scheduler.Setup(config)
	//scheduler.SetGrpc(grpcEngine)
	scheduler.SetTaskManager(taskManager)

	// 启动各个组件
	//go grpcEngine.Start()
	go httpEngine.Start()
	go taskManager.Start()
	go coordinator.Start()
	go event.Start()
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
	time.Sleep(200 * time.Second)
}
