package database

import (
	"RappaMaster/config"
	"RappaMaster/paradigm"
	pb "RappaMaster/pb/service"
	"RappaMaster/types"
	"encoding/hex"
	"fmt"
	"path"
)

func (dbs *DatabaseService) UpdateSlotFromSchedule(slot types.ScheduleSlot) error {
	params := []interface{}{
		slot.SlotHash(),
		slot.Task,
		slot.NodeID,
		slot.Size,
	}
	_, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/create_slot.sql"), false, params...)
	return err
}
func (dbs *DatabaseService) CommitSlot(request *pb.SlotCommitRequest) error {
	params := []interface{}{
		request.Size,
		hex.EncodeToString(request.Commitment),
		hex.EncodeToString(request.Signature),
		request.SlotHash,
	}
	_, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/commit_slot.sql"), false, params...)
	return err
}

func (o DatabaseService) SetSlotError(slotHash paradigm.SlotHash, e paradigm.InvalidCommitType, epoch int32) {
	slotQuery := o.GetSlot(slotHash)
	slotQuery.Epoch = epoch
	slotQuery.Err = paradigm.InvalidCommitTypeToString(e)
	//slot.CommitSlot.SetEpoch(epoch)
	o.db.Model(slotQuery).Select("epoch", "err", "status").Updates(slotQuery)
}

func (o DatabaseService) SetSlotFinish(slotHash paradigm.SlotHash, commitSlotItem *paradigm.CommitSlotItem) {
	slotQuery := o.GetSlot(slotHash)
	// 更新slot状态，这里应该是指针
	slotQuery.CommitSlot = commitSlotItem
	slotQuery.Status = paradigm.Finished
	slotQuery.Epoch = commitSlotItem.Epoch
	o.db.Model(slotQuery).Select("status", "epoch", "commit_slot").Updates(slotQuery)
}

func (o DatabaseService) GetSlot(slotHash paradigm.SlotHash) *paradigm.Slot {
	slotQuery := paradigm.Slot{}
	o.db.Where(paradigm.Slot{SlotID: slotHash}).Attrs(paradigm.NewSlotWithSlotID(slotHash)).FirstOrCreate(&slotQuery)
	return &slotQuery
}

// GetFinalizedSlotsCount 获取已完成的slot数量
func (o DatabaseService) GetFinalizedSlotsCount() (int64, error) {
	var count int64

	err := o.db.Model(&paradigm.Slot{}).
		Where("status = ?", paradigm.Finished). // 使用 Finished 状态筛选
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count finalized slots: %v", err)
	}

	// 添加日志记录
	paradigm.Log("DEBUG", fmt.Sprintf("Found %d finalized slots", count))

	return count, nil
}

func (o DatabaseService) DownUnFinishedSlots(epoch int32) error {
	updateMap := make(map[string]interface{})
	updateMap["epoch"] = epoch
	updateMap["err"] = paradigm.InvalidCommitTypeToString(paradigm.DOWN_FAILED)
	updateMap["status"] = paradigm.Failed
	return o.db.Model(&paradigm.Slot{}).Where("status = ?", paradigm.Processing).Updates(updateMap).Error
}
