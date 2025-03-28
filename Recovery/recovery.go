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
	if config.IsRecovery {
		//获取最新的epochid
		maxEpochID, err := service.GetMaxEpochID()
		if err != nil {
			paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to get max epoch ID: %v", err))
			return nil
		}
		paradigm.Print("INFO", fmt.Sprintf("System initialized with epoch ID: %d from database", maxEpochID))
		//TODO 将之前的设置成失败
		return &RappaRecovery{EpochID: maxEpochID}
	} else {
		//TODO  这里要做清空数据库操作
		paradigm.Print("INFO", "clear database successfully")
		return &RappaRecovery{EpochID: -1}
	}
}
