package ChainUpper

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
	"sync"
	"time"
)

type MockerChainUpper struct {
	pendingTransactions chan paradigm.Transaction // 交易channel
	transactionPool     []paradigm.Transaction    // 所有的交易
	unprocessedIndex    int                       // 未处理的交易index，包括这一index
	mu                  sync.Mutex
}

func (c *MockerChainUpper) Start() {
	timeStart := time.Now()
	go func() {
		for {
			if time.Since(timeStart) >= 10*time.Second {
				timeStart = time.Now()
				c.UpChain()
			}
		}
	}()
	for {
		//LogWriter.Log("DEBUG", "111")
		select {
		case transaction := <-c.pendingTransactions:
			// 先简单写一下
			c.mu.Lock()
			c.transactionPool = append(c.transactionPool, transaction)
			c.mu.Unlock()
		default:
			continue
		}
	}
}
func (c *MockerChainUpper) UpChain() {
	// 这里简单写一下，应该是用异步上链组件接入这里
	pack := func() []paradigm.Transaction {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.unprocessedIndex == len(c.transactionPool) {
			return []paradigm.Transaction{}
		}
		packedTransactions := c.transactionPool[c.unprocessedIndex:]
		c.unprocessedIndex = len(c.transactionPool)
		return packedTransactions
	}
	packedTransaction := pack()
	if len(packedTransaction) > 0 {
		LogWriter.Log("CHAINUP", fmt.Sprintf("up %d transactions to blockchain...", len(packedTransaction)))
	} else {
		LogWriter.Log("WARNING", "Nothing to up to Blockchain..., len(transactionPool) = 0")
	}
}
func NewMockerChainUpper(pendingTransactions chan paradigm.Transaction) *MockerChainUpper {
	return &MockerChainUpper{
		pendingTransactions: pendingTransactions,
		transactionPool:     make([]paradigm.Transaction, 0),
		unprocessedIndex:    0,
		mu:                  sync.Mutex{},
	}
}
