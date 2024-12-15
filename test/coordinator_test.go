package test

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/Coordinator"
	"BHLayer2Node/paradigm"
	"BHLayer2Node/utils"
	"os"
	"path/filepath"
	"testing"
)

var testConfig *Config.BHLayer2NodeConfig

// 用于测试初始化
func TestMain(m *testing.M) {
	rootPath, _ := utils.GetProjectRoot()
	testConfig = Config.LoadBHLayer2NodeConfig(filepath.Join(rootPath, "config.json"))
	// 执行所有测试
	exitCode := m.Run()
	// 显式退出
	os.Exit(exitCode)
}

// 测试 Coordinatror GRPC
func TestCoordinator(t *testing.T) {
	// 初始化 channels
	unprocessedTasks := make(chan paradigm.UnprocessedTask, testConfig.MaxUnprocessedTaskPoolSize)
	pendingSchedule := make(chan paradigm.TaskSchedule, testConfig.MaxPendingSchedulePoolSize)
	scheduledTasks := make(chan paradigm.TaskSchedule, testConfig.MaxScheduledTasksPoolSize)
	commitSlots := make(chan paradigm.CommitSlotItem, testConfig.MaxCommitSlotItemPoolSize)
	coordinator := Coordinator.NewCoordinator(testConfig, pendingSchedule, unprocessedTasks, scheduledTasks, commitSlots)
	go coordinator.Start()
	pendingSchedule <- paradigm.TaskSchedule{
		Sign:  "FakeSign",
		Slot:  0,
		Size:  50,
		Model: "ctgan",
		Params: map[string]interface{}{
			"condition_column": "native-country",
			"condition_value":  "United-States",
		},
		Schedules: []paradigm.ScheduleItem{
			{
				NID:  0,
				Size: 25,
			},
			{
				NID:  1,
				Size: 25,
			},
		},
	}
	result, ok := <-scheduledTasks
	if !ok || len(result.Schedules) != 2 {
		t.Errorf("failed task!")
	}
}
