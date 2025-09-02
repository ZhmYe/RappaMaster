package database

import (
	"RappaMaster/config"
	"RappaMaster/task"
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"testing"

	"github.com/FISCO-BCOS/go-sdk/v3/types"
)

var dbs *DatabaseService

func TestMain(m *testing.M) {
	c := config.DatabaseConfig{}
	c.SetDefault()
	dbs = NewDatabaseService(c)
	if err := dbs.Init(); err != nil {
		panic(err.Error())
	}

	m.Run()
}

func TestDatabase(t *testing.T) {
	if dbs == nil {
		t.Fatal("database not initialized")
	}
	t.Log("database initialized")
}

func TestDatabaseCreateTask(t *testing.T) {
	tsk := task.DefaultTaskForTest()
	receipt := generateTestReceipt("test task hash")

	// 执行测试逻辑
	if err := dbs.CreateTask(*tsk, receipt); err != nil {
		t.Fatalf("create task failed: %v", err)
	}
	t.Log("create task success")
}

func generateTestReceipt(data string) types.Receipt {
	hasher := sha3.New256()
	hasher.Write([]byte(data))
	return types.Receipt{
		TransactionHash: hex.EncodeToString(hasher.Sum([]byte{})),
	}
}
