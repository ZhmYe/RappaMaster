package task

import (
	"RappaMaster/database"
	"RappaMaster/merkle"
	"RappaMaster/paradigm"
	"fmt"
)

// TaskCompletionManager handles task completion logic
type TaskCompletionManager struct {
	db *database.DatabaseService
}

// NewTaskCompletionManager creates a new task completion manager
func NewTaskCompletionManager(db *database.DatabaseService) *TaskCompletionManager {
	return &TaskCompletionManager{
		db: db,
	}
}

// CheckAndCompleteTask checks if a task can be completed and processes it
func (tcm *TaskCompletionManager) CheckAndCompleteTask(taskSign string, currentEpoch int) error {
	// Check if task has enough justified slots to be completed
	canComplete, err := tcm.db.CheckTaskCompletion(taskSign)
	if err != nil {
		return fmt.Errorf("failed to check task completion: %v", err)
	}

	if !canComplete {
		paradigm.Log("DEBUG", fmt.Sprintf("Task %s not ready for completion", taskSign))
		return nil
	}

	// Get all justified slots for this task
	justifiedSlots, err := tcm.db.GetJustifiedSlotsByTask(taskSign)
	if err != nil {
		return fmt.Errorf("failed to get justified slots: %v", err)
	}

	if len(justifiedSlots) == 0 {
		return fmt.Errorf("no justified slots found for task %s", taskSign)
	}

	// Calculate task root from justified slots
	taskRoot, err := tcm.calculateTaskRoot(justifiedSlots)
	if err != nil {
		return fmt.Errorf("failed to calculate task root: %v", err)
	}

	// Update task with root and completion status
	err = tcm.db.UpdateTaskRoot(taskSign, taskRoot, currentEpoch)
	if err != nil {
		return fmt.Errorf("failed to update task root: %v", err)
	}

	// Finalize all justified slots for this task
	err = tcm.db.FinalizeTaskSlots(taskSign, currentEpoch)
	if err != nil {
		return fmt.Errorf("failed to finalize task slots: %v", err)
	}

	paradigm.Log("INFO", fmt.Sprintf("Task %s completed with root %s, finalized %d slots", 
		taskSign, taskRoot, len(justifiedSlots)))

	return nil
}

// calculateTaskRoot calculates merkle root from justified slot hashes
func (tcm *TaskCompletionManager) calculateTaskRoot(slotHashes []string) (string, error) {
	if len(slotHashes) == 0 {
		return "", fmt.Errorf("no slots to calculate root from")
	}

	// Build merkle tree from justified slots
	tree := merkle.BuildSlotMerkleTree(slotHashes)
	if tree == nil {
		return "", fmt.Errorf("failed to build merkle tree")
	}

	return tree.GetRootHash(), nil
}

// ProcessCompletedTasks checks all active tasks for completion
func (tcm *TaskCompletionManager) ProcessCompletedTasks(currentEpoch int) error {
	// Get all active tasks that might be ready for completion
	activeTasks, err := tcm.getActiveTasks()
	if err != nil {
		return fmt.Errorf("failed to get active tasks: %v", err)
	}

	var completedCount int
	for _, taskSign := range activeTasks {
		err := tcm.CheckAndCompleteTask(taskSign, currentEpoch)
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to process task %s: %v", taskSign, err))
			continue
		}
		completedCount++
	}

	if completedCount > 0 {
		paradigm.Log("INFO", fmt.Sprintf("Processed %d completed tasks in epoch %d", completedCount, currentEpoch))
	}

	return nil
}

// getActiveTasks retrieves all active tasks that are not yet completed
func (tcm *TaskCompletionManager) getActiveTasks() ([]string, error) {
	query := "SELECT sign FROM task WHERE done = FALSE"
	rows, err := tcm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var taskSign string
		if err := rows.Scan(&taskSign); err == nil {
			tasks = append(tasks, taskSign)
		}
	}

	return tasks, nil
}

// GetTaskProgress returns the progress of a specific task
func (tcm *TaskCompletionManager) GetTaskProgress(taskSign string) (*TaskProgress, error) {
	query := `
		SELECT 
			t.sign,
			t.expected,
			t.done,
			t.taskRoot,
			COALESCE(SUM(CASE WHEN s.status = 'committed' THEN s.finished ELSE 0 END), 0) as committedProgress,
			COALESCE(SUM(CASE WHEN s.status = 'justified' THEN s.finished ELSE 0 END), 0) as justifiedProgress,
			COALESCE(SUM(CASE WHEN s.status = 'finalized' THEN s.finished ELSE 0 END), 0) as finalizedProgress,
			COUNT(CASE WHEN s.status = 'committed' THEN 1 END) as committedSlots,
			COUNT(CASE WHEN s.status = 'justified' THEN 1 END) as justifiedSlots,
			COUNT(CASE WHEN s.status = 'finalized' THEN 1 END) as finalizedSlots
		FROM task t 
		LEFT JOIN slot s ON t.id = s.taskID 
		WHERE t.sign = ?
		GROUP BY t.id
	`

	var progress TaskProgress
	err := tcm.db.QueryRow(query, taskSign).Scan(
		&progress.TaskSign,
		&progress.Expected,
		&progress.Done,
		&progress.TaskRoot,
		&progress.CommittedProgress,
		&progress.JustifiedProgress,
		&progress.FinalizedProgress,
		&progress.CommittedSlots,
		&progress.JustifiedSlots,
		&progress.FinalizedSlots,
	)

	return &progress, err
}

// TaskProgress represents the progress status of a task
type TaskProgress struct {
	TaskSign           string  `json:"taskSign"`
	Expected           int64   `json:"expected"`
	Done               bool    `json:"done"`
	TaskRoot           *string `json:"taskRoot"`
	CommittedProgress  int64   `json:"committedProgress"`
	JustifiedProgress  int64   `json:"justifiedProgress"`
	FinalizedProgress  int64   `json:"finalizedProgress"`
	CommittedSlots     int     `json:"committedSlots"`
	JustifiedSlots     int     `json:"justifiedSlots"`
	FinalizedSlots     int     `json:"finalizedSlots"`
}

// IsReadyForCompletion checks if task has enough justified progress
func (tp *TaskProgress) IsReadyForCompletion() bool {
	return !tp.Done && tp.JustifiedProgress >= tp.Expected
}

// GetCompletionPercentage returns the completion percentage based on justified progress
func (tp *TaskProgress) GetCompletionPercentage() float64 {
	if tp.Expected == 0 {
		return 0
	}
	return float64(tp.JustifiedProgress) / float64(tp.Expected) * 100
}