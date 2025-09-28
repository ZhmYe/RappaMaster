package epoch

import (
	"RappaMaster/config"
	"RappaMaster/helper"
	"RappaMaster/types"
	"context"
	"time"
)

// EpochManager advances new epochs, when a timeout is exceeded
// It create a new epoch into db, and get the information of the last epoch to generate a heartbeat
type EpochManager struct {
	d         time.Duration
	epochTree types.EpochTree
	ticker    *time.Ticker
}

// init 我们需要从数据库里，将之前那些已经commit了但是还没有justified的数据取出，作为当前重启后的epochTree的初始内容
func (em *EpochManager) init() error {
	em.epochTree.Clear()
	return helper.GlobalServiceHelper.DB.InitEpochTree(&em.epochTree)
}
func (em *EpochManager) advance(ctx context.Context) {
	for {
		select {
		case <-em.ticker.C:
			// Advance to next epoch
			// 这里我们首先将epochTree里的内容取出，group后传递给grpcEngine
			// 然后清空epochTree，可以继续处理
			// 这里数据库里的commit的内容可能不会完全被同步到内存，但是数据库本身保证了ACID
			// 我们允许一部分数据没有在最及时的时间被justified，如果机器没有崩溃，内存里的channel最后一定会取到
			// 如果崩溃了，重启的时候initEpochTree会取出那些数据
			em.ticker.Stop() // 先暂停一下
			// 这里select会阻塞下面的更新
			evidences := em.epochTree.Evidences()
			// 将evidences发送给grpc，并等待回复
			failedNodeIDs := helper.GlobalServiceHelper.SendIntegrityEvidence(evidences)
			for len(failedNodeIDs) != 0 {
				// 当不为0的时候我们要一直推进
				for _, nodeID := range failedNodeIDs {
					em.epochTree.Prune(nodeID)
				}
				evidences = em.epochTree.Evidences()
				failedNodeIDs = helper.GlobalServiceHelper.SendIntegrityEvidence(evidences)
			}
			// 此时没有错误了
			// 获取所有通过justified的slot
			if len(evidences) != 0 {
				slots := make([]string, 0)
				for _, e := range evidences {
					slots = append(slots, e.Slots()...)
				}
				err := helper.GlobalServiceHelper.DB.JustifiedSlot(slots) // 这里会更新任务进度
				if err != nil {
					panic(err) // we panic, something should recover
				}
			}
			em.epochTree.Clear() // 清空
			err := helper.GlobalServiceHelper.DB.AdvanceEpoch()
			if err != nil {
				helper.GlobalServiceHelper.ReportError(err)
				panic(err) // we panic, something should recover
			}
			em.ticker.Reset(em.d)
		case req := <-helper.GlobalServiceHelper.EpochUpdateQueue:
			// 如果有新的数据
			err := em.epochTree.Update(req)
			if err != nil {
				helper.GlobalServiceHelper.ReportError(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (em *EpochManager) Start(ctx context.Context) {
	defer em.ticker.Stop()
	err := em.init()
	if err != nil {
		panic(err)
	}
	go em.advance(ctx)

}

func NewEpochManager(config config.ComponentConfig) *EpochManager {
	return &EpochManager{
		d:      config.EpochTimeDuration,
		ticker: time.NewTicker(config.EpochTimeDuration),
	}
}
