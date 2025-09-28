package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

type TaskID int32
type Task struct {
	taskID   TaskID
	sign     string
	name     string
	model    SupportModelType
	expected int64
	finish   int64
	txHash   string
	isCreate bool
}

func (task *Task) SetTxHash(txHash string) {
	if task.txHash != "" {
		return // TODO We ignore it
	}
	task.txHash = txHash
}
func (task *Task) Remain() int64 {
	return max(0, task.expected-task.finish)
}
func (task *Task) Name() string {
	return task.name
}
func (task *Task) Sign() string {
	return task.sign
}
func (task *Task) Model() string {
	return ModelTypeToString(task.model)
}
func (task *Task) Expected() int64 {
	return task.expected
}
func (task *Task) FromRowData(data map[string]interface{}) {
	if id, ok := data["id"]; ok {
		task.taskID = id.(TaskID)
	}
	if sign, ok := data["sign"]; ok {
		task.sign = sign.(string)
	}
	if name, ok := data["name"]; ok {
		task.name = name.(string)
	}

	if model, ok := data["model"]; ok {
		task.model = NameToModelType(model.(string))
	}
	if expected, ok := data["expected"]; ok {
		task.expected = expected.(int64)
	}
	if finish, ok := data["finish"]; ok {
		task.finish = finish.(int64)
	}
	if txHash, ok := data["txHash"]; ok {
		task.txHash = txHash.(string)
	}
}
func (t *Task) NotBeenScheduled() bool {
	return t.isCreate
}

// SimpleTaskFromSign since sign is known, so this task has been scheduled
func SimpleTaskFromSign(sign string) *Task {
	return &Task{
		taskID:   -1,
		sign:     sign,
		name:     "",
		model:    NOTIMPORTANT,
		expected: 0,
		finish:   0,
		txHash:   "",
		isCreate: false,
	}
}

func NewTask(name string, model SupportModelType, expected int64) *Task {
	hash := sha256.New()
	hash.Write([]byte(name))
	hash.Write([]byte(fmt.Sprintf("%d", expected)))
	hash.Write([]byte(ModelTypeToString(model)))
	hash.Write([]byte(time.Now().Format(time.RFC3339Nano)))
	hash.Write([]byte(fmt.Sprintf("%d", rand.Int())))
	return &Task{
		taskID:   -1,
		sign:     hex.EncodeToString(hash.Sum([]byte{})),
		name:     name,
		model:    model,
		expected: expected,
		finish:   0,
		txHash:   "",
		isCreate: true,
	}
}

func DefaultTaskForTest() *Task {
	return NewTask("Test Task", CTGAN, 1000)
}
