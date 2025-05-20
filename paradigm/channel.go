package paradigm

import (
	pb "BHLayer2Node/pb/service"
)

type RappaChannel struct {
	Config           *BHLayer2NodeConfig
	InitTasks        chan *SynthTaskTrackItem
	UnprocessedTasks chan UnprocessedTask
	//PendingRequestPool chan UnprocessedTask
	PendingSchedules chan SynthTaskSchedule
	OracleSchedules  chan *SynthTaskSchedule
	//PendingSchedule        chan TaskSchedule
	ScheduledTasks              chan SynthTaskSchedule
	CommitSlots                 chan CommitSlotItem
	EpochHeartbeat              chan *pb.HeartbeatRequest
	PendingTransactions         chan Transaction
	EpochEvent                  chan bool
	DevTransactionChannel       chan []*PackedTransaction
	ToCollectorSlotChannel      chan CollectSlotItem
	BlockchainQueryChannel      chan Query // 传递给queryHandler的链上信息查询
	BlockchainInfoUpdateChannel chan bool  // TODO queryHandler定时获取最新的区块数量
	MonitorHeartbeatChannel     chan NodeHeartbeatReport
	MonitorAdviceChannel        chan *AdviceRequest // todo
	MonitorOracleChannel        chan interface{}    // todo
	MonitorQueryChannel         chan Query

	ToCollectorRequestChannel chan HttpCollectRequest
	SlotCollectChannel        chan RecoverConnection
	QueryChannel              chan Query
	EpochInitTaskChannel      chan *Task
	UnScheduledSlotChannel    chan *Slot //临时处理不能调度的slot
	// ============================== DEBUG用的Channel==========================
	FakeCollectSignChannel chan [2]interface{} // 传递sign和size
	//SlotRecoverChannel     chan RecoverResponse
}

func NewRappaChannel(config *BHLayer2NodeConfig) *RappaChannel {
	return &RappaChannel{
		Config:           config,
		InitTasks:        make(chan *SynthTaskTrackItem, config.MaxUnprocessedTaskPoolSize),
		UnprocessedTasks: make(chan UnprocessedTask, config.MaxUnprocessedTaskPoolSize),
		//PendingRequestPool:    pendingSchedule,
		PendingSchedules: make(chan SynthTaskSchedule, config.MaxPendingSchedulePoolSize),
		OracleSchedules:  make(chan *SynthTaskSchedule, config.MaxPendingSchedulePoolSize),
		//PendingSchedule:           make(chan TaskSchedule, config.MaxPendingSchedulePoolSize),
		ScheduledTasks:              make(chan SynthTaskSchedule, config.MaxScheduledTasksPoolSize),
		CommitSlots:                 make(chan CommitSlotItem, config.MaxCommitSlotItemPoolSize),
		EpochHeartbeat:              make(chan *pb.HeartbeatRequest, 1),
		PendingTransactions:         make(chan Transaction, config.MaxCommitSlotItemPoolSize), // todo,
		EpochEvent:                  make(chan bool, 1),
		DevTransactionChannel:       make(chan []*PackedTransaction, config.MaxCommitSlotItemPoolSize), // todo
		ToCollectorSlotChannel:      make(chan CollectSlotItem, config.MaxCommitSlotItemPoolSize),      // todo
		ToCollectorRequestChannel:   make(chan HttpCollectRequest, config.MaxCommitSlotItemPoolSize),   // todo
		BlockchainQueryChannel:      make(chan Query, config.MaxCommitSlotItemPoolSize),                // todo
		BlockchainInfoUpdateChannel: make(chan bool, 1),
		MonitorAdviceChannel:        make(chan *AdviceRequest, config.MaxCommitSlotItemPoolSize),      // todo
		MonitorHeartbeatChannel:     make(chan NodeHeartbeatReport, config.MaxCommitSlotItemPoolSize), // todo
		MonitorOracleChannel:        make(chan interface{}, config.MaxCommitSlotItemPoolSize),         // todo
		MonitorQueryChannel:         make(chan Query, config.MaxCommitSlotItemPoolSize),               // todo
		SlotCollectChannel:          make(chan RecoverConnection, config.MaxCommitSlotItemPoolSize),   // todo
		QueryChannel:                make(chan Query, config.MaxCommitSlotItemPoolSize),               // todo
		EpochInitTaskChannel:        make(chan *Task, config.MaxCommitSlotItemPoolSize),
		UnScheduledSlotChannel:      make(chan *Slot, config.MaxCommitSlotItemPoolSize),
		FakeCollectSignChannel:      make(chan [2]interface{}, config.MaxCommitSlotItemPoolSize), // todo
		//SlotRecoverChannel:     slotRecoverChannel,
	}
}
