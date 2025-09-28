package helper

import (
	"RappaMaster/config"
	"RappaMaster/database"
	fisco_bcos_client "RappaMaster/fisco-bcos-client"
	pb "RappaMaster/pb/service"
	"RappaMaster/redis"
	"RappaMaster/types"
	"fmt"
	"sync"
)

var GlobalServiceHelper RappaChannel

func init() {
	dbs := database.NewDatabaseService(config.GlobalSystemConfig.DBConfig)
	if err := dbs.Init(); err != nil {
		panic(err.Error())
	}
	client, err := fisco_bcos_client.NewRappaFBClient(config.GlobalSystemConfig.FBConfig)
	if err != nil {
		panic(err.Error())
	}
	redisService := redis.NewRedisService(config.GlobalSystemConfig.RedisConfig)
	if err := redisService.Init(); err != nil {
		panic(err.Error())
	}
	GlobalServiceHelper = RappaChannel{
		DB:               dbs,
		Chain:            client,
		Redis:            redisService,
		ErrorHandler:     make(chan error, 100),
		UnprocessedTasks: make(chan types.Task, 100),
		ScheduleQueue:    make(chan string, 100),
		upchainBuffer:    make(chan types.Transaction, 100),
	}
}

type RappaChannel struct {
	DB               *database.DatabaseService // shared db
	Chain            *fisco_bcos_client.RappaFBClient
	Redis            *redis.RedisService // redis
	UnprocessedTasks chan types.Task
	ScheduleQueue    chan string // sign
	SlotSchedule     chan types.ScheduleSlot
	ErrorHandler     chan error
	upchainBuffer    chan types.Transaction
	EpochUpdateQueue chan *pb.SlotCommitRequest
	EvidenceQueue    chan types.BlockedGrpcPayload[types.EpochIntegrityEvidence, error]
}

func (rc *RappaChannel) UpdateNewTaskTrack(t types.Task) {
	rc.UnprocessedTasks <- t
}

func (rc *RappaChannel) ReportError(err error) {
	rc.ErrorHandler <- err
}

func (rc *RappaChannel) ScheduleTask(sign string) {
	rc.ScheduleQueue <- sign
}

func (rc *RappaChannel) SendToSchedule(slots ...types.ScheduleSlot) {
	fmt.Printf("new slots to schedule, len = %d\n", len(slots))
	for _, slot := range slots {
		go func() {
			rc.SlotSchedule <- slot
		}()
	}
}

func (rc *RappaChannel) UpdateEpochTree(slot *pb.SlotCommitRequest) {
	go func() {
		rc.EpochUpdateQueue <- slot
	}()
}

// SendIntegrityEvidence 这里返回值是失败的nodeID
func (rc *RappaChannel) SendIntegrityEvidence(evidences []types.EpochIntegrityEvidence) []types.NodeID {
	var wg sync.WaitGroup
	wg.Add(len(evidences))
	res := make([]types.NodeID, 0)
	for _, evidence := range evidences {
		go func() {
			defer wg.Done()
			message, conn := types.NewBlockedGrpcPayload[types.EpochIntegrityEvidence, error](evidence)
			rc.EvidenceQueue <- message
			if err := <-conn; err != nil {
				res = append(res, evidence.NodeID())
			}
		}()
	}
	wg.Wait()
	return res

}
