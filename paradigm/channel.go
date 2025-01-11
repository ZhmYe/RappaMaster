package paradigm

import (
	"BHLayer2Node/Config"
	pb "BHLayer2Node/pb/service"
)

type RappaChannel struct {
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
	ToCollectorSlotChannel chan CommitSlotItem

	ToCollectorRequestChannel chan CollectRequest
	SlotCollectChannel        chan RecoverRequest
	// ============================== DEBUG用的Channel==========================
	FakeCollectSignChannel chan [2]interface{} // 传递sign和size
	//SlotRecoverChannel     chan RecoverResponse
}

func NewRappaChannel(config *Config.BHLayer2NodeConfig) *RappaChannel {
	//initTasks := make(chan UnprocessedTask, config.MaxUnprocessedTaskPoolSize)
	//unprocessedTasks := make(chan UnprocessedTask, config.MaxUnprocessedTaskPoolSize)
	////pendingRequestPool := make(chan paradigm.UnprocessedTask, config.MaxHttpRequestPoolSize)
	//pendingSchedule := make(chan TaskSchedule, config.MaxPendingSchedulePoolSize)
	//scheduledTasks := make(chan TaskSchedule, config.MaxScheduledTasksPoolSize)
	//commitSlots := make(chan CommitSlotItem, config.MaxCommitSlotItemPoolSize)
	//epochHeartbeat := make(chan *pb.HeartbeatRequest, 1)
	////slotToVotes := make(chan paradigm.CommitSlotItem, config.MaxCommitSlotItemPoolSize)
	//pendingTransactions := make(chan Transaction, config.MaxCommitSlotItemPoolSize) // todo
	//epochEvent := make(chan bool, 1)
	//devTransactionChannel := make(chan []*PackedTransaction, config.MaxCommitSlotItemPoolSize) // todo
	//toCollectSlotChanel := make(chan CommitSlotItem, config.MaxCommitSlotItemPoolSize)         // todo
	//toCollectRequestChannel := make(chan CollectRequest, config.MaxCommitSlotItemPoolSize)     // todo
	//slotCollectChannel := make(chan RecoverRequest, config.MaxCommitSlotItemPoolSize)          // todo

	//slotRecoverChannel := make(chan RecoverResponse, config.MaxCommitSlotItemPoolSize) // todo
	return &RappaChannel{
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
		ToCollectorSlotChannel:    make(chan CommitSlotItem, config.MaxCommitSlotItemPoolSize),       // todo
		ToCollectorRequestChannel: make(chan CollectRequest, config.MaxCommitSlotItemPoolSize),       // todo
		SlotCollectChannel:        make(chan RecoverRequest, config.MaxCommitSlotItemPoolSize),       // todo
		FakeCollectSignChannel:    make(chan [2]interface{}, config.MaxCommitSlotItemPoolSize),       // todo
		//SlotRecoverChannel:     slotRecoverChannel,
	}
}
