package Date

import (
	"time"
)

// DateRecord 每日的一些信息记录
// 设置一个起始日期

type DateRecord struct {
	date            time.Time // 日期
	SynthData       int32     // 合成数据总量
	NbInitTasks     int32     // 新建任务数量
	NbFinalizedSlot int32     // 完成slot数量
	//nbCommitSlot    int32     // 提交slot数量
	NbTransactions int32 // 上链交易数量
	NbFinishTasks  int32 // 完成任务数量
}

func (r *DateRecord) UpdateProcess(process int32) {
	r.SynthData += process
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

func NewDateRecord(date time.Time) *DateRecord {
	return &DateRecord{
		date:            date,
		SynthData:       0,
		NbInitTasks:     0,
		NbFinalizedSlot: 0,
		//nbCommitSlot:    0,
		NbTransactions: 0,
		NbFinishTasks:  0,
	}
}
