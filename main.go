package main

import (
	"BHLayer2Node/ChainUpper"
	"BHLayer2Node/Collector"
	"BHLayer2Node/Config"
	"BHLayer2Node/Coordinator"
	"BHLayer2Node/Dev"
	"BHLayer2Node/Event"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/Schedule"
	"BHLayer2Node/Task"
	"BHLayer2Node/paradigm"
	"fmt"
)

func main() {
	config := Config.LoadBHLayer2NodeConfig("config.json")

	rappaChannel := paradigm.NewRappaChannel(config)
	// 初始化各个组件
	//grpcEngine := Grpc.NewFakeGrpcEngine(pendingSlotPool, pendingSlotRecord)
	//grpcEngine.Setup(*config)
	httpEngine := HTTP.NewFakeHttpEngine(rappaChannel)
	httpEngine.Setup(*config)
	event := Event.NewEvent(rappaChannel)
	coordinator := Coordinator.NewCoordinator(rappaChannel)
	taskManager := Task.NewTaskManager(rappaChannel)
	// chainUpper, _ := ChainUpper.NewMockerChainUpper(rappaChannel) // todo @XQ 测试的时候用的是这个mocker
	dev := Dev.NewDev(rappaChannel)
	collector := Collector.NewCollector(rappaChannel)
	chainUpper, _ := ChainUpper.NewChainUpper(rappaChannel, config)
	chainQuery, _ := ChainUpper.NewChainQuery(rappaChannel, config)

	// 初始化 Scheduler
	scheduler := Schedule.NewScheduler(rappaChannel)
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

	go chainUpper.Start()
	go collector.Start()
	// 启动 Scheduler
	if err := scheduler.Start(); err != nil {
		LogWriter.Log("ERROR", fmt.Sprintf("Failed to start scheduler: %v", err))
		return
	}
	chainQuery.Start()

	// 主程序保持运行，等待任务完成
	LogWriter.Log("INFO", "Main program is running...")
	//time.Sleep(200 * time.Second)
	dev.Start()
}
