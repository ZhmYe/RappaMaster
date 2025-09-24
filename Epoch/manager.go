package epoch

import (
	"RappaMaster/crypto"
	"RappaMaster/database"
	"RappaMaster/merkle"
	"RappaMaster/paradigm"
	"encoding/json"
	"fmt"
	"sort"
)

// EpochProcessor handles epoch-level merkle tree operations
type EpochProcessor struct {
	db             *database.DatabaseService
	sigManager     *crypto.SignatureManager
	currentEpochID int
}

// NewEpochProcessor creates a new epoch processor
func NewEpochProcessor(db *database.DatabaseService) *EpochProcessor {
	return &EpochProcessor{
		db:         db,
		sigManager: crypto.NewSignatureManager(),
	}
}

// SlotsByTask groups slots by task for merkle tree calculation
type SlotsByTask map[string][]string // taskSign -> []slotHash

// SlotsByNode groups slots by node for merkle tree calculation
type SlotsByNode map[int][]string // nodeID -> []slotHash

// TaskMerkleInfo contains task merkle tree information
type TaskMerkleInfo struct {
	TaskSign string `json:"taskSign"`
	TaskRoot string `json:"taskRoot"`
	Slots    []string `json:"slots"`
}

// ProcessEpochMerkleTree processes merkle trees for an epoch
func (ep *EpochProcessor) ProcessEpochMerkleTree(epochID int) error {
	// Get all committed slots in this epoch
	committedSlots, err := ep.getCommittedSlotsInEpoch(epochID)
	if err != nil {
		return fmt.Errorf("failed to get committed slots: %v", err)
	}

	if len(committedSlots) == 0 {
		paradigm.Log("INFO", fmt.Sprintf("No committed slots found in epoch %d", epochID))
		return nil
	}

	// Group slots by task
	slotsByTask := ep.groupSlotsByTask(committedSlots)
	
	// Calculate merkle root for each task
	var taskMerkleInfos []TaskMerkleInfo
	for taskSign, slotHashes := range slotsByTask {
		taskTree := merkle.BuildSlotMerkleTree(slotHashes)
		if taskTree == nil {
			continue
		}
		
		taskRoot := taskTree.GetRootHash()
		taskMerkleInfos = append(taskMerkleInfos, TaskMerkleInfo{
			TaskSign: taskSign,
			TaskRoot: taskRoot,
			Slots:    slotHashes,
		})
		
		// Generate and store merkle proofs for each slot in this task
		err := ep.generateAndStoreSlotProofs(taskTree, slotHashes, taskSign, epochID)
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to generate slot proofs for task %s: %v", taskSign, err))
		}
	}

	// Calculate epoch merkle root from task roots
	epochRoot, err := ep.calculateEpochRoot(taskMerkleInfos)
	if err != nil {
		return fmt.Errorf("failed to calculate epoch root: %v", err)
	}

	// Generate and store task merkle proofs
	err = ep.generateAndStoreTaskProofs(taskMerkleInfos, epochRoot, epochID)
	if err != nil {
		paradigm.Log("ERROR", fmt.Sprintf("Failed to generate task proofs: %v", err))
	}

	// Update epoch with root hash
	err = ep.updateEpochRoot(epochID, epochRoot)
	if err != nil {
		return fmt.Errorf("failed to update epoch root: %v", err)
	}

	paradigm.Log("INFO", fmt.Sprintf("Processed epoch %d merkle tree, root: %s", epochID, epochRoot))
	return nil
}

// getCommittedSlotsInEpoch retrieves all committed slots in an epoch
func (ep *EpochProcessor) getCommittedSlotsInEpoch(epochID int) (map[string]SlotInfo, error) {
	return ep.db.GetCommittedSlotsInEpoch(epochID)
}

// SlotInfo represents slot information from database
type SlotInfo struct {
	SlotHash   string
	TaskSign   string
	NodeID     int
	Commitment string
	Signature  string
}

// groupSlotsByTask groups slots by their task
func (ep *EpochProcessor) groupSlotsByTask(slots map[string]SlotInfo) SlotsByTask {
	result := make(SlotsByTask)
	
	for slotHash, slot := range slots {
		if _, exists := result[slot.TaskSign]; !exists {
			result[slot.TaskSign] = make([]string, 0)
		}
		result[slot.TaskSign] = append(result[slot.TaskSign], slotHash)
	}
	
	// Sort slots within each task for deterministic ordering
	for taskSign := range result {
		sort.Strings(result[taskSign])
	}
	
	return result
}

// calculateEpochRoot calculates the epoch merkle root from task roots
func (ep *EpochProcessor) calculateEpochRoot(taskInfos []TaskMerkleInfo) (string, error) {
	if len(taskInfos) == 0 {
		return "", fmt.Errorf("no tasks to process")
	}

	// Sort tasks by taskSign for deterministic ordering
	sort.Slice(taskInfos, func(i, j int) bool {
		return taskInfos[i].TaskSign < taskInfos[j].TaskSign
	})

	// Extract task data for merkle tree
	var taskData []merkle.TaskData
	for _, info := range taskInfos {
		taskData = append(taskData, merkle.TaskData{
			TaskSign: info.TaskSign,
			TaskRoot: info.TaskRoot,
		})
	}

	// Build epoch merkle tree
	epochTree := merkle.BuildTaskMerkleTree(taskData)
	if epochTree == nil {
		return "", fmt.Errorf("failed to build epoch merkle tree")
	}

	return epochTree.GetRootHash(), nil
}

// generateAndStoreSlotProofs generates merkle proofs for slots within a task
func (ep *EpochProcessor) generateAndStoreSlotProofs(taskTree *merkle.MerkleTree, slotHashes []string, taskSign string, epochID int) error {
	for _, slotHash := range slotHashes {
		proof, err := taskTree.GenerateProof(slotHash)
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to generate proof for slot %s: %v", slotHash, err))
			continue
		}

		// Convert proof to JSON
		proofJSON, err := json.Marshal(proof)
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to marshal proof for slot %s: %v", slotHash, err))
			continue
		}

		// Store proof in database
		err = ep.updateSlotMerkleProof(slotHash, string(proofJSON))
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to store proof for slot %s: %v", slotHash, err))
		}
	}

	return nil
}

// generateAndStoreTaskProofs generates merkle proofs for tasks within epoch
func (ep *EpochProcessor) generateAndStoreTaskProofs(taskInfos []TaskMerkleInfo, epochRoot string, epochID int) error {
	// Build task merkle tree
	var taskData []merkle.TaskData
	for _, info := range taskInfos {
		taskData = append(taskData, merkle.TaskData{
			TaskSign: info.TaskSign,
			TaskRoot: info.TaskRoot,
		})
	}

	epochTree := merkle.BuildTaskMerkleTree(taskData)
	if epochTree == nil {
		return fmt.Errorf("failed to build epoch tree for proofs")
	}

	// Generate proof for each task
	for _, info := range taskInfos {
		// Find the leaf hash for this task
		leafData := fmt.Sprintf("%s:%s", info.TaskSign, info.TaskRoot)
		leafHash := merkle.Hash(leafData)
		
		proof, err := epochTree.GenerateProof(leafHash)
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to generate task proof for %s: %v", info.TaskSign, err))
			continue
		}

		// Convert proof to JSON
		proofJSON, err := json.Marshal(proof)
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to marshal task proof for %s: %v", info.TaskSign, err))
			continue
		}

		// Store task merkle proof for all slots in this task
		err = ep.updateTaskMerkleProofs(info.Slots, string(proofJSON))
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to store task proofs for %s: %v", info.TaskSign, err))
		}
	}

	return nil
}

// updateSlotMerkleProof updates slot's merkle proof in database
func (ep *EpochProcessor) updateSlotMerkleProof(slotHash, proof string) error {
	// TODO: Implement database update
	// UPDATE slot SET merkleProof = ? WHERE slotHash = ?
	return nil
}

// updateTaskMerkleProofs updates task merkle proofs for multiple slots
func (ep *EpochProcessor) updateTaskMerkleProofs(slotHashes []string, proof string) error {
	// TODO: Implement database update
	// UPDATE slot SET taskMerkleProof = ? WHERE slotHash IN (?)
	return nil
}

// updateEpochRoot updates epoch root hash in database
func (ep *EpochProcessor) updateEpochRoot(epochID int, root string) error {
	// TODO: Implement database update
	// UPDATE epoch SET epochRoot = ? WHERE id = ?
	return nil
}

// JustifySlots moves committed slots to justified state after BLS verification
func (ep *EpochProcessor) JustifySlots(epochID int, nodeSignatures map[int]string) ([]string, error) {
	// Get epoch root
	epochRoot, err := ep.getEpochRoot(epochID)
	if err != nil {
		return nil, fmt.Errorf("failed to get epoch root: %v", err)
	}

	// Validate BLS signatures
	validNodes, err := ep.sigManager.ValidateEpochSignatures(epochID, epochRoot, nodeSignatures)
	if err != nil {
		return nil, fmt.Errorf("failed to validate signatures: %v", err)
	}

	// Get slots from valid nodes
	justifiedSlots, err := ep.getSlotsFromNodes(epochID, validNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get slots from valid nodes: %v", err)
	}

	// Update slot status to justified
	err = ep.updateSlotStatus(justifiedSlots, "justified")
	if err != nil {
		return nil, fmt.Errorf("failed to update slot status: %v", err)
	}

	paradigm.Log("INFO", fmt.Sprintf("Justified %d slots in epoch %d from %d valid nodes", 
		len(justifiedSlots), epochID, len(validNodes)))

	return justifiedSlots, nil
}

// getEpochRoot retrieves epoch root from database
func (ep *EpochProcessor) getEpochRoot(epochID int) (string, error) {
	// TODO: Implement database query
	// SELECT epochRoot FROM epoch WHERE id = ?
	return "", nil
}

// getSlotsFromNodes retrieves slots committed by specific nodes in an epoch
func (ep *EpochProcessor) getSlotsFromNodes(epochID int, nodeIDs []int) ([]string, error) {
	// TODO: Implement database query
	// SELECT slotHash FROM slot WHERE commitEpoch = ? AND nodeID IN (?) AND status = 'committed'
	return nil, nil
}

// updateSlotStatus updates status of multiple slots
func (ep *EpochProcessor) updateSlotStatus(slotHashes []string, status string) error {
	// TODO: Implement database update
	// UPDATE slot SET status = ? WHERE slotHash IN (?)
	return nil
}