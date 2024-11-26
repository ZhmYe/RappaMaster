package Schedule

import (
	"BHCoordinator/Config"
	"BHCoordinator/LogWriter"
	"BHCoordinator/Monitor"
	"BHCoordinator/Task"
	"BHCoordinator/paradigm"
	"fmt"
	"sync"
)

// Scheduler 用于分配任务
type Scheduler struct {
	//grpcEngine       Grpc.GrpcInterface
	unprocessedTasks chan paradigm.UnprocessedTask
	pendingSchedules chan paradigm.TaskSchedule // 等待coordinator发送的调度方案

	config  *Config.BHCoordinatorConfig
	manager *Task.TaskManager // 监控任务运行状态
	monitor *Monitor.Monitor  // 监控节点状态
	// 任务状态管理
	//tasks map[string]*Task.Task // 按 Sign 记录未完成的任务
	mu sync.Mutex // 保护 unprocessedTasks 的读写
}

func (s *Scheduler) Start() error {
	// 检查 grpcEngine 是否初始化
	//if s.grpcEngine == nil {
	//	return fmt.Errorf("gRPC engine is not initialized. Call SetGrpc() before Start()")
	//}
	LogWriter.Log("INFO", "Scheduler started, waiting for tasks...")
	// 处理待调度的task，这些task有可能是由前端http新发的，也可能是task在前一个slot未完成返工的
	processUnprocessedTasks := func() {
		for request := range s.unprocessedTasks {
			LogWriter.Log("SCHEDULE", fmt.Sprintf("Schedule Unprocessed Task: Sign=%s, Total Size=%d", request.Sign, request.Size))
			// 首先判断这个slot是不是合法的slot，如果是已经过期的slot直接可以拒绝
			// 这里没在里面加锁，因为slot一定是前进的，如果request.Slot比现在的slot那么一定是过期的
			// 而如果是刚好一样但在那一瞬间过期了，在taskManager里会判断 todo
			//if isValid, err := s.manager.CheckSlotIsValid(request.Sign, request.Slot); !isValid {
			//	LogWriter.Log("ERROR", err.Error())
			//	continue
			//}
			//if _, exist := s.tasks[request.Sign]; exist {
			//	LogWriter.Log("ERROR", fmt.Sprintf("Task %s has been added!!!", request.Sign))
			//	continue // 不做任何处理
			//}
			//task := Task.NewTask(request.Sign, request.Model, request.Params, request.Size)
			s.process(request)
			//s.tasks[request.Sign] = task
		}
	}

	go processUnprocessedTasks()
	//go processPendingSlotRecord()

	return nil
}

// process 分配Slot，并更新任务状态
func (s *Scheduler) process(slot paradigm.UnprocessedTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取节点列表
	nIDs := []int{1, 2}
	if len(nIDs) == 0 {
		LogWriter.Log("ERROR", "No connected nodes available for scheduling")
		return
	}
	allocSizes := s.monitor.Advice(nIDs, slot.Size)
	if len(allocSizes) != len(nIDs) {
		panic("Error in Monitor Advice(), len(nIDs) != len(allocSizes)")
	}
	// 构建分配计划
	schedules := make([]paradigm.ScheduleItem, 0)
	for i, id := range nIDs {
		allocatedSize := allocSizes[i]
		schedules = append(schedules, paradigm.ScheduleItem{
			Size: allocatedSize,
			NID:  id,
		})
	}
	taskSchedule := paradigm.TaskSchedule{
		Sign:      slot.Sign,
		Slot:      slot.Slot,
		Size:      slot.Size,
		Model:     slot.Model,
		Params:    slot.Params,
		Schedules: schedules,
	}
	//slot.UpdateSchedule(schedules)
	LogWriter.Log("SCHEDULE", fmt.Sprintf("Schedule New Slot for Task %s, Slot: %d, Slot Size: %d", slot.Sign, slot.Slot, slot.Size))
	s.pendingSchedules <- taskSchedule
}

func (s *Scheduler) Setup(config *Config.BHCoordinatorConfig) {
	s.config = config
	//s.pendingRequestPool = make(chan HTTP.UnprocessedTask, config.MaxHttpRequestPoolSize)
	//s.pendingSlotPool = make(chan PendingSlotItem, config.MaxGrpcRequestPoolSize)
	//s.pendingSlotRecord = make(chan SlotRecord, config.MaxGrpcRequestPoolSize)
}

//	func (s *Scheduler) SetGrpc(grpc Grpc.GrpcInterface) {
//		s.grpcEngine = grpc
//	}
func (s *Scheduler) SetTaskManager(manager *Task.TaskManager) {
	s.manager = manager
}

// NewScheduler 创建新的 Scheduler todo
func NewScheduler(unprocessedTasks chan paradigm.UnprocessedTask, pendingSchedule chan paradigm.TaskSchedule) *Scheduler {
	return &Scheduler{
		//grpcEngine:       nil,
		unprocessedTasks: unprocessedTasks,
		pendingSchedules: pendingSchedule,
		config:           nil,
		mu:               sync.Mutex{},
	}
}
