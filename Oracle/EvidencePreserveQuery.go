package Oracle

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"sort"
)

/***
	存证溯源界面
***/

// EvidencePreserveTaskTxQuery 根据交易哈希查询Task
// 查询Http Json格式:
/***
{
	"query": "EvidencePreserveTaskTxQuery",
	"txHash": "0x123456",
}
***/
// 传入Task生成response
type EvidencePreserveTaskTxQuery struct {
	txHash          string // 交易哈希
	responseChannel chan paradigm.Response
}

func (q *EvidencePreserveTaskTxQuery) GenerateResponse(data interface{}) paradigm.Response {
	task := data.(*paradigm.Task)
	// 要展示的内容
	// 1.任务的基本信息
	info := make(map[string]interface{})
	info["taskID"] = task.Sign             // 任务标识
	info["total"] = task.Size              // 任务合成总量
	info["process"] = task.Process         // 已合成的数量
	info["schedule"] = len(task.Schedules) // 调度数量
	info["commit"] = len(task.Schedules)   // 提交的slot的数量 todo
	info["params"] = task.Params           // 任务的一些参数，包括模型、输入、数据集等，待定 todo
	// 2. 交易的基本信息
	tx := make(map[string]interface{})
	tx["txHash"] = task.TxReceipt.TransactionHash          // 交易哈希
	tx["blockHash"] = task.TxReceipt.BlockNumber           // 区块哈希 TODO @XQ 如果Receipt里没有，那么就想办法在上链的时候拿到然后放到task里
	tx["blockHeight"] = task.TxReceipt.BlockNumber         // 区块高度
	tx["contractAddress"] = task.TxReceipt.ContractAddress // 合约地址
	tx["abi"] = "InitTask"                                 // todo @XQ 根据你的合约修改，这里因为是查询Task，所以就是Task初始化的那个交易接口
	tx["MerkleRoot"] = task.TxReceipt.ReceiptProof         // todo @XQ 看一下这个东西要怎么拿到root以及怎么验证，这里暂时后面要改成只有一个root
	tx["MerkleProof"] = task.TxReceipt.ReceiptProof        // todo @XQ 这里就是正常的proof
	// 3. 时间轴 展示于前端左下角，即该Task在不同epoch里的具体情况
	// 根据epoch得到时间轴，就是一些字符串 TODO 看一下具体的格式
	timeline := make([][2]string, 0) // 这里暂时就先写成这样
	// 第一个时间是initTask的时间 TODO @XQ 所以每一笔交易在上链的时候都要记录时间，可以作为一个字段放在Task/Slot/Epoch中
	timeline = append(timeline, [2]string{"initTask Time", fmt.Sprintf("向区块链提交合成任务, 任务标识: %s, 交易哈希: %s", task.Sign, task.TxReceipt.TransactionHash)}) // 这里的时间要换成具体的时间，然后后面的文字就中文吧，因为前端要求是中文
	// 然后根据epoch来写,写在下面
	totalCommitNumber, totalInvalidNumber, totalSlotNumber := 0, 0, 0
	epochs := make([]int32, 0)               // 这里是epoch的列表，因为会有一些epoch是空的，所以这里后面要排序，下面用一个map
	epochSlotMap := make(map[int32][3]int32) // 这里映射某个epoch该任务的提交数量和invalid数量和合成总量
	//4. Schedule 列表
	schedules := make([]map[interface{}]interface{}, 0)
	for _, schedule := range task.Schedules {
		scheduleInfo := make(map[interface{}]interface{})
		scheduleInfo["scheduleID"] = schedule.ScheduleID         // 这里前端要加上“Schedule”
		scheduleInfo["scheduleSize"] = schedule.Size             // 调度的总量
		scheduleInfo["scheduleNumber"] = len(schedule.NodeIDMap) // 调度节点的数量
		totalSlotNumber += len(schedule.Slots)
		process, nbCommit, nbInvalid := int32(0), 0, 0
		slots := make([]map[interface{}]interface{}, 0)
		// 所有的slot
		for _, slot := range schedule.Slots {
			epoch := slot.Epoch
			if epoch != -1 {
				if _, exist := epochSlotMap[epoch]; !exist {
					epochSlotMap[epoch] = [3]int32{0, 0, 0}
					epochs = append(epochs, epoch)
				}
			}
			if slot.Status == paradigm.Finished {
				nbCommit++
				process += slot.CommitSlot.Process
				totalCommitNumber++
				result := epochSlotMap[epoch]
				result[2] += slot.CommitSlot.Process
				result[0]++
				epochSlotMap[epoch] = result

			}
			if slot.Status == paradigm.Failed {
				nbInvalid++
				totalInvalidNumber++
				result := epochSlotMap[epoch]
				result[1]++
				epochSlotMap[epoch] = result
			}
			slots = append(slots, slot.Json())
		}
		scheduleInfo["process"] = process
		scheduleInfo["commitNumber"] = nbCommit
		scheduleInfo["invalidNumber"] = nbInvalid
		schedules = append(schedules, scheduleInfo)
	}
	// 5. 合成进度，每个epoch一共合成了多少数据
	epochProcess := make([]int32, 0) // TODO

	// 6. 调度完成情况
	scheduleDistribution := [3]int{totalCommitNumber, totalInvalidNumber, totalSlotNumber - totalCommitNumber - totalInvalidNumber}
	// 处理timeline
	sort.Slice(epochs, func(i int, j int) bool {
		return epochs[i] < epochs[j]
	})
	for _, epoch := range epochs {
		epochData := epochSlotMap[epoch]
		nbCommitInEpoch, nbInvalidInEpoch, synthData := epochData[0], epochData[1], epochData[2]
		timeline = append(timeline, [2]string{fmt.Sprintf("Epoch %d", epoch), fmt.Sprintf("提交单元数: %d, 检测异常单元数: %d, 合成数据: %d ", nbCommitInEpoch, nbInvalidInEpoch, synthData)}) // TODO
		epochProcess = append(epochProcess, synthData)
	}
	response := make(map[interface{}]interface{})
	response["task_info"] = info
	response["tx_info"] = tx
	response["timeline"] = timeline
	response["schedules"] = schedules
	response["epochs"] = epochs
	response["epochData"] = epochSlotMap
	response["epochProcessData"] = epochProcess
	response["scheduleDistributionData"] = scheduleDistribution
	return paradigm.NewSuccessResponse(response)

}
func (q *EvidencePreserveTaskTxQuery) ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool {
	if txHash, ok := rawData["txHash"].(string); ok {
		q.txHash = txHash
		return true
	}
	return false
}
func (q *EvidencePreserveTaskTxQuery) SendResponse(response paradigm.Response) {
	q.responseChannel <- response
	close(q.responseChannel)
}
func (q *EvidencePreserveTaskTxQuery) ReceiveResponse() paradigm.Response {
	return <-q.responseChannel
}
func NewEvidencePreserveTaskTxQuery(rawData map[interface{}]interface{}) *EvidencePreserveTaskTxQuery {
	responseChannel := make(chan paradigm.Response, 1)
	query := new(EvidencePreserveTaskTxQuery)
	query.ParseRawDataFromHttpEngine(rawData)
	query.responseChannel = responseChannel
	return query
}

// EvidencePreserveTaskIDQuery 根据任务ID查询Task
type EvidencePreserveTaskIDQuery struct {
}

// EvidencePreserveEpochIDQuery 根据EpochID查询Epoch
type EvidencePreserveEpochIDQuery struct {
}

// EvidencePreserveEpochTxQuery 根据交易哈希查询Epoch
type EvidencePreserveEpochTxQuery struct {
}
