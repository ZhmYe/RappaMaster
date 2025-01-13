package Coordinator

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/handler"
	pb "BHLayer2Node/pb/service"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"sync"
	"time"
)

func (c *Coordinator) sendHeartbeat(heartbeat *pb.HeartbeatRequest) {
	nodeAddresses := c.connManager.GetNodeAddresses()
	var wg sync.WaitGroup
	disconnected := make(chan int, len(nodeAddresses))                       // 用于说明节点失联
	response := make(chan *pb.HeartbeatResponse, len(nodeAddresses))         // 用于给voteHandler传递
	voteHandler := handler.NewVoteHandler(heartbeat, c.commitSlot, response) // 处理投票结果
	go voteHandler.Process()                                                 // 准备处理投票
	// 遍历所有节点
	for nodeID, address := range nodeAddresses {
		wg.Add(1) // 增加 WaitGroup 计数器
		go func(nodeID int, address string, heartbeat *pb.HeartbeatRequest) {
			defer wg.Done() // 减少 WaitGroup 计数器

			// 建立grpc连接
			conn, err := c.connManager.GetConn(nodeID)
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to connect to node %d at %s: %v", nodeID, address, err))
				disconnected <- nodeID // 暂定为当作失联
				return
			}
			client := pb.NewRappaExecutorClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 发送心跳
			resp, err := client.Heartbeat(ctx, heartbeat, grpc.WaitForReady(true))
			//TODO grpc在建立连接时其实就可以判断出节点是否失联
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to send heartbeat to node %d: %v", nodeID, err))
				//rejectedChannel <- 0 // 默认统计为未接受
				disconnected <- nodeID // 暂定为当作失联
				return
			}
			response <- resp // 交给voteHandler处理

		}(nodeID, address.GetAddrStr(), heartbeat)
	}

	// 等待所有节点处理完成
	wg.Wait()
	close(response)     // 关闭response，voteHandler开始处理投票结果
	close(disconnected) // 关闭disconnected, 后续monitor结束处理 todo

}
