package task

import (
	"RappaMaster/paradigm"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Task struct {
	taskID   int32
	sign     string
	name     string
	model    paradigm.SupportModelType
	expected int32
	finish   int32
	txHash   string
}

func (task *Task) SetTxHash(txHash string) {
	if task.txHash != "" {
		return // TODO We ignore it
	}
	task.txHash = txHash
}
func (task *Task) Name() string {
	return task.name
}
func (task *Task) Sign() string {
	return task.sign
}
func (task *Task) Model() string {
	return paradigm.ModelTypeToString(task.model)
}
func (task *Task) Expected() int32 {
	return task.expected
}
func NewTask(name string, model paradigm.SupportModelType, expected int32) *Task {
	hash := sha256.New()
	hash.Write([]byte(name))
	hash.Write([]byte(fmt.Sprintf("%d", expected)))
	hash.Write([]byte(paradigm.ModelTypeToString(model)))
	return &Task{
		taskID:   -1,
		sign:     hex.EncodeToString(hash.Sum([]byte{})),
		name:     name,
		model:    model,
		expected: expected,
		finish:   0,
		txHash:   "",
	}
}

func DefaultTaskForTest() *Task {
	return NewTask("Test Task", paradigm.CTGAN, 1000)
}
