package database

import (
	"RappaMaster/config"
	types2 "RappaMaster/types"
	"encoding/hex"
	"fmt"
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
	tsk := types2.DefaultTaskForTest()
	receipt := generateTestReceipt("test task hash")

	if err := dbs.CreateTask(*tsk, receipt); err != nil {
		t.Fatalf("create task failed: %v", err)
	}
	t.Log("create task success")
}

func TestDatabaseGetTaskWithSign(t *testing.T) {
	//sign := task.DefaultTaskForTest().Sign()
	if tsk, err := dbs.GetTaskBySign("dbbaf9823b9d43102f7dfbf4b6ad92c85f9a2fc26954c7122622208fc56701bb"); err != nil {
		t.Fatalf("get task by sign failed: %v", err)
	} else {
		fmt.Printf("%+v\n", tsk)
	}
	t.Log("get task by sign success")
}

func TestDatabaseGetEpoch(t *testing.T) {
	if epochID, err := dbs.GetCurrentEpoch(); err != nil {
		t.Fatalf("get current epoch failed: %v", err)
	} else if epochID == -1 {
		t.Fatalf("get current epoch fail: epochID == -1")
	} else {
		fmt.Printf("epochID: %d\n", epochID)
	}
	t.Log("get epoch success")
}

func generateTestReceipt(data string) types.Receipt {
	hasher := sha3.New256()
	hasher.Write([]byte(data))
	return types.Receipt{
		TransactionHash: hex.EncodeToString(hasher.Sum([]byte{})),
	}
}
