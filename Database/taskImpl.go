package Database

import (
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// 获取任务
func (o DatabaseService) GetTask(taskHash paradigm.TaskHash) (*paradigm.Task, error) {
	taskQuery := paradigm.Task{}
	err := o.db.Where(paradigm.Task{Sign: taskHash}).Take(&taskQuery).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else {
		// 获取txHash
		tx := paradigm.DevReference{}
		o.db.Take(&tx, taskQuery.TID)
		taskQuery.TxReceipt = &tx.TxReceipt
		taskQuery.TxBlockHash = tx.TxBlockHash
		taskQuery.TxHash = tx.TxHash
		return &taskQuery, nil
	}
}

// 更新任务schedule
func (o DatabaseService) UpdateScheduleInTask(schedule *paradigm.SynthTaskSchedule) {
	task, err := o.GetTask(schedule.TaskID)
	if err != nil {
		panic(fmt.Sprintf("task not found of %s", schedule.TaskID))
	}
	task.Schedules = append(task.Schedules, schedule)
	task.ScheduleMap[schedule.ScheduleID] = len(task.Schedules)
	o.db.Model(task).Select("schedules", "schedule_map").Updates(task)
}

// 创建任务
func (o DatabaseService) SetTask(task *paradigm.Task) {
	o.db.Omit("end_time").Create(task)
}

// 更新任务
func (o DatabaseService) UpdateTask(task *paradigm.Task) {
	o.db.Model(task).Updates(task)
}

// GetTaskByID 通过任务标识查询任务
func (o DatabaseService) GetTaskByID(taskID string) (*paradigm.Task, error) {
	var task paradigm.Task
	err := o.db.Where("sign = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}

	tx := paradigm.DevReference{}
	if err = o.db.Take(&tx, task.TID).Error; err != nil {
		return nil, fmt.Errorf("failed to get associated transaction: %v", err)
	}
	task.TxReceipt = &tx.TxReceipt
	task.TxBlockHash = tx.TxBlockHash
	task.TxHash = tx.TxHash

	// 更新每个schedule中的slots信息
	for i, schedule := range task.Schedules {
		// 为每个schedule创建新的slots切片
		var updatedSlots []*paradigm.Slot

		// 查询该schedule下的所有slots
		for _, slot := range schedule.Slots {
			var dbSlot paradigm.Slot
			if err := o.db.Where("slot_id = ?", slot.SlotID).First(&dbSlot).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 如果slot不存在，保留原有slot
					updatedSlots = append(updatedSlots, slot)
				} else {
					return nil, fmt.Errorf("failed to query slot %s: %v", slot.SlotID, err)
				}
			} else {
				// 使用数据库中的最新slot信息
				updatedSlots = append(updatedSlots, &dbSlot)
			}
		}

		// 更新schedule的slots
		task.Schedules[i].Slots = updatedSlots
	}

	return &task, nil
}

// GetTaskByTxHash 通过交易哈希查询任务
func (o DatabaseService) GetTaskByTxHash(txHash string) (*paradigm.Task, error) {
	tx, err := o.GetTransactionByHash(txHash)
	if err != nil {
		return nil, err
	}

	if tx.Rf != paradigm.InitTaskTx {
		return nil, fmt.Errorf("transaction is not an init task transaction")
	}

	return o.GetTaskByID(tx.TaskID)
}

// GetAllTasks 查询所有任务
func (o DatabaseService) GetAllTasks() (map[string]*paradigm.Task, error) {
	var tasks []*paradigm.Task
	err := o.db.Order("start_time DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	tasksMap := make(map[string]*paradigm.Task)
	for _, task := range tasks {
		tx := paradigm.DevReference{}
		if err := o.db.Take(&tx, task.TID).Error; err == nil {
			task.TxReceipt = &tx.TxReceipt
			task.TxBlockHash = tx.TxBlockHash
			task.TxHash = tx.TxHash
		}
		tasksMap[task.Sign] = task
	}
	return tasksMap, nil
}

// GetSynthDataByModel 综合数据查询实现
func (o DatabaseService) GetSynthDataByModel() (map[paradigm.SupportModelType]int32, error) {
	// 创建结果map
	synthData := make(map[paradigm.SupportModelType]int32)

	// 查询所有任务
	var tasks []*paradigm.Task
	err := o.db.Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %v", err)
	}

	// 按模型类型统计已完成任务的处理量
	for _, task := range tasks {
		// 使用 IsFinish() 方法判断任务是否完成
		if task.IsFinish() {
			if currentValue, exists := synthData[task.Model]; exists {
				synthData[task.Model] = currentValue + task.Process
			} else {
				synthData[task.Model] = task.Process
			}
		}
	}

	return synthData, nil
}
