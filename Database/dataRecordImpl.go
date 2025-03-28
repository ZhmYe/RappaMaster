package Database

import (
	"BHLayer2Node/Date"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"
)

// 这里尝试创建或获取dates
func (o DatabaseService) GetDateRecord(date time.Time) *Date.DateRecord {
	// 获取最新的datarecord
	var count int64
	o.db.Model(&Date.DateRecord{}).Count(&count)
	duration := int64(paradigm.GetDateDuration(date))
	var record *Date.DateRecord
	if count > duration {
		record = &Date.DateRecord{}
		o.db.Last(record)
	} else {
		for duration >= count {
			record = Date.NewDateRecord(paradigm.GetGenesisDate().Add(time.Duration(24*count) * time.Hour))
			o.db.Create(record)
			count++
		}
	}
	return record
}

// 更新dateRecord
func (o DatabaseService) UpdateDateRecord(record *Date.DateRecord) {
	o.db.Save(record)
}

func (o DatabaseService) GetDateRecords() ([]*Date.DateRecord, error) {
	var records []*Date.DateRecord
	err := o.db.Order("date asc").Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query date records: %v", err)
	}
	return records, nil
}
