package Tracker

import (
	"RappaMaster/helper"
	"RappaMaster/paradigm"
	"RappaMaster/types"
	"context"
	"fmt"
	"time"
)

const (
	TASK_EXPIRE_TIME = 2 * time.Minute
)

// Tracker retrieves the latest tasks from httpEngine and maintains their progress;
// When a slot completes the commit process, the tracker will receive this request to update the progress of a task;
// At the same time, the tracker interacts with Redis and uses Redis to complete the delay queue. When a task scheduling expires, if the task has not been completed yet, it will be rescheduled
type Tracker struct{}

func (tracker *Tracker) processNewTask(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-helper.GlobalServiceHelper.UnprocessedTasks:
			fmt.Println("receive new task")
			if t.NotBeenScheduled() {
				helper.GlobalServiceHelper.ScheduleQueue <- t.Sign()
			}
			err := helper.GlobalServiceHelper.Redis.Set(ctx, t.Sign(), "", TASK_EXPIRE_TIME).Err()
			if err != nil {
				helper.GlobalServiceHelper.ReportError(paradigm.RaiseError(paradigm.RedisError, "Set Task Expire Time failed", err))
				return
			}
		}
	}
}

func (tracker *Tracker) listenExpiration(ctx context.Context) {
	// config set notify-keyspace-events Ex
	pubsub := helper.GlobalServiceHelper.Redis.PSubscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}
			// when tracker find that a task is expired, it first check the (finish >= expected), if yes, then this task is finish, no need to re-schedule
			// @YZM: We have changed it to pass it to the Scheduler no matter what, so that the Scheduler can determine whether it has been completed

			//tsk := new(task.Task)
			//isFinish, err := helper.GlobalServiceHelper.DB.CheckTaskIsFinish(sign, tsk)
			//if err != nil {
			//	helper.GlobalServiceHelper.ReportError(err)
			//	return
			//}
			//if !isFinish {
			//	// re-schedule
			//	fmt.Println(fmt.Sprintf("task is not finished, remain: %d, re-schedule", tsk.Remain()))
			helper.GlobalServiceHelper.ScheduleTask(msg.Payload)
			helper.GlobalServiceHelper.UnprocessedTasks <- *types.SimpleTaskFromSign(msg.Payload)
			//}
		case <-ctx.Done():
			return
		}
	}
}

func StartAll(ctx context.Context) {
	tracker := new(Tracker)
	go tracker.listenExpiration(ctx)
	tracker.processNewTask(ctx)
}
