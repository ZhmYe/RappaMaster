package handler

// HeartBeat Master的心跳
// 一个epoch发送一次，包含
// 1. 这个epoch中commit的Task Slot(sign, slot, nid) 这里暂时假定节点就是简单的查看自己本地有没有，先不管其它可能的错误
// 2. 上一个epoch中finalize的Task Slot
// 3. 这一epoch中未完成且过期的任务的新slot(sign, slot)

//type HeartBeatHandler struct {
//	monitor     *Monitor.Monitor // 处理节点状态信息监控
//	voteHandler VoteHandler      // 处理投票
//
//}

//
//func (handler *HeartBeatHandler) Start() {
//	processHeartbeat := func() {
//		for {
//			select {
//			case heartbeat := <-handler.epochHeartbeat:
//				// 说明epoch更新需要发心跳给节点
//			}
//		}
//	}
//}
