package helper

import (
	"RappaMaster/config"
	"RappaMaster/database"
	fisco_bcos_client "RappaMaster/fisco-bcos-client"
	"RappaMaster/redis"
	"RappaMaster/transaction"
	"RappaMaster/types"
	"fmt"
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
		upchainBuffer:    make(chan transaction.Transaction, 100),
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
	upchainBuffer    chan transaction.Transaction
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
