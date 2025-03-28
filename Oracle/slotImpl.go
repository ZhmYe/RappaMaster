package Oracle

import (
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// 这里定义slot的数据库操作
func (o *PersistedOracle) UpdateSlotFromSchedule(slot *paradigm.Slot) {
	slotQuery := paradigm.Slot{}
	err := o.db.Where(paradigm.Slot{SlotID: slot.SlotID}).Take(&slotQuery).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		o.db.Create(slot)
	} else {
		if slot.Status == paradigm.Failed {
			// 更新除了提交commitslot和epoch以外的值
			o.db.Model(slot).Omit("commit_slot", "epoch").Updates(slot)
		} else {
			o.db.Model(slot).Select("task_id", "schedule_id", "schedule_size").Updates(slot)
		}
	}
}

func (o *PersistedOracle) setSlotError(slotHash paradigm.SlotHash, e paradigm.InvalidCommitType, epoch int32) {
	slotQuery := o.getSlot(slotHash)
	slotQuery.Epoch = epoch
	slotQuery.Err = paradigm.InvalidCommitTypeToString(e)
	//slot.CommitSlot.SetEpoch(epoch)
	o.db.Model(slotQuery).Select("epoch", "err", "status").Updates(slotQuery)
}

func (o *PersistedOracle) setSlotFinish(slotHash paradigm.SlotHash, commitSlotItem *paradigm.CommitSlotItem) {
	slotQuery := o.getSlot(slotHash)
	// 更新slot状态，这里应该是指针
	slotQuery.CommitSlot = commitSlotItem
	slotQuery.Status = paradigm.Finished
	slotQuery.Epoch = commitSlotItem.Epoch
	o.db.Model(slotQuery).Select("status", "epoch", "commit_slot").Updates(slotQuery)
}

func (o *PersistedOracle) getSlot(slotHash paradigm.SlotHash) *paradigm.Slot {
	slotQuery := paradigm.Slot{}
	o.db.Where(paradigm.Slot{SlotID: slotHash}).Attrs(paradigm.NewSlotWithSlotID(slotHash)).FirstOrCreate(&slotQuery)
	return &slotQuery
}

// GetFinalizedSlotsCount 获取已完成的slot数量
func (o *PersistedOracle) GetFinalizedSlotsCount() (int64, error) {
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
