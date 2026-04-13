package Database

import (
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// SetPlatformTask 创建或更新大任务
func (o DatabaseService) SetPlatformTask(task *paradigm.PlatformTask) error {
	return o.db.Save(task).Error
}

// GetAllPlatformTasks 获取所有大任务, 按创建时间倒序
func (o DatabaseService) GetAllPlatformTasks() ([]*paradigm.PlatformTask, error) {
	var tasks []*paradigm.PlatformTask
	err := o.db.Preload("SubTasks").Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetPlatformTaskByID 通过ID获取大任务
func (o DatabaseService) GetPlatformTaskByID(id string) (*paradigm.PlatformTask, error) {
	var task paradigm.PlatformTask
	err := o.db.Preload("SubTasks").Where("id = ?", id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// UpdatePlatformTask 更新大任务
func (o DatabaseService) UpdatePlatformTask(task *paradigm.PlatformTask) error {
	return o.db.Model(task).Updates(task).Error
}

// GetNextPlatformTaskID 获取下一个可用的 TSK ID
func (o DatabaseService) GetNextPlatformTaskID() (string, error) {
	var count int64
	err := o.db.Model(&paradigm.PlatformTask{}).Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TSK-%04d", 1001+count), nil
}

// GetFinishedTasks 获取所有已完成的任务，用于分析股票列表
func (o DatabaseService) GetFinishedTasks() ([]*paradigm.Task, error) {
	var tasks []*paradigm.Task
	// 查询已完成的任务，并且预加载所属的平台任务以获取平台任务名称
	err := o.db.Where("status = ?", paradigm.Finished).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
