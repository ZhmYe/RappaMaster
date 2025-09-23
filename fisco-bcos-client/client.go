package fisco_bcos_client

import (
	"RappaMaster/config"
	"RappaMaster/fisco-bcos-client/contract/store"
	"RappaMaster/paradigm"
	"RappaMaster/transaction"
	"encoding/hex"
	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/net/context"
	"sync"
	"time"
)

var defaultMockerReceipt = types.Receipt{
	BlockNumber:     -1,
	ContractAddress: "",
	From:            "",
	GasUsed:         "",
	Hash:            "",
	Input:           "",
	Logs:            nil,
	Message:         "",
	Output:          "",
	Status:          0,
	To:              "",
	TransactionHash: hex.EncodeToString([]byte("test transaction hash")),
	ReceiptProof:    nil,
	Version:         0,
}

type RappaFBClient struct {
	*client.Client
	contractInstance *Store.Store
	contractAddress  common.Address
}

func (c *RappaFBClient) SendWithSync(tx transaction.Transaction) (types.Receipt, error) {
	return c.processTransaction(tx)
}
func (c *RappaFBClient) SendWithAsync(tx transaction.Transaction, wg *sync.WaitGroup) (types.Receipt, error) {
	defer wg.Done()
	return c.processTransaction(tx)
}

func (c *RappaFBClient) processTransaction(tx transaction.Transaction) (types.Receipt, error) {
	if config.DEBUG {
		// in debug mode, we just return a fake receipt
		return defaultMockerReceipt, nil
	}
	keys, values := tx.CallData()
	storeSession := &Store.StoreSession{Contract: c.contractInstance, CallOpts: *c.GetCallOpts(), TransactOpts: *c.GetTransactOpts()}
	_, _receipt, err := storeSession.SetItems(keys, values)
	if err != nil {
		return types.Receipt{}, paradigm.RaiseError(paradigm.UpchainError, "client failed to call SetItems", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	receipt, err := c.GetTransactionReceipt(ctx, common.HexToHash(_receipt.TransactionHash), true)
	if err != nil {
		return types.Receipt{}, paradigm.RaiseError(paradigm.UpchainError, "client failed to call SetItems", err)
	}
	return *receipt, nil
}

func NewRappaFBClient(c config.FBConfig) (*RappaFBClient, error) {
	if config.DEBUG {
		return &RappaFBClient{}, nil
	}
	privateKey, _ := hex.DecodeString(c.PrivateKey)
	client, err := client.DialContext(context.Background(), &client.Config{
		IsSMCrypto:  false,
		GroupID:     c.GroupID,
		PrivateKey:  privateKey,
		Host:        c.FiscoBcosHost,
		Port:        c.FiscoBcosPort,
		TLSCaFile:   c.TLSCaFile,
		TLSCertFile: c.TLSCertFile,
		TLSKeyFile:  c.TLSKeyFile,
	})
	if err != nil {
		return nil, paradigm.RaiseError(paradigm.UpchainError, "client failed to dial client", err)
	}
	address, _, instance, err := Store.DeployStore(client.GetTransactOpts(), client)
	if err != nil {
		return nil, paradigm.RaiseError(paradigm.UpchainError, "client failed to call DeployStore", err)
	}
	return &RappaFBClient{
		Client:           client,
		contractInstance: instance,
		contractAddress:  address,
	}, nil
}
