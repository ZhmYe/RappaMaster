package Grpc

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"sync"
)

// NodeGrpcManager gRPC 连接管理
type NodeGrpcManager struct {
	nodeAddresses map[int]*paradigm.BHNodeAddress // 节点地址映射，节点ID -> 地址
	connPool      map[int]*grpc.ClientConn        // 按nodeId分组的连接池
	mu            sync.RWMutex                    // 使用读写锁保护 connPool
}

func NewNodeGrpcManager(nodeAddresses map[int]*paradigm.BHNodeAddress) *NodeGrpcManager {
	p := &NodeGrpcManager{
		connPool:      make(map[int]*grpc.ClientConn),
		nodeAddresses: nodeAddresses,
	}
	return p
}

// GetConn 获取一个 gRPC 连接
func (p *NodeGrpcManager) GetConn(nodeId int) (*grpc.ClientConn, error) {
	// 如果连接池中有可用的连接，直接返回
	p.mu.RLock()
	if conn, ok := p.connPool[nodeId]; ok {
		// 检查连接是否健康
		if conn.GetState() == connectivity.Ready {
			p.mu.RUnlock() // 连接正常，释放读锁
			return conn, nil
		} else {
			//关闭和删除连接
			p.mu.RUnlock() // 释放读锁
			p.mu.Lock()    // 获取写锁
			conn.Close()
			delete(p.connPool, nodeId)
			p.mu.Unlock() // 释放写锁
			return nil, fmt.Errorf("the connection is bad,retry")
		}
	} else {
		p.mu.RUnlock() // 释放读锁
		p.mu.Lock()
		defer p.mu.Unlock()
		newConn, err := grpc.Dial(p.nodeAddresses[nodeId].GetAddrStr(), grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Node{%d}: %v", nodeId, err)
		}
		p.connPool[nodeId] = newConn
		return newConn, nil
	}
}

// GetNodeAddresses 获取节点的地址列表
func (p *NodeGrpcManager) GetNodeAddresses() map[int]*paradigm.BHNodeAddress {
	return p.nodeAddresses
}

// CloseAll 关闭连接池中的所有连接
func (p *NodeGrpcManager) CloseAll() {
	for nodeId, conn := range p.connPool {
		conn.Close()
		paradigm.Log("GrpcPort", fmt.Sprintf("Closed all connections for %d\n", nodeId))
	}
}
