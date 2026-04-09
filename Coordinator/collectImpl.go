package Coordinator

import (
	"BHLayer2Node/paradigm"
	pb "BHLayer2Node/pb/service"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"sync"
	"time"
)

func (c *Coordinator) sendCollect(request paradigm.SlotRecoverConnection) {
	responseChannel := request.ResponseChannel // 这里用于返回给collect Instance
	// nodeAddresses := c.connManager.GetNodeAddresses() // 所有节点
	var wg sync.WaitGroup
	recoverRequest := pb.RecoverRequest{Mission: fmt.Sprintf("Mission_%s", time.Now().Format("2006-01-02_15-04-05")), Hashs: []string{string(request.Hash)}}

	sendToNode := func(nodeID int) {
		nodeAddresses := c.connManager.GetNodeAddresses()
		address, ok := nodeAddresses[nodeID]
		if !ok {
			paradigm.Error(paradigm.NetworkError, fmt.Sprintf("Node %d address not found", nodeID))
			return
		}
		addrStr := address.GetAddrStr()
		wg.Add(1)
		go func(nodeID int, address string, request *pb.RecoverRequest) {
			defer wg.Done() // 减少 WaitGroup 计数器

			// 建立grpc连接
			conn, err := c.connManager.GetConn(nodeID)
			if err != nil {
				paradigm.Error(paradigm.NetworkError, fmt.Sprintf("Failed to connect to node %d at %s: %v", nodeID, address, err))
				return
			}
			client := pb.NewRappaExecutorClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// 发送调度请求
			resp, err := client.Collect(ctx, request, grpc.WaitForReady(true))
			if err != nil {
				paradigm.Error(paradigm.NetworkError, fmt.Sprintf("Failed to send collect request to node %d: %v", nodeID, err))
				return
			}
			responseChannel <- resp // 将节点的resp返回给collect instance

		}(nodeID, addrStr, &recoverRequest)
	}

	if request.NodeId != -1 {
		// 指定了节点，直接发送
		sendToNode(request.NodeId)
	} else {
		// 没有指定节点，广播
		nodeAddresses := c.connManager.GetNodeAddresses()
		for nodeID := range nodeAddresses {
			sendToNode(nodeID)
		}
	}

	wg.Wait()
	// 注意：不能在这里close(responseChannel)，因为channel是外部传入且可能共用的
}
