package Recovery

import (
	"BHLayer2Node/Database"
	"BHLayer2Node/paradigm"
	"fmt"
)

type RappaRecovery struct {
	EpochID int32 //当前epoch数
}

func RecoverFromDataBase(config *paradigm.BHLayer2NodeConfig, service *Database.DatabaseService) *RappaRecovery {
	if config.IsAutoMigrate {
		if err := service.AutoMigrate(); err != nil {
			paradigm.Error(paradigm.DatabaseError, fmt.Sprintf("auto migrate failed: %v", err))
			return nil
		}
	}
	if config.IsRecovery {
		//获取最新的epochid
		maxEpochID, err := service.GetMaxEpochID()
		if err != nil {
			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to get max epoch ID: %v", err))
			return nil
		}
		paradigm.Print("INFO", fmt.Sprintf("System initialized with epoch ID: %d from database", maxEpochID))
		// 将之前未完成的任务设置成失败
		err = processUnFinished(maxEpochID, service)
		if err != nil {
			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to process unFinishedTask: %v", err))
			return nil
		}
		return &RappaRecovery{EpochID: maxEpochID}
	} else {
		//这里要做清空数据库操作
		if err := service.TruncateAll(); err != nil {
			paradigm.Error(paradigm.DatabaseError, fmt.Sprintf("truncate failed: %v", err))
			return nil
		}
		paradigm.Print("INFO", "clear database successfully")
		return &RappaRecovery{EpochID: -1}
	}
}

func processUnFinished(epoch int32, service *Database.DatabaseService) error {
	newestTime, err := service.GetNewestDateTime()
	if err != nil {
		return err
	}
	tasks, err := service.GetTasksAfter(newestTime)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		service.DownUnFinishedSlotsByTaskID(epoch, task.Sign)
	}
	return nil
}
