package test

import (
	"BHLayer2node/Config"
	"BHLayer2node/Network/Grpc"
	"BHLayer2node/paradigm"
	"fmt"
	"testing"
	"time"
)

func TestFakeGrpcEngine(t *testing.T) {
	// 创建 channels 模拟任务的调度和记录
	config := Config.LoadBHLayer2NodeConfig("")

	pendingSlotPool := make(chan paradigm.PendingSlotItem, 5)
	pendingSlotRecord := make(chan paradigm.SlotRecord, 5)

	// 创建 FakeGrpcEngine 实例
	grpcEngine := Grpc.NewFakeGrpcEngine(pendingSlotPool, pendingSlotRecord)

	// 初始化配置
	grpcEngine.Setup(*config)

	// 启动 GrpcEngine
	grpcEngine.Start()

	// 模拟向 PendingSlotPool 添加任务
	go func() {
		for i := 0; i < 3; i++ {
			schedule := []paradigm.BHLayer2NodeSchedule{
				{NID: 0, Size: 100},
				{NID: 1, Size: 100},
				{NID: 2, Size: 100},
			}
			slot := paradigm.PendingSlotItem{
				Sign:     fmt.Sprintf("Task-%d", i+1),
				Slot:     i + 1,
				Size:     300,
				Model:    "ModelA",
				Params:   map[string]interface{}{"param1": "value1"},
				Schedule: schedule,
			}
			pendingSlotPool <- slot
			t.Logf("Submitted PendingSlotItem: %+v", slot)
			time.Sleep(1 * time.Second)
		}
		close(pendingSlotPool)
	}()

	// 检查 PendingSlotRecord 是否接收到任务记录
	go func() {
		for record := range pendingSlotRecord {
			t.Logf("Received SlotRecord: %+v", record)
		}
	}()

	// 等待任务处理完成
	time.Sleep(10 * time.Second)
}
