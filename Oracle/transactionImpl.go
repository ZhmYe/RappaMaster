package Oracle

import (
	"BHLayer2Node/paradigm"
	"errors"
	"gorm.io/gorm"
)

// 获取交易
func (o *PersistedOracle) getTransaction(txHash string) (*paradigm.DevReference, error) {
	txQuery := paradigm.DevReference{}
	err := o.db.Where(paradigm.DevReference{TxHash: txHash}).Take(&txQuery).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else {
		return &txQuery, nil
	}
}

// 保存交易
func (o *PersistedOracle) setTransaction(tx *paradigm.DevReference) {
	o.db.Create(tx)
}
