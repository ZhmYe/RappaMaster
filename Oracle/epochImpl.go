package Oracle

import (
	"BHLayer2Node/paradigm"
	"fmt"

	"gorm.io/gorm"
)

// // 记录epoch
// func (o *PersistedOracle) setEpoch(epochRecord *paradigm.DevEpoch) {
// 	o.db.Create(epochRecord)
// }

func (o *PersistedOracle) saveEpochRecord(epoch *paradigm.DevEpoch) error {
	return o.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Create(epoch)
		if result.Error != nil {
			return result.Error
		}
		return nil
	})
}

func (o *PersistedOracle) setEpoch(epochRecord *paradigm.DevEpoch) error {
	if err := o.saveEpochRecord(epochRecord); err != nil {
		paradigm.Error(paradigm.RuntimeError,
			fmt.Sprintf("Failed to save epoch record: %v", err))
		return err
	}
	return nil
}
