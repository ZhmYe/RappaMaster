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
}

func NewRappaChannel(config *Config.BHLayer2NodeConfig) *RappaChannel {
	initTasks := make(chan UnprocessedTask, config.MaxUnprocessedTaskPoolSize)
	unprocessedTasks := make(chan UnprocessedTask, config.MaxUnprocessedTaskPoolSize)
	//pendingRequestPool := make(chan paradigm.UnprocessedTask, config.MaxHttpRequestPoolSize)
	pendingSchedule := make(chan TaskSchedule, config.MaxPendingSchedulePoolSize)
	scheduledTasks := make(chan TaskSchedule, config.MaxScheduledTasksPoolSize)
	commitSlots := make(chan CommitSlotItem, config.MaxCommitSlotItemPoolSize)
	epochHeartbeat := make(chan *pb.HeartbeatRequest, 1)
	//slotToVotes := make(chan paradigm.CommitSlotItem, config.MaxCommitSlotItemPoolSize)
	pendingTransactions := make(chan Transaction, config.MaxCommitSlotItemPoolSize) // todo
	epochEvent := make(chan bool, 1)
	devTransactionChannel := make(chan []*PackedTransaction, config.MaxCommitSlotItemPoolSize) // todo
	toCollectSlotChanel := make(chan CommitSlotItem, config.MaxCommitSlotItemPoolSize)         // todo
	return &RappaChannel{
		InitTasks:        initTasks,
		UnprocessedTasks: unprocessedTasks,
		//PendingRequestPool:    pendingSchedule,
		PendingSchedule:        pendingSchedule,
		ScheduledTasks:         scheduledTasks,
		CommitSlots:            commitSlots,
		EpochHeartbeat:         epochHeartbeat,
		PendingTransactions:    pendingTransactions,
		EpochEvent:             epochEvent,
		DevTransactionChannel:  devTransactionChannel,
		ToCollectorSlotChannel: toCollectSlotChanel,
	}
}
