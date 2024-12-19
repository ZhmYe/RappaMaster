package Grpc

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"time"
)

// gRPC 连接管理
type NodeGrpcManager struct {
	nodeAddresses map[int]*Config.BHNodeAddress // 节点地址映射，节点ID -> 地址
	connPool      map[int]*grpc.ClientConn      // 按nodeId分组的连接池
}

// 创建一个新的连接池
func NewNodeGrpcPool(nodeAddresses map[int]*Config.BHNodeAddress, maxConns int, connExpiry time.Duration) *NodeGrpcManager {
	p := &NodeGrpcManager{
		connPool:      make(map[int]*grpc.ClientConn),
		nodeAddresses: nodeAddresses,
	}
	return p
}

// 获取一个 gRPC 连接
func (p *NodeGrpcManager) GetConn(nodeId int) (*grpc.ClientConn, error) {
	// 如果连接池中有可用的连接，直接返回
	if conn, ok := p.connPool[nodeId]; ok {
		// 检查连接是否健康
		if conn.GetState() == connectivity.Ready {
			return conn, nil
		} else {
			//关闭和删除连接
			conn.Close()
			delete(p.connPool, nodeId)
			return nil, fmt.Errorf("the connection is bad,retry")
		}
	} else {
		newConn, err := grpc.Dial(p.nodeAddresses[nodeId].GetAddrStr(), grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Node{%d}: %v", nodeId, err)
		}
		p.connPool[nodeId] = newConn
		return newConn, nil
	}
}

// 获取节点的地址列表
func (p *NodeGrpcManager) GetNodeAddresses() map[int]*Config.BHNodeAddress {
	return p.nodeAddresses
}

// 关闭连接池中的所有连接
func (p *NodeGrpcManager) CloseAll() {
	for nodeId, conn := range p.connPool {
		conn.Close()
		LogWriter.Log("GrpcPort", fmt.Sprintf("Closed all connections for %d\n", nodeId))
	}
}
