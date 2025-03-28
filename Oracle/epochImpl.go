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

// GetEpochByID 通过纪元标识查询纪元
func (o *PersistedOracle) GetEpochByID(epochID int32) (*paradigm.DevEpoch, error) {
	var epoch paradigm.DevEpoch
	err := o.db.Where("epoch_id = ?", epochID).First(&epoch).Error
	if err != nil {
		return nil, err
	}
	tx := paradigm.DevReference{}
	if err := o.db.Take(&tx, epoch.TID).Error; err != nil {
		return nil, fmt.Errorf("failed to get associated transaction: %v", err)
	}
	epoch.TxHash = tx.TxHash
	epoch.TxBlockHash = tx.TxBlockHash
	epoch.TxReceipt = &tx.TxReceipt

	return &epoch, nil
}

// GetEpochByTxHash 通过交易哈希查询纪元
func (o *PersistedOracle) GetEpochByTxHash(txHash string) (*paradigm.DevEpoch, error) {
	tx, err := o.GetTransactionByHash(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}
	if tx.Rf != paradigm.EpochTx {
		return nil, fmt.Errorf("transaction is not an epoch transaction")
	}
	if tx.EpochID == -1 {
		return nil, fmt.Errorf("invalid epoch ID in transaction")
	}
	epoch, err := o.GetEpochByID(tx.EpochID)
	if err != nil {
		return nil, fmt.Errorf("failed to get epoch: %v", err)
	}

	return epoch, nil
}

// GetLatestEpochs 查询 limit 条最新纪元
func (o *PersistedOracle) GetLatestEpochs(limit int) ([]*paradigm.DevEpoch, error) {
	var epochs []*paradigm.DevEpoch
	err := o.db.Order("epoch_id desc").Limit(limit).Find(&epochs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query latest epochs: %v", err)
	}
	for _, epoch := range epochs {
		tx := paradigm.DevReference{}
		if err := o.db.Take(&tx, epoch.TID).Error; err != nil {
			// 记录错误但继续处理其他 epoch
			paradigm.Log("ERROR", fmt.Sprintf("Failed to get transaction for epoch %d: %v", epoch.EpochID, err))
			continue
		}
		epoch.TxHash = tx.TxHash
		epoch.TxBlockHash = tx.TxBlockHash
		epoch.TxReceipt = &tx.TxReceipt
	}

	return epochs, nil
}
