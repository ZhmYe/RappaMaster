package Oracle

import (
	"BHLayer2Node/Date"
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// TODO 这里会有很多并发读写的问题，后面要用TryLock,如果发现被锁住了，那么直接返回还未更新
func (d *PersistedOracle) processDBQuery() {
	// 来自Http的query,需要发回去，因此Query里需要有一个通道
	// TODO 这里的txMap应该会有并发读写的问题，而且不能将这个协程合并，会影响性能
	for query := range d.channel.QueryChannel {
		paradigm.Print("ORACLE", fmt.Sprintf("Receive a Query: %v", query.ToHttpJson()))
		switch query.(type) {
		case *Query.EvidencePreserveTaskIDQuery:
			item := query.(*Query.EvidencePreserveTaskIDQuery)
			task := &paradigm.Task{}
			if err := d.db.Where("sign = ?", item.TaskID).First(task).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Task does not exist in database"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Task does not exist in database")
				continue
			}
			item.SendResponse(item.GenerateResponse(task))
		case *Query.EvidencePreserveTaskTxQuery: // 未检测
			// 根据txHash查询Task
			item := query.(*Query.EvidencePreserveTaskTxQuery)
			ref := &paradigm.DevReference{}
			if err := d.db.Where("tx_hash = ?", item.TxHash).First(ref).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Transaction does not exist in database"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Transaction does not exist in database")
				continue
			}
			if ref.Rf == paradigm.EpochTx {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Is a EpochRecord Transaction"))
				item.SendResponse(errorResponse)
				continue
			}
			// 查询关联的任务
			task := &paradigm.Task{}
			if err := d.db.Where("sign = ?", ref.TaskID).First(task).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "Associated task not found"))
				item.SendResponse(errorResponse)
				continue
			}
			item.SendResponse(item.GenerateResponse(task))
		case *Query.EvidencePreserveEpochIDQuery:
			item := query.(*Query.EvidencePreserveEpochIDQuery)
			epoch := &paradigm.DevEpoch{}
			if err := d.db.Where("epoch_id = ?", item.EpochID).First(epoch).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Epoch does not exist in database"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Epoch does not exist in database")
				continue
			}
			//test
			paradigm.Log("DEBUG", fmt.Sprintf("epoch search by ID ret: %+v", epoch))

			item.SendResponse(item.GenerateResponse(epoch))
		case *Query.EvidencePreserveEpochTxQuery:
			item := query.(*Query.EvidencePreserveEpochTxQuery)
			ref := &paradigm.DevReference{}
			if err := d.db.Where("tx_hash = ?", item.TxHash).First(ref).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Transaction does not exist in database"))
				item.SendResponse(errorResponse)
				continue
			}
			if ref.Rf == paradigm.InitTaskTx {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Is a InitTask Transaction"))
				item.SendResponse(errorResponse)
				continue
			}
			if ref.EpochID == -1 {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "No EpochID found"))
				item.SendResponse(errorResponse)
				continue
			}
			// 查询关联的 epoch
			epoch := &paradigm.DevEpoch{}
			if err := d.db.Where("epoch_id = ?", ref.EpochID).First(epoch).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Epoch not found"))
				item.SendResponse(errorResponse)
				continue
			}
			item.SendResponse(item.GenerateResponse(epoch))
		case *Query.BlockchainLatestInfoQuery:
			item := query.(*Query.BlockchainLatestInfoQuery)
			// 1. 获取最新的20笔交易
			var latestTxRefs []paradigm.DevReference
			if err := d.db.Order("upchain_time desc").Limit(20).Find(&latestTxRefs).Error; err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to query latest transactions: %v", err))
				continue
			}

			// 构建 PackedTransaction 列表
			latestTxs := make([]*paradigm.PackedTransaction, 0)
			for _, ref := range latestTxRefs {
				packedTx := &paradigm.PackedTransaction{
					Receipt:     &ref.TxReceipt,
					BlockHash:   ref.TxBlockHash,
					UpchainTime: ref.UpchainTime,
					Id:          int(ref.TID),
				}
				latestTxs = append(latestTxs, packedTx)
			}

			// 2. 获取最新的20个epoch
			var latestEpochs []*paradigm.DevEpoch
			if err := d.db.Order("epoch_id desc").Limit(20).Find(&latestEpochs).Error; err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to query latest epochs: %v", err))
				continue
			}

			// 3. 获取所有已完成(finalized)的slot数量
			var nbFinalized int64
			if err := d.db.Model(&paradigm.Slot{}).Where("status = ?", "finalized").Count(&nbFinalized).Error; err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to count finalized slots: %v", err))
				continue
			}

			// 4. 按模型类型统计合成数据总量
			synthData := make(map[paradigm.SupportModelType]int32)
			var tasks []*paradigm.Task
			if err := d.db.Find(&tasks).Error; err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to query tasks: %v", err))
				continue
			}

			for _, task := range tasks {
				if task.IsFinish() {
					if value, ok := synthData[task.Model]; ok {
						synthData[task.Model] = value + task.Process
					} else {
						synthData[task.Model] = task.Process
					}
				}
			}

			// 5. 获取最新区块号
			var nbBlock int32
			if len(latestTxs) > 0 {
				nbBlock = int32(latestTxs[0].Receipt.BlockNumber)
			}

			// 6. 获取总交易数
			var nbTransaction int64
			if err := d.db.Model(&paradigm.DevReference{}).Count(&nbTransaction).Error; err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to count transactions: %v", err))
				continue
			}

			// 构建响应信息
			info := paradigm.LatestBlockchainInfo{
				LatestTxs:     latestTxs,
				LatestEpoch:   latestEpochs,
				NbFinalized:   int32(nbFinalized),
				SynthData:     synthData,
				NbEpoch:       int32(len(latestEpochs)),
				NbBlock:       nbBlock,
				NbTransaction: int32(nbTransaction),
			}

			// 发送响应
			item.SendResponse(item.GenerateResponse(info))
		case *Query.BlockchainBlockNumberQuery: // 链上查询
			item := query.(*Query.BlockchainBlockNumberQuery)
			d.channel.BlockchainQueryChannel <- item
			block := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(block))
		case *Query.BlockchainBlockHashQuery: // 链上查询
			item := query.(*Query.BlockchainBlockHashQuery)
			d.channel.BlockchainQueryChannel <- item
			block := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(block))
		case *Query.BlockchainTransactionQuery: // 链上查询 未通过
			item := query.(*Query.BlockchainTransactionQuery)
			// 从区块链获取交易信息
			d.channel.BlockchainQueryChannel <- item
			tx := item.ReceiveInfo()
			txInfo := tx.(paradigm.TransactionInfo)
			// 从数据库查询交易引用信息
			ref := &paradigm.DevReference{}
			if err := d.db.Where("tx_hash = ?", txInfo.TxHash).First(ref).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					errorResponse := paradigm.NewErrorResponse(
						paradigm.NewRappaError(paradigm.ValueError,
							"Transaction does not exist in database"))
					item.SendResponse(errorResponse)
					paradigm.Error(paradigm.ValueError,
						fmt.Sprintf("Transaction %s does not exist in database", txInfo.TxHash))
					continue
				}
				// 处理其他数据库错误
				errorResponse := paradigm.NewErrorResponse(
					paradigm.NewRappaError(paradigm.RuntimeError,
						fmt.Sprintf("Database error: %v", err)))
				item.SendResponse(errorResponse)
				continue
			}

			// 根据引用类型设置 ABI
			switch ref.Rf {
			case paradigm.InitTaskTx:
				txInfo.Abi = "InitTask"
			case paradigm.SlotTX:
				txInfo.Abi = "TaskProcess"
			case paradigm.EpochTx:
				txInfo.Abi = "EpochRecord"
			default:
				errorResponse := paradigm.NewErrorResponse(
					paradigm.NewRappaError(paradigm.RuntimeError,
						"Unknown transaction reference type"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError,
					fmt.Sprintf("Unknown transaction reference type: %v", ref.Rf))
				continue
			}
			item.SendResponse(item.GenerateResponse(txInfo))
		case *Query.NodesStatusQuery:
			item := query.(*Query.NodesStatusQuery)
			d.channel.MonitorQueryChannel <- item
			status := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(status))
		case *Query.DateSynthDataQuery:
			item := query.(*Query.DateSynthDataQuery)
			var records []*Date.DateRecord
			if err := d.db.Find(&records).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "Failed to query date records"))
				item.SendResponse(errorResponse)
				continue
			}
			item.SendResponse(item.GenerateResponse(records))
		case *Query.DateTransactionQuery:
			item := query.(*Query.DateTransactionQuery)
			var records []*Date.DateRecord
			if err := d.db.Find(&records).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "Failed to query date records"))
				item.SendResponse(errorResponse)
				continue
			}
			item.SendResponse(item.GenerateResponse(records))
		case *Query.SynthTaskQuery:
			item := query.(*Query.SynthTaskQuery)
			var tasks []*paradigm.Task
			if err := d.db.Find(&tasks).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "Failed to query tasks"))
				item.SendResponse(errorResponse)
				continue
			}
			// 转换为 map 形式
			tasksMap := make(map[string]*paradigm.Task)
			for _, task := range tasks {
				tasksMap[task.Sign] = task
			}
			item.SendResponse(item.GenerateResponse(tasksMap))
		case *Query.CollectTaskQuery:
			item := query.(*Query.CollectTaskQuery)
			task := &paradigm.Task{}
			if err := d.db.Where("sign = ?", item.TaskID()).First(task).Error; err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, "Task does not exist in database"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, "Task does not exist in database")
				continue
			}
			go item.SendResponse(item.GenerateResponse(task.GetCollector()))
		default:
			panic("Unsupported Query Type!!!")
		}
	}
}
