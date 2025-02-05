package paradigm

import (
	"BHLayer2Node/Config"
	pb "BHLayer2Node/pb/service"
)

type RappaChannel struct {
	Config           *Config.BHLayer2NodeConfig
	InitTasks        chan UnprocessedTask
	UnprocessedTasks chan UnprocessedTask
	//PendingRequestPool chan UnprocessedTask
	PendingSchedule        chan TaskSchedule
	ScheduledTasks         chan TaskSchedule
	CommitSlots            chan CommitSlotItem
	EpochHeartbeat         chan *pb.HeartbeatRequest
	PendingTransactions    chan Transaction
	EpochEvent             chan bool
	DevTransactionChannel  chan []*PackedTransaction
	ToCollectorSlotChannel chan CollectSlotItem

	ToCollectorRequestChannel chan CollectRequest
	SlotCollectChannel        chan RecoverConnection

	QueryChannel    chan QueryRequest
	ResponseChannel chan QueryResponse
	// ============================== DEBUG用的Channel==========================
	FakeCollectSignChannel chan [2]interface{} // 传递sign和size
	//SlotRecoverChannel     chan RecoverResponse
}

func NewRappaChannel(config *Config.BHLayer2NodeConfig) *RappaChannel {
	return &RappaChannel{
		Config:           config,
		InitTasks:        make(chan UnprocessedTask, config.MaxUnprocessedTaskPoolSize),
		UnprocessedTasks: make(chan UnprocessedTask, config.MaxUnprocessedTaskPoolSize),
		//PendingRequestPool:    pendingSchedule,
		PendingSchedule:           make(chan TaskSchedule, config.MaxPendingSchedulePoolSize),
		ScheduledTasks:            make(chan TaskSchedule, config.MaxScheduledTasksPoolSize),
		CommitSlots:               make(chan CommitSlotItem, config.MaxCommitSlotItemPoolSize),
		EpochHeartbeat:            make(chan *pb.HeartbeatRequest, 1),
		PendingTransactions:       make(chan Transaction, config.MaxCommitSlotItemPoolSize), // todo,
		EpochEvent:                make(chan bool, 1),
		DevTransactionChannel:     make(chan []*PackedTransaction, config.MaxCommitSlotItemPoolSize), // todo
		ToCollectorSlotChannel:    make(chan CollectSlotItem, config.MaxCommitSlotItemPoolSize),      // todo
		ToCollectorRequestChannel: make(chan CollectRequest, config.MaxCommitSlotItemPoolSize),       // todo
		SlotCollectChannel:        make(chan RecoverConnection, config.MaxCommitSlotItemPoolSize),    // todo
		FakeCollectSignChannel:    make(chan [2]interface{}, config.MaxCommitSlotItemPoolSize),       // todo
		//SlotRecoverChannel:     slotRecoverChannel,
		QueryChannel:    make(chan QueryRequest, config.QueueBufferSize),
		ResponseChannel: make(chan QueryResponse, config.QueueBufferSize),
	}
}
