package Mocker

import (
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

// MockerExecutionNode 模拟执行节点
type MockerExecutionNode struct {
	nodeID   int
	ip       string
	slotData map[string]string // 模拟当前节点的任务分配情况
	service.UnimplementedCoordinatorServer
	mu         sync.Mutex                   // 保护 slotData 的并发访问
	commitSlot chan paradigm.CommitSlotItem // 这里模拟直接commit
}

// NewMockerExecutionNode 创建新的节点
func NewMockerExecutionNode(nodeID int, ip string, commitSlot chan paradigm.CommitSlotItem) *MockerExecutionNode {
	return &MockerExecutionNode{
		nodeID:     nodeID,
		ip:         ip,
		slotData:   make(map[string]string),
		commitSlot: commitSlot,
	}
}

func (m *MockerExecutionNode) Start() {
	server := grpc.NewServer()
	service.RegisterCoordinatorServer(server, m)

	listener, err := net.Listen("tcp", m.ip) // 监听端口
	if err != nil {

		panic(fmt.Errorf("failed to listen on port 50052: %v", err))
	}

	paradigm.Log("DEBUG", fmt.Sprintf("MockerExecutionNode is listening on port %s", m.ip))
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
