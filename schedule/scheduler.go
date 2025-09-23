package schedule

import (
	"RappaMaster/helper"
	"RappaMaster/types"
	"context"
	"fmt"
	"math"
)

// The Scheduler is used to reschedule tasks that have not been completed yet (when timeout occurs).
type Scheduler struct {
	monitor *Monitor
}

func (s *Scheduler) schedule(task *types.Task) []types.ScheduleSlot {
	toSchedule := task.Remain()
	if toSchedule <= 0 {
		return []types.ScheduleSlot{}
	}
	nodes, total := s.monitor.TopNodes()
	if len(nodes) == 0 {
		helper.GlobalServiceHelper.ReportError(fmt.Errorf("no active nodes"))
		return []types.ScheduleSlot{}
	}
	res := make([]types.ScheduleSlot, 0)
	// compute weight(speed), nodes[i].speed / total
	remaining := toSchedule
	for i := range nodes {
		weight := float64(nodes[i].Speed()) / float64(total)
		alloc := int64(math.Ceil(float64(toSchedule) * weight))

		if alloc <= 0 {
			alloc = 0
			break
		}
		res = append(res, types.ScheduleSlot{
			NodeID: nodes[i].NodeID,
			Size:   alloc,
			Task:   task.Sign(),
		})
		remaining -= alloc

		if remaining <= 0 {
			break
		}
	}
	if remaining > 0 {
		res[0].Size += remaining
	}
	return res
}

func (s *Scheduler) processUnFinishedTask(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case sign := <-helper.GlobalServiceHelper.ScheduleQueue:
			tsk := new(types.Task)
			isFinish, err := helper.GlobalServiceHelper.DB.CheckTaskIsFinish(sign, tsk)
			if err != nil {
				helper.GlobalServiceHelper.ReportError(err)
				continue
			}
			if !isFinish {
				helper.GlobalServiceHelper.SendToSchedule(s.schedule(tsk)...)
			}

		}
	}
}
func NewScheduler() *Scheduler {
	return &Scheduler{
		monitor: new(Monitor), //todo
	}
}
func StartAll(ctx context.Context) {
	scheduler := NewScheduler()
	go scheduler.processUnFinishedTask(ctx)
	go scheduler.monitor.Start(ctx)
}
