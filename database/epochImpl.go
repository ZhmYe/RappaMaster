package database

import (
	"RappaMaster/config"
	"RappaMaster/paradigm"
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
		return -1, paradigm.RaiseError(paradigm.DatabaseError, "invalid parse result", errors.New("data[current_epoch] is not int64"))
	} else {
		return currentEpoch, nil
	}
}

func (dbs *DatabaseService) AdvanceEpoch() error {
	_, err := dbs.script(path.Join(config.ProjectRootPath, "database/sql/advance_new_epoch.sql"), false)
	return err
}

// UpdateEpochRoot updates the epoch root hash
func (dbs *DatabaseService) UpdateEpochRoot(epochID int, root string) error {
	query := "UPDATE epoch SET epochRoot = ? WHERE id = ?"
	_, err := dbs.db.Exec(query, root, epochID)
	return err
}

// GetEpochRoot retrieves the epoch root hash
func (dbs *DatabaseService) GetEpochRoot(epochID int) (string, error) {
	var root string
	query := "SELECT epochRoot FROM epoch WHERE id = ?"
	err := dbs.db.QueryRow(query, epochID).Scan(&root)
	return root, err
}

// GetCommittedSlotsInEpoch retrieves all committed slots in an epoch
func (dbs *DatabaseService) GetCommittedSlotsInEpoch(epochID int) (map[string]SlotInfo, error) {
	query := `
		SELECT s.slotHash, t.sign, s.nodeID, s.commitment, s.signature 
		FROM slot s 
		JOIN task t ON s.taskID = t.id 
		WHERE s.commitEpoch = ? AND s.status = 'committed'
	`
	
	rows, err := dbs.db.Query(query, epochID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make(map[string]SlotInfo)
	for rows.Next() {
		var slot SlotInfo
		err := rows.Scan(&slot.SlotHash, &slot.TaskSign, &slot.NodeID, &slot.Commitment, &slot.Signature)
		if err != nil {
			continue
		}
		slots[slot.SlotHash] = slot
	}

	return slots, nil
}

// SlotInfo represents slot information from database
type SlotInfo struct {
	SlotHash   string
	TaskSign   string
	NodeID     int
	Commitment string
	Signature  string
}

// UpdateSlotMerkleProof updates slot's merkle proof
func (dbs *DatabaseService) UpdateSlotMerkleProof(slotHash, proof string) error {
	query := "UPDATE slot SET merkleProof = ? WHERE slotHash = ?"
	_, err := dbs.db.Exec(query, proof, slotHash)
	return err
}

// UpdateTaskMerkleProofs updates task merkle proofs for multiple slots
func (dbs *DatabaseService) UpdateTaskMerkleProofs(slotHashes []string, proof string) error {
	if len(slotHashes) == 0 {
		return nil
	}
	
	// Build placeholders for IN clause
	placeholders := ""
	args := []interface{}{proof}
	for i, hash := range slotHashes {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, hash)
	}
	
	query := "UPDATE slot SET taskMerkleProof = ? WHERE slotHash IN (" + placeholders + ")"
	_, err := dbs.db.Exec(query, args...)
	return err
}

// GetSlotsFromNodes retrieves slots committed by specific nodes in an epoch
func (dbs *DatabaseService) GetSlotsFromNodes(epochID int, nodeIDs []int) ([]string, error) {
	if len(nodeIDs) == 0 {
		return nil, nil
	}
	
	// Build placeholders for IN clause
	placeholders := ""
	args := []interface{}{epochID}
	for i, nodeID := range nodeIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, nodeID)
	}
	
	query := "SELECT slotHash FROM slot WHERE commitEpoch = ? AND nodeID IN (" + placeholders + ") AND status = 'committed'"
	rows, err := dbs.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slotHashes []string
	for rows.Next() {
		var slotHash string
		if err := rows.Scan(&slotHash); err == nil {
			slotHashes = append(slotHashes, slotHash)
		}
	}

	return slotHashes, nil
}

// UpdateSlotStatus updates status of multiple slots
func (dbs *DatabaseService) UpdateSlotStatus(slotHashes []string, status string) error {
	if len(slotHashes) == 0 {
		return nil
	}
	
	// Build placeholders for IN clause
	placeholders := ""
	args := []interface{}{status}
	for i, hash := range slotHashes {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, hash)
	}
	
	query := "UPDATE slot SET status = ? WHERE slotHash IN (" + placeholders + ")"
	_, err := dbs.db.Exec(query, args...)
	return err
}

// GetJustifiedSlotsByTask retrieves justified slots grouped by task
func (dbs *DatabaseService) GetJustifiedSlotsByTask(taskSign string) ([]string, error) {
	query := `
		SELECT s.slotHash 
		FROM slot s 
		JOIN task t ON s.taskID = t.id 
		WHERE t.sign = ? AND s.status = 'justified'
		ORDER BY s.slotHash
	`
	
	rows, err := dbs.db.Query(query, taskSign)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slotHashes []string
	for rows.Next() {
		var slotHash string
		if err := rows.Scan(&slotHash); err == nil {
			slotHashes = append(slotHashes, slotHash)
		}
	}

	return slotHashes, nil
}

// UpdateTaskRoot updates task root hash and completion status
func (dbs *DatabaseService) UpdateTaskRoot(taskSign, taskRoot string, finishEpoch int) error {
	query := "UPDATE task SET taskRoot = ?, done = TRUE, finishEpoch = ?, finishDate = NOW() WHERE sign = ?"
	_, err := dbs.db.Exec(query, taskRoot, finishEpoch, taskSign)
	return err
}

// FinalizeTaskSlots updates all justified slots of a task to finalized status
func (dbs *DatabaseService) FinalizeTaskSlots(taskSign string, finalizeEpoch int) error {
	query := `
		UPDATE slot s 
		JOIN task t ON s.taskID = t.id 
		SET s.status = 'finalized', s.finalizeEpoch = ? 
		WHERE t.sign = ? AND s.status = 'justified'
	`
	_, err := dbs.db.Exec(query, finalizeEpoch, taskSign)
	return err
}

// CheckTaskCompletion checks if a task has enough justified slots to be completed
func (dbs *DatabaseService) CheckTaskCompletion(taskSign string) (bool, error) {
	query := `
		SELECT 
			t.expected,
			COALESCE(SUM(s.finished), 0) as totalFinished
		FROM task t 
		LEFT JOIN slot s ON t.id = s.taskID AND s.status = 'justified'
		WHERE t.sign = ?
		GROUP BY t.id, t.expected
	`
	
	var expected, totalFinished int64
	err := dbs.db.QueryRow(query, taskSign).Scan(&expected, &totalFinished)
	if err != nil {
		return false, err
	}
	
	return totalFinished >= expected, nil
}
