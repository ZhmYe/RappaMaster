package test

import (
	"BHLayer2Node/Oracle"
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
	"testing"
)

func TestOracle(t *testing.T) {
	oracle := Oracle.NewPersistedOracle()
	oracle.UpdateSlotFromSchedule(&paradigm.Slot{
		SlotID:       "fakeSign",
		TaskID:       "fakeSign",
		ScheduleID:   32,
		ScheduleSize: 12,
		Status:       paradigm.Failed,
		Err:          "error template",
		CommitSlot: &paradigm.CommitSlotItem{
			JustifiedSlot: &service.JustifiedSlot{
				Nid:        2,
				Process:    32,
				Sign:       "fakeSign",
				Slot:       12,
				Epoch:      2121,
				Commitment: nil,
				Padding:    nil,
				Store:      0,
			},
			InvalidType: 0,
		},
		Epoch: 0,
	})
}
