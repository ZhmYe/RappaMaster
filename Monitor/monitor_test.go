package Monitor

import (
	"BHLayer2Node/paradigm"
	"testing"
)

func TestSelectLeastLoadedNodeWithReservationsBalancesBatch(t *testing.T) {
	monitor := newTestMonitor()
	reserved := map[int32]int{}
	for i := 0; i < 51; i++ {
		nodeID := monitor.SelectLeastLoadedNodeWithReservations(reserved)
		reserved[nodeID]++
	}

	if len(reserved) != 4 {
		t.Fatalf("expected all 4 nodes to receive reservations, got %#v", reserved)
	}
	for nodeID, count := range reserved {
		if count < 12 || count > 13 {
			t.Fatalf("node %d should receive 12 or 13 tasks, got %d; all=%#v", nodeID, count, reserved)
		}
	}
}

func TestAdviceReturnsOnlyIdleNodeForSingleSlot(t *testing.T) {
	monitor := newTestMonitor()
	counts := map[int32]int{}
	for i := 0; i < 4; i++ {
		request := paradigm.NewAdviceRequest(1, 1)
		monitor.advice(request)
		resp := request.ReceiveResponse()
		if len(resp.NodeIDs) != 1 || len(resp.ScheduleSize) != 1 {
			t.Fatalf("single-slot advice should return exactly one node, got %#v", resp)
		}
		counts[resp.NodeIDs[0]]++
	}

	for nodeID := int32(0); nodeID < 4; nodeID++ {
		if counts[nodeID] != 1 {
			t.Fatalf("node %d should receive exactly one single-slot reservation, all=%#v", nodeID, counts)
		}
	}

	request := paradigm.NewAdviceRequest(1, 1)
	monitor.advice(request)
	resp := request.ReceiveResponse()
	if len(resp.NodeIDs) != 0 || len(resp.ScheduleSize) != 0 {
		t.Fatalf("expected no advice when all nodes already have reserved work, got %#v", resp)
	}
}

func newTestMonitor() *Monitor {
	channel := &paradigm.RappaChannel{
		Config: &paradigm.BHLayer2NodeConfig{
			BHNodeAddressMap: map[int]*paradigm.BHNodeAddress{
				0: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9000},
				1: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9001},
				2: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9002},
				3: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9003},
			},
		},
	}
	return NewMonitor(channel)
}
