package database

import (
	"RappaMaster/config"
	pb "RappaMaster/pb/service"
	"RappaMaster/types"
	"encoding/hex"
	"path"
	"strings"
)

func (dbs *DatabaseService) UpdateSlotFromSchedule(slot types.ScheduleSlot) error {
	params := []interface{}{
		slot.SlotHash(),
		slot.Task,
		slot.NodeID,
		slot.Size,
	}
	_, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/create_slot.sql"), false, params...)
	return err
}
func (dbs *DatabaseService) CommitSlot(request *pb.SlotCommitRequest) error {
	params := []interface{}{
		request.Size,
		hex.EncodeToString(request.Commitment),
		hex.EncodeToString(request.Signature),
		request.Sign,
		request.SlotHash,
	}
	_, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/commit_slot.sql"), false, params...)
	return err
}

func (dbs *DatabaseService) JustifiedSlot(slotHashes []string) error {
	params := []interface{}{
		len(slotHashes),
		strings.Join(slotHashes, ","),
	}
	_, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/justified_slot.sql"), false, params...)
	return err
}
