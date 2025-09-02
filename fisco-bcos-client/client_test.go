package fisco_bcos_client

import (
	"RappaMaster/config"
	"RappaMaster/task"
	"RappaMaster/transaction"
	"fmt"
	"testing"
)

var clt *RappaFBClient

func TestMain(m *testing.M) {
	c := config.FBConfig{}
	c.SetDefault()
	var err error
	clt, err = NewRappaFBClient(c)
	if err != nil {
		panic(err.Error())
	}

	m.Run()
}

func TestSendWithSync(t *testing.T) {
	receipt, err := clt.SendWithSync(generateTransaction())
	if err != nil {
		t.Errorf("SendWithSync err: %s", err.Error())
	}
	fmt.Printf("%+v\n", receipt)
}

func generateTransaction() transaction.Transaction {
	tasks := []task.Task{*task.DefaultTaskForTest()}
	return transaction.NewInitTaskTransaction(tasks)
}
