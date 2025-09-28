package fisco_bcos_client

import (
	"RappaMaster/config"
	"RappaMaster/types"
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

func generateTransaction() types.Transaction {
	tasks := []types.Task{*types.DefaultTaskForTest()}
	return types.NewInitTaskTransaction(tasks)
}
