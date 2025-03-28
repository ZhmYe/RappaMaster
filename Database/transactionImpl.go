package Database

import (
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// 获取交易
func (o DatabaseService) GetTransaction(txHash string) (*paradigm.DevReference, error) {
	txQuery := paradigm.DevReference{}
	err := o.db.Where(paradigm.DevReference{TxHash: txHash}).Take(&txQuery).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else {
		return &txQuery, nil
	}
}

// 保存交易
func (o DatabaseService) SetTransaction(tx *paradigm.DevReference) {
	o.db.Create(tx)
}

// GetTransactionByHash 通过交易哈希查询交易
func (o DatabaseService) GetTransactionByHash(txHash string) (*paradigm.DevReference, error) {
	var tx paradigm.DevReference
	err := o.db.Where("tx_hash = ?", txHash).First(&tx).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}
	return &tx, nil
}

// GetLatestTransactions 查询 limit 条最新交易
func (o DatabaseService) GetLatestTransactions(limit int) ([]paradigm.DevReference, error) {
	var txs []paradigm.DevReference
	err := o.db.Order("upchain_time desc").Limit(limit).Find(&txs).Error
	return txs, err
}

// GetTransactionCount 查询交易总数
func (o DatabaseService) GetTransactionCount() (int64, error) {
	var count int64
	err := o.db.Model(&paradigm.DevReference{}).Count(&count).Error
	return count, err
}
