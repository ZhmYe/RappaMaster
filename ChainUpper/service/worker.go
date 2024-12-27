package service

import (
	SlotCommit "BHLayer2Node/ChainUpper/contract/slotCommit"
	"BHLayer2Node/LogWriter"
	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"math/big"
	"fmt"
)

func Worker(id int, queue chan map[string]interface{}, instance *SlotCommit.SlotCommit, client *client.Client) {
	// 每次收集 batchsize 个transaction再一起上链，但是实际产生的数据很少 设置为1，收到后直接上链
	batchSize := 1 
	// 上链参数
	var signs [][32]byte
	var slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int

	for {
		select {
		case result := <-queue: // 尝试从通道中接收数据
			if result != nil { // 判断是否接收到有效值
				// log.Printf("Worker %d Received result: %v", id, result)
				LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d Received Transaction: %v", id, result))

				sign := result["Sign"].(string)
				slot := result["Slot"].(int32)
				process := result["Process"].(int32)
				nid := result["ID"].(int32)
				epoch := result["Epoch"].(int)

				// 转换 sign 为 bytes32
				signBytes32 := toBytes32(sign)

				signs = append(signs, signBytes32)
				slotsBigInt = append(slotsBigInt, new(big.Int).SetUint64(uint64(slot)))
				processesBigInt = append(processesBigInt, new(big.Int).SetUint64(uint64(process)))
				nidsBigInt = append(nidsBigInt, new(big.Int).SetUint64(uint64(nid)))
				epochsBigInt = append(epochsBigInt, new(big.Int).SetUint64(uint64(epoch)))

				// 每当收集到batchSize个transaction的信息时，调用批量上链函数
				if len(signs) >= batchSize {
					consumerImpl(signs, slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt, instance, client)
					LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d completed Batch transactions Upchain. Count: %d\n", id, len(signs)))
					// 检查上链的数据是否成功 这边signs应该是有重复的，后续可以去重再查询
					// getTransactionImpl(signs, instance, client)
					// 清空数组
					signs = nil
					slotsBigInt = nil
					processesBigInt = nil
					nidsBigInt = nil
					epochsBigInt = nil

				}
				// LogWriter.Log("CHAINUP", fmt.Sprintf("Worker %d: finished Transaction: %v\n", id, result))
			} else {
				LogWriter.Log("ERROR", fmt.Sprintf("Upchain channel closed, received nil value"))
				return
			}
		}
	}
}

// 调用合约函数进行批量数据上链
func consumerImpl(signs [][32]byte, slotsBigInt []*big.Int, processesBigInt []*big.Int, nidsBigInt []*big.Int, epochsBigInt []*big.Int, instance *SlotCommit.SlotCommit, client *client.Client) {
	slotCommitSession := &SlotCommit.SlotCommitSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}

	_, _, err := slotCommitSession.CommitSlotsBatch(
		signs,
		slotsBigInt,
		processesBigInt,
		nidsBigInt,
		epochsBigInt,
	)
	if err != nil {
		LogWriter.Log("ERROR", fmt.Sprintf("Failed to call CommitSlotsBatch: %v", err))
	}
}


func toBytes32(s string) [32]byte {
	var b [32]byte
	copy(b[:], s)
	return b
}

func getTransactionImpl(signs [][32]byte, instance *SlotCommit.SlotCommit, client *client.Client) {
	slotCommitSession := &SlotCommit.SlotCommitSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}
	for id, sign := range signs {
		ret, err := slotCommitSession.GetSlotCommits(sign)
		if err != nil {
			LogWriter.Log("ERROR", fmt.Sprintf("Failed to call getSlotBySign %d %v: %v", id, sign, err))
		}
		LogWriter.Log("CHAINUP", fmt.Sprintf("Sign %d: %v, get chain transaction: %v", id, sign, ret))
	}
	
}