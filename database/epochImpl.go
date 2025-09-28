package database

import (
	"RappaMaster/config"
	pb "RappaMaster/pb/service"
	"RappaMaster/types"
	"encoding/hex"
	"errors"
	"path"
)

func (dbs *DatabaseService) GetCurrentEpoch() (int64, error) {
	result, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/query_current_epoch.sql"), true)
	if err != nil {
		return -1, err
	}
	data := make(map[string]interface{})
	result.Scan(data)
	if currentEpoch, ok := data["current_epoch"].(int64); !ok {
		return -1, types.RaiseError(types.DatabaseError, "invalid parse result", errors.New("data[current_epoch] is not int64"))
	} else {
		return currentEpoch, nil
	}
}

func (dbs *DatabaseService) AdvanceEpoch() error {
	_, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/advance_new_epoch.sql"), false)
	return err
}

func (dbs *DatabaseService) InitEpochTree(epochTree *types.EpochTree) error {
	result, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/query_unjustified_slot.sql"), true)
	if err != nil {
		return err
	}
	var data []map[string]interface{}
	result.Scan(data)
	for _, row := range data {
		// row有三个字段taskID, nodeID和commitment
		if sign, o1 := row["task_sign"].(string); o1 {
			if nodeID, o2 := row["node_id"].(int64); o2 {
				if c, o3 := row["commitment"].(string); o3 {
					commitment, err := hex.DecodeString(c)
					if err != nil {
						return types.RaiseError(types.DatabaseError, "invalid parse result", err)
					}
					if slotHash, o4 := row["slotHash"].(string); o4 {
						if err = epochTree.Update(&pb.SlotCommitRequest{
							SlotHash:   slotHash,
							Sign:       sign,
							NodeID:     int32(nodeID),
							Size:       0,
							Commitment: commitment,
							Signature:  nil,
							Votes:      nil,
							Store:      0,
						}); err != nil {
							return types.RaiseError(types.DatabaseError, "update epochTree failed", err)
						}
					} else {
						return types.RaiseError(types.DatabaseError, "invalid parse result", errors.New("data[slotHash] is not string"))
					}

				} else {
					return types.RaiseError(types.DatabaseError, "invalid parse result", errors.New("data[commitment] is not string"))
				}
			} else {
				return types.RaiseError(types.DatabaseError, "invalid parse result", errors.New("data[node_id] is not int64"))
			}

		} else {
			return types.RaiseError(types.DatabaseError, "invalid parse result", errors.New("data[task_id] is not int64"))
		}
	}
	return nil
}
