package main

import (
	"BHLayer2node/Config"
	"BHLayer2node/LogWriter"
	"BHLayer2node/Network/Grpc"
	"BHLayer2node/Network/HTTP"
	"BHLayer2node/Schedule"
	"BHLayer2node/Tracker"
	"BHLayer2node/paradigm"
	"fmt"
	"time"
)

func main() {
	config := Config.LoadBHLayer2NodeConfig("")

	// 初始化 channels
	pendingRequestPool := make(chan paradigm.HttpTaskRequest, config.MaxHttpRequestPoolSize)
	pendingSlotPool := make(chan paradigm.PendingSlotItem, config.MaxGrpcRequestPoolSize)
	pendingSlotRecord := make(chan paradigm.SlotRecord, config.MaxGrpcRequestPoolSize)

	// 初始化各个组件
	grpcEngine := Grpc.NewFakeGrpcEngine(pendingSlotPool, pendingSlotRecord)
	grpcEngine.Setup(*config)
	httpEngine := HTTP.NewFakeHttpEngine(pendingRequestPool)
	httpEngine.Setup(*config)
	tracker := Tracker.NewTracker(*config)

	// 初始化 Scheduler
	scheduler := Schedule.NewScheduler(pendingSlotPool, pendingRequestPool, pendingSlotRecord)

	// 配置 Scheduler
	scheduler.Setup(config)
	scheduler.SetGrpc(grpcEngine)
	scheduler.SetTracker(tracker)

	// 启动各个组件
	go grpcEngine.Start()
	go httpEngine.Start()

	// 启动 Scheduler
	if err := scheduler.Start(); err != nil {
		LogWriter.Log("ERROR", fmt.Sprintf("Failed to start scheduler: %v", err))
		return
	}

	//// 模拟前端 HTTP 请求
	//go func() {
	//	for i := 0; i < 5; i++ {
	//		request := paradigm.HttpTaskRequest{
	//			Sign:   fmt.Sprintf("Task-%d", i+1),
	//			Size:   1000,
	//			Model:  "TestModel",
	//			Params: map[string]interface{}{"param1": "value1"},
	//		}
	//		pendingRequestPool <- request
	//		LogWriter.Log("DEBUG", fmt.Sprintf("Submitted HttpTaskRequest: %+v", request))
	//		time.Sleep(2 * time.Second)
	//	}
	//}()

	// 主程序保持运行，等待任务完成
	LogWriter.Log("INFO", "Main program is running...")
	time.Sleep(200 * time.Second)
}
