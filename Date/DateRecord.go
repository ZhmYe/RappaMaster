package Date

import (
	"BHLayer2Node/paradigm"
	"time"
)

// DateRecord 每日的一些信息记录
// 设置一个起始日期

type DateRecord struct {
	date            time.Time                           // 日期
	SynthData       map[paradigm.SupportModelType]int32 // 合成数据总量,按模型种类区分
	NbInitTasks     int32                               // 新建任务数量
	NbFinalizedSlot int32                               // 完成slot数量
	//nbCommitSlot    int32     // 提交slot数量
	NbTransactions      int32            // 上链交易数量
	NbFinishTasks       int32            // 完成任务数量
	DatasetDistribution map[string]int32 // 每天不同数据集的合成数量（任务数）
}

func (r *DateRecord) UpdateProcess(process int32, modelType paradigm.SupportModelType) {
	if value, ok := r.SynthData[modelType]; ok {
		r.SynthData[modelType] = value + process
	} else {
		r.SynthData[modelType] = process
	}
}

func (r *DateRecord) UpdateInitTasks(n int32) {
	r.NbInitTasks += n
}
func (r *DateRecord) UpdateFinalized(n int32) {
	r.NbFinalizedSlot += n
}
func (r *DateRecord) UpdateFinishTasks(n int32) {
	r.NbFinishTasks += n
}
func (r *DateRecord) Date() time.Time {
	return r.date
}

//	func (r *DateRecord) UpdateCommit(n int32) {
//		r.nbCommitSlot += n
//	}
func (r *DateRecord) UpdateTransactions(n int32) {
	r.NbTransactions += n
}
func (r *DateRecord) UpdateDateset(dataset string) {
	if dataset == "" {
		return
	}
	if _, exist := r.DatasetDistribution[dataset]; !exist {
		r.DatasetDistribution[dataset] = 0
	}
	r.DatasetDistribution[dataset] += 1
}
func NewDateRecord(date time.Time) *DateRecord {
	return &DateRecord{
		date:            date,
		SynthData:       make(map[paradigm.SupportModelType]int32),
		NbInitTasks:     0,
		NbFinalizedSlot: 0,
		//nbCommitSlot:    0,
		NbTransactions:      0,
		NbFinishTasks:       0,
		DatasetDistribution: make(map[string]int32),
	}
}
