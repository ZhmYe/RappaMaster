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
	keys, values := tx.CallData()
	// LogWriter.Log("DEBUG", fmt.Sprintf("TX_%d convert to:[key]%s [value]%s", tType, keys, values))
	storeSession := &Store.StoreSession{Contract: c.contractInstance, CallOpts: *c.GetCallOpts(), TransactOpts: *c.GetTransactOpts()}
	// _, receipt, err := storeSession.SetItem(key, value)
	_, _receipt, err := storeSession.SetItems(keys, values)
	if err != nil {
		return types.Receipt{}, paradigm.RaiseError(paradigm.UpchainError, "client failed to call SetItems", err)
	}
	// 获得有merkleProof的receipt
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	receipt, err := c.GetTransactionReceipt(ctx, common.HexToHash(_receipt.TransactionHash), true)
	if err != nil {
		return types.Receipt{}, paradigm.RaiseError(paradigm.UpchainError, "client failed to call SetItems", err)
	}
	//} else {
	//	// LogWriter.Log("DEBUG", fmt.Sprintf("Receipt with merkleProof: %s", _receipt)) //debug
	//	// block, _ := w.client.GetBlockByNumber(ctx, int64(_receipt.BlockNumber), false, false)
	//	blockHash, _ := c.GetBlockHashByNumber(ctx, int64(_receipt.BlockNumber))
	//	ptxs := packedParam.BuildDevTransactions([]*types.Receipt{_receipt}, blockHash.Hex())
	//	w.devPackedTransaction <- ptxs // 传递到dev
	//}
	return *receipt, nil
}

func NewRappaFBClient(config config.FBConfig) (*RappaFBClient, error) {
	privateKey, _ := hex.DecodeString(config.PrivateKey)
	client, err := client.DialContext(context.Background(), &client.Config{
		IsSMCrypto:  false,
		GroupID:     config.GroupID,
		PrivateKey:  privateKey,
		Host:        config.FiscoBcosHost,
		Port:        config.FiscoBcosPort,
		TLSCaFile:   config.TLSCaFile,
		TLSCertFile: config.TLSCertFile,
		TLSKeyFile:  config.TLSKeyFile,
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
