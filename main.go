package main

import (
	"BHLayer2Node/ChainUpper"
	"BHLayer2Node/Collector"
	"BHLayer2Node/Config"
	"BHLayer2Node/Coordinator"
	"BHLayer2Node/Epoch"
	"BHLayer2Node/Event"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/Oracle"
	"BHLayer2Node/Schedule"
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
	//httpEngine := HTTP.NewHttpEngine(rappaChannel)
	event := Event.NewEvent(rappaChannel)
	coordinator := Coordinator.NewCoordinator(rappaChannel)
	epochManager := Epoch.NewEpochManager(rappaChannel)
	chainUpper, _ := ChainUpper.NewMockerChainUpper(rappaChannel) // todo @XQ 测试的时候用的是这个mocker
	oracle := Oracle.NewOracle(rappaChannel)
	collector := Collector.NewCollector(rappaChannel)
	//chainUpper, err := ChainUpper.NewChainUpper(rappaChannel, config)
	//if err != nil {
	//	LogWriter.Log("ERROR", fmt.Sprintf("Failed to initialize ChainUpper: %v", err))
	//}

	// 初始化 Scheduler
	scheduler := Schedule.NewScheduler(rappaChannel)
	// 配置 Scheduler
	//scheduler.Setup(config)
	//scheduler.SetGrpc(grpcEngine)
	//scheduler.SetTaskManager(taskManager)

	// 启动各个组件
	//go grpcEngine.Start()
	go httpEngine.Start()
	go epochManager.Start()
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

	// 主程序保持运行，等待任务完成
	LogWriter.Log("INFO", "Main program is running...")
	//time.Sleep(200 * time.Second)
	oracle.Start()
}
