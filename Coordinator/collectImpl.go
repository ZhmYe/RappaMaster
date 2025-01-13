package Coordinator

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"sync"
)

func (c *Coordinator) sendCollect(request paradigm.RecoverConnection) {
	responseChannel := request.ResponseChannel        // 这里用于返回给collect Instance
	nodeAddresses := c.connManager.GetNodeAddresses() // 所有节点
	var wg sync.WaitGroup
	wg.Add(len(nodeAddresses))
	recoverRequest := pb.RecoverRequest{Mission: request.Mission, Hashs: request.Hashs}
	for nodeID, address := range nodeAddresses {
		// 并行处理各个节点
		go func(nodeID int, address string, request *pb.RecoverRequest) {
			defer wg.Done() // 减少 WaitGroup 计数器

			// 建立grpc连接
			conn, err := c.connManager.GetConn(nodeID)
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to connect to node %d at %s: %v", nodeID, address, err))
				//wg.Done()
				return
			}
			client := pb.NewRappaExecutorClient(conn)
			ctx := context.WithoutCancel(context.Background())
			//defer cancel()

			// 发送调度请求
			resp, err := client.Collect(ctx, request, grpc.WaitForReady(true))
			if err != nil {
				LogWriter.Log("ERROR", fmt.Sprintf("Failed to send collect request to node %d: %v", nodeID, err))
				//rejectedChannel <- 0 // 默认统计为未接受
				//wg.Done()
				return
			}
			for _, chunk := range resp.Chunks {
				fmt.Println(chunk.Row, chunk.Col)
			}
			responseChannel <- *resp // 将节点的resp返回给collect instance

		}(nodeID, address.GetAddrStr(), &recoverRequest)
	}
	wg.Wait()
	// 发送完了所有Recover Request
	close(responseChannel) // 此时collect instance开始恢复

}
