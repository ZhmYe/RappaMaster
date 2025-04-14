package Database

import (
	"BHLayer2Node/Collector"
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
	o.db.Model(task).Select("*").Updates(task)
}

func (o DatabaseService) IncrementTaskProcess(taskSign string, slot *paradigm.CommitRecord) error {
	// 使用原子操作增加任务进度
	if slot.State() != paradigm.FINALIZE {
		return fmt.Errorf("the commit Slot is not finalized") // 只能提交finalized的，因为已经通过投票了所以不需要check
	}
	query := "UPDATE tasks SET process = process + ? WHERE sign = ?"
	result := o.db.Exec(query, slot.Process, taskSign)
	if result.Error != nil {
		return fmt.Errorf("failed to increment task process: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("task with sign %s not found or no update needed", taskSign)
	}
	// debug
	paradigm.Log("DEBUG", fmt.Sprintf("Atomically incremented task %s process by %d", taskSign, slot.Process))
	return nil
}

// 增量更新并返回更新后的任务
func (o DatabaseService) IncrementTaskProcessAndGet(taskSign string, slot *paradigm.CommitRecord) (*paradigm.Task, error) {
	err := o.IncrementTaskProcess(taskSign, slot)
	if err != nil {
		return nil, err
	}
	return o.GetTask(taskSign)
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
func (o DatabaseService) GetAllTasks() ([]*paradigm.Task, error) {
	var tasks []*paradigm.Task
	err := o.db.Order("start_time DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		tx := paradigm.DevReference{}
		if err := o.db.Take(&tx, task.TID).Error; err == nil {
			task.TxReceipt = &tx.TxReceipt
			task.TxBlockHash = tx.TxBlockHash
			task.TxHash = tx.TxHash
		}
	}
	return tasks, nil
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
		if task.Status == paradigm.Finished {
			if currentValue, exists := synthData[task.Model]; exists {
				synthData[task.Model] = currentValue + task.Process
			} else {
				synthData[task.Model] = task.Process
			}
		}
	}

	return synthData, nil
}

// DownUnFinishedTasks 未完成的任务设置为失败
func (o DatabaseService) DownUnFinishedTasks() error {
	result := o.db.Model(&paradigm.Task{}).
		Where("status = ?", paradigm.Processing).
		Updates(map[string]interface{}{
			"status": paradigm.Failed,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to reset processing tasks: %w", result.Error)
	}

	paradigm.Log("INFO", fmt.Sprintf("Reset %d processing tasks to failed status", result.RowsAffected))
	return nil
}

// RecoverCollector 恢复任务的Collector
func (o DatabaseService) RecoverCollector(task *paradigm.Task) error {
	if task.Status != paradigm.Finished {
		return fmt.Errorf("task is not finished, cannot be downloaded")
	}
	if task.Collector == nil {
		task.Collector = Collector.NewCollector(task.Sign, task.OutputType, o.channel)
	}
	var slots []*paradigm.Slot
	err := o.db.Where("task_id = ? AND status = ?", task.Sign, paradigm.Finished).Find(&slots).Error
	if err != nil {
		return fmt.Errorf("failed to query slots table: %w", err)
	}
	for _, slot := range slots {
		collectSlot := paradigm.CollectSlotItem{
			Sign:        task.Sign,
			Hash:        slot.SlotID,
			Size:        slot.ScheduleSize,
			PaddingSize: slot.CommitSlot.GetPadding(),
			StoreMethod: slot.CommitSlot.GetStore(),
		}
		task.Collector.ProcessSlotUpdate(collectSlot)
	}
	return nil
}
