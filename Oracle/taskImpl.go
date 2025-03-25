package Oracle

import (
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// 获取任务
func (o *PersistedOracle) getTask(taskHash paradigm.TaskHash) (*paradigm.Task, error) {
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
func (o *PersistedOracle) updateScheduleInTask(schedule *paradigm.SynthTaskSchedule) {
	task, err := o.getTask(schedule.TaskID)
	if err != nil {
		panic(fmt.Sprintf("task not found of %s", schedule.TaskID))
	}
	task.Schedules = append(task.Schedules, schedule)
	task.ScheduleMap[schedule.ScheduleID] = len(task.Schedules)
	o.db.Model(task).Select("schedules", "schedule_map").Updates(task)
}

// 创建任务
func (o *PersistedOracle) setTask(task *paradigm.Task) {
	o.db.Omit("end_time").Create(task)
}

// 更新任务
func (o *PersistedOracle) updateTask(task *paradigm.Task) {
	o.db.Model(task).Updates(task)
}
