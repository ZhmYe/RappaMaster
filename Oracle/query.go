package Oracle

import (
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"fmt"
)

// TODO 这里会有很多并发读写的问题，后面要用TryLock,如果发现被锁住了，那么直接返回还未更新
func (d *Oracle) processQuery() {
	// 来自Http的query,需要发回去，因此Query里需要有一个通道
	// TODO 这里的txMap应该会有并发读写的问题，而且不能将这个协程合并，会影响性能
	for query := range d.channel.QueryChannel {
		paradigm.Print("ORACLE", fmt.Sprintf("Receive a Query: %v", query.ToHttpJson()))
		switch query.(type) {
		case *Query.EvidencePreserveTaskIDQuery:
			item := query.(*Query.EvidencePreserveTaskIDQuery)
			if task, exist := d.tasks[item.TaskID]; !exist {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Task does not exist in Oracle"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Task does not exist in Oracle")
				continue
			} else {
				item.SendResponse(item.GenerateResponse(task)) // 有ref一定有task
			}
		case *Query.EvidencePreserveTaskTxQuery:
			// 根据txHash查询Task
			item := query.(*Query.EvidencePreserveTaskTxQuery)
			if _, exist := d.txMap[item.TxHash]; !exist {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Transaction does not exist in Oracle"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Transaction does not exist in Oracle")
				continue
			}
			ref := d.txMap[item.TxHash]
			if ref.Rf == paradigm.EpochTx {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "%s is a EpochUpdate Transaction, not a Task-related Transaction"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "%s is a EpochUpdate Transaction, not a Task-related Transaction")
			} else {
				// 如果是Slot或者initTask，那么都会有对应的TaskID
				if ref.TaskID == "" {
					errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "runtime error"))
					item.SendResponse(errorResponse)
					paradigm.Error(paradigm.RuntimeError, "Reference has not TaskID But is not a EpochTx Rf")
					continue
				}
				item.SendResponse(item.GenerateResponse(d.tasks[ref.TaskID])) // 有ref一定有task
			}
		case *Query.EvidencePreserveEpochIDQuery:
			item := query.(*Query.EvidencePreserveEpochIDQuery)
			if epoch, exist := d.epochs[item.EpochID]; !exist {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Epoch does not exist in Oracle"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Epoch does not exist in Oracle")
				continue
			} else {
				//fmt.Println(111, epoch)
				item.SendResponse(item.GenerateResponse(epoch)) // 有ref一定有task
			}
		case *Query.EvidencePreserveEpochTxQuery:
			item := query.(*Query.EvidencePreserveEpochTxQuery)
			if _, exist := d.txMap[item.TxHash]; !exist {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Transaction does not exist in Oracle"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Transaction does not exist in Oracle")
				continue
			}
			ref := d.txMap[item.TxHash]
			// 无论如何ref.EpochID都是有的
			if ref.EpochID == -1 {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "runtime error"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError, "Reference has no EpochID")
				continue
			}
			// 有ref但是没有epoch，要么是代码有问题，要么是交易对应的是taskProcess，此时epoch还没更新 TODO
			if epoch, exist := d.epochs[ref.EpochID]; !exist {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Epoch has not been update"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError, "Epoch has not been update")
			} else {
				item.SendResponse(item.GenerateResponse(epoch))
			}
		case *Query.BlockchainLatestInfoQuery:
			item := query.(*Query.BlockchainLatestInfoQuery)
			info := paradigm.LatestBlockchainInfo{
				LatestTxs:     d.latestTxs,
				LatestEpoch:   d.latestEpochs,
				NbFinalized:   d.nbFinalized,
				SynthData:     d.synthData,
				NbEpoch:       int32(len(d.epochs)),
				NbBlock:       int32(d.latestTxs[len(d.latestTxs)-1].Receipt.BlockNumber), // TODO
				NbTransaction: int32(len(d.txMap)),                                        // TODO
			}
			item.SendResponse(item.GenerateResponse(info))
		case *Query.BlockchainBlockNumberQuery:
			item := query.(*Query.BlockchainBlockNumberQuery)
			d.channel.BlockchainQueryChannel <- item
			block := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(block))
		case *Query.BlockchainBlockHashQuery:
			item := query.(*Query.BlockchainBlockHashQuery)
			d.channel.BlockchainQueryChannel <- item
			block := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(block))
		case *Query.BlockchainTransactionQuery: // TODO: 改为链上查询
			item := query.(*Query.BlockchainTransactionQuery)
			// if _, exist := d.txMap[item.TxHash]; !exist {
			// 	errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Transaction does not exist in Oracle"))
			// 	item.SendResponse(errorResponse)
			// 	paradigm.Error(paradigm.ValueError, "Transaction does not exist in Oracle")
			// 	continue
			// }
			// ref := d.txMap[item.TxHash]
			// item.SendResponse(item.GenerateResponse(ref))
			d.channel.BlockchainQueryChannel <- item
			tx := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(tx))
		case *Query.NodesStatusQuery:
			item := query.(*Query.NodesStatusQuery)
			d.channel.MonitorQueryChannel <- item
			status := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(status))
		case *Query.DateSynthDataQuery:
			item := query.(*Query.DateSynthDataQuery)
			records := d.dates
			item.SendResponse(item.GenerateResponse(records))
		case *Query.DateTransactionQuery:
			item := query.(*Query.DateTransactionQuery)
			records := d.dates
			item.SendResponse(item.GenerateResponse(records))
		case *Query.SynthTaskQuery:
			item := query.(*Query.SynthTaskQuery)
			tasks := d.tasks
			item.SendResponse(item.GenerateResponse(tasks))
		case *Query.CollectTaskQuery:
			item := query.(*Query.CollectTaskQuery)
			if task, exist := d.tasks[item.TaskID()]; !exist {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Task does not exist in Oracle"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Task does not exist in Oracle")
				continue
			} else {
				go item.SendResponse(item.GenerateResponse(task.GetCollector())) // 有ref一定有task
			}
		default:
			panic("Unsupported Query Type!!!")
		}
	}
}
