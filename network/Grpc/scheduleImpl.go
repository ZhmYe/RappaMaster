package Grpc

import (
	"RappaMaster/helper"
	"RappaMaster/paradigm"
	pb "RappaMaster/pb/service"
	"RappaMaster/types"
	"context"
	"fmt"
	"google.golang.org/grpc"
)

func (ge *GrpcEngine) sendSchedule(slot types.ScheduleSlot) {
	if slot.NodeID >= len(ge.nodeAddress) {
		helper.GlobalServiceHelper.ReportError(paradigm.RaiseError(paradigm.RuntimeError, "Invalid Schedule, Node ID out of range", fmt.Errorf("%d >= %d", slot.NodeID, len(ge.nodeAddress))))
		return
	}
	request := pb.ScheduleRequest{
		Sign:  slot.Task,
		Size:  int32(slot.Size),
		Model: paradigm.ModelTypeToString(slot.Model),
	}
	conn, err := ge.GetConnection(slot.NodeID)
	if err != nil {
		helper.GlobalServiceHelper.ReportError(err)
		return
	}
	client := pb.NewRappaSynthesizerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), ge.ConnTimeout)
	defer cancel()

	resp, err := client.Schedule(ctx, &request, grpc.WaitForReady(true))
	if err != nil {
		helper.GlobalServiceHelper.ReportError(paradigm.RaiseError(paradigm.NetworkError, fmt.Sprintf("Failed to send schedule to node %d", slot.NodeID), err))
		return
	}
	// TODO 如果这里没写成功怎么办，是否应该是先写，再调度
	if resp.Accept {
		if err := helper.GlobalServiceHelper.DB.UpdateSlotFromSchedule(slot); err != nil {
			helper.GlobalServiceHelper.ReportError(err)
		}

	} else {
		helper.GlobalServiceHelper.ReportError(paradigm.RaiseError(paradigm.NetworkError, fmt.Sprintf("Fail to send schedule to node %d", slot.NodeID), fmt.Errorf("node reject to accept the schedule")))
	}
}
