package Schedule

import (
	"BHLayer2node/Config"
	"BHLayer2node/LogWriter"
	"BHLayer2node/Network/Grpc"
	"BHLayer2node/Network/HTTP"
	"BHLayer2node/Schedule/Task"
	"BHLayer2node/Tracker"
	"BHLayer2node/paradigm"
	"fmt"
	"sync"
)

// Scheduler 用于分配任务
type Scheduler struct {
	grpcEngine         Grpc.GrpcInterface
	pendingRequestPool chan paradigm.HttpTaskRequest
	pendingSlotPool    chan paradigm.PendingSlotItem
	pendingSlotRecord  chan paradigm.SlotRecord
	config             *Config.BHLayer2NodeConfig
	tracker            *Tracker.Tracker
	// 任务状态管理
	tasks map[string]*Task.Task // 按 Sign 记录未完成的任务
	mu    sync.Mutex            // 保护 unprocessedTasks 的读写
}

func (s *Scheduler) Start() error {
	// 检查 grpcEngine 是否初始化
	if s.grpcEngine == nil {
		return fmt.Errorf("gRPC engine is not initialized. Call SetGrpc() before Start()")
	}
	LogWriter.Log("INFO", "Scheduler started, waiting for tasks...")
	// 处理新的request，来自前端,由http加入channel中
	processPendingRequestPool := func() {
		for request := range s.pendingRequestPool {
			LogWriter.Log("SCHEDULE", fmt.Sprintf("Processing request: Sign=%s, TotalSize=%d", request.Sign, request.Size))
			// 这里的任务应该是新的，新建一个Task，然后放到自己的pendingSlotPool中用于向grpcEngine说明需要发送给其它节点
			// 处理任务
			// 首先判断这个任务是否是新的，如果不是新的，那么忽略，这是代码问题，报warning或者error，但不panic出来了
			if _, exist := s.tasks[request.Sign]; exist {
				LogWriter.Log("ERROR", fmt.Sprintf("Task %s has been added!!!", request.Sign))
				continue // 不做任何处理
			}
			task := Task.NewTask(request.Sign, request.Model, request.Params, request.Size)
			s.process(task)
			s.tasks[request.Sign] = task
		}
	}
	// 处理由grpc得到的slotRecord，这是统计得到的一个slot里的完成情况，用于更新task
	processPendingSlotRecord := func() {
		for record := range s.pendingSlotRecord {
			LogWriter.Log("SCHEDULE", fmt.Sprintf("Processing record: Sign=%s, Slot=%d", record.Sign, record.Slot))
			task := s.tasks[record.Sign]
			task.Update(record)
			if task.IsFinish() {
				// todo 要放到chainUpper上准备上链一些信息

			} else {
				// 如果没有完成，那么就处理这一task
				s.process(task)
			}
		}
	}

	go processPendingRequestPool()
	go processPendingSlotRecord()

	return nil
}

// process 分配Slot，并更新任务状态
func (s *Scheduler) process(task *Task.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取节点列表
	nIDs := s.grpcEngine.GetConnectIndex()
	if len(nIDs) == 0 {
		LogWriter.Log("ERROR", "No connected nodes available for scheduling")
		return
	}
	slotSize := s.tracker.Advice(nIDs, task.Remain())
	// 分配任务：当前 Slot 的数据量
	allocSize := slotSize / len(nIDs)
	remainder := slotSize % len(nIDs)
	slot, err := task.Next(slotSize)
	if err != nil {
		panic(err)
	}
	// 构建分配计划
	schedules := make([]paradigm.BHLayer2NodeSchedule, 0)
	for i, id := range nIDs {
		allocatedSize := allocSize
		if i == len(nIDs)-1 {
			allocatedSize += remainder // 将余数分配给最后一个节点
		}

		schedules = append(schedules, paradigm.BHLayer2NodeSchedule{
			//Sign:   slot.Sign,
			//Slot:   slot.Slot,
			Size: allocatedSize,
			NID:  id,
			//Model:  slot.Model,
			//Params: slot.Params,
		})
	}
	slot.UpdateSchedule(schedules)
	LogWriter.Log("SCHEDULE", fmt.Sprintf("Schedule New Slot for Task %s, Slot: %d, Slot Size: %d, Remain Size: %d", slot.Sign, slot.Slot, slot.Size, task.Remain()))
	s.pendingSlotPool <- slot
}

func (s *Scheduler) Setup(config *Config.BHLayer2NodeConfig) {
	s.config = config
	//s.pendingRequestPool = make(chan HTTP.HttpTaskRequest, config.MaxHttpRequestPoolSize)
	//s.pendingSlotPool = make(chan PendingSlotItem, config.MaxGrpcRequestPoolSize)
	//s.pendingSlotRecord = make(chan SlotRecord, config.MaxGrpcRequestPoolSize)
}
func (s *Scheduler) SetGrpc(grpc Grpc.GrpcInterface) {
	s.grpcEngine = grpc
}
func (s *Scheduler) SetTracker(tracker *Tracker.Tracker) {
	s.tracker = tracker
}

// NewScheduler 创建新的 Scheduler todo
func NewScheduler(pendingSlotPool chan paradigm.PendingSlotItem, pendingRequestPool chan HTTP.HttpTaskRequest, pendingSlotRecord chan paradigm.SlotRecord) *Scheduler {
	return &Scheduler{
		grpcEngine:         nil,
		pendingRequestPool: pendingRequestPool,
		pendingSlotPool:    pendingSlotPool,
		pendingSlotRecord:  pendingSlotRecord,
		config:             nil,
		tracker:            nil,
		tasks:              make(map[string]*Task.Task),
		mu:                 sync.Mutex{},
	}
}
