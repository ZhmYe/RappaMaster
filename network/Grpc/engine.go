package Grpc

import (
	"RappaMaster/config"
	"RappaMaster/helper"
	pb "RappaMaster/pb/service"
	"RappaMaster/types"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sync"
)

type GrpcEngine struct {
	config.GrpcConfig
	nodeAddress []types.NodeGrpcAddress
	connPool    []*grpc.ClientConn
	mu          sync.RWMutex
	pb.UnimplementedRappaMasterServer
}

func (ge *GrpcEngine) Start(ctx context.Context) {
	for i := 0; i < len(ge.nodeAddress); i++ {
		if err := ge.connect(i); err != nil {
			helper.GlobalServiceHelper.ReportError(err)
		}
	}
	processSchedule := func() {
		for slot := range helper.GlobalServiceHelper.SlotSchedule {
			ge.sendSchedule(slot)
		}
	}
	processRappaMasterServer := func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", ge.Port))
		if err != nil {
			panic(fmt.Errorf("failed to listen: %v", err))
		}
		server := grpc.NewServer(grpc.MaxSendMsgSize(ge.MessageLimitSize), grpc.MaxRecvMsgSize(ge.MessageLimitSize))
		pb.RegisterRappaMasterServer(server, ge)
		types.Print("INFO", fmt.Sprintf("Coordinator gRPC server is running on :%d", ge.Port))
		if err := server.Serve(lis); err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}
	}
	go processSchedule()

	go processRappaMasterServer()
	for {
		select {
		case <-ctx.Done():
			ge.CloseAll()
			return
		}
	}
}

func (ge *GrpcEngine) connect(nodeId int) error {
	ge.mu.RUnlock() // 释放读锁
	ge.mu.Lock()
	defer ge.mu.Unlock()
	newConn, err := grpc.NewClient(ge.nodeAddress[nodeId].String(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(ge.MessageLimitSize), // 设置最大接收消息为 1gb
		grpc.MaxCallSendMsgSize(ge.MessageLimitSize), // 设置最大发送消息为 1gb
	))
	if err != nil {
		return types.RaiseError(types.NetworkError, fmt.Sprintf("failed to connect to Node{%d}", nodeId), err)
	}
	ge.connPool[nodeId] = newConn
	return nil
}

func (ge *GrpcEngine) GetConnection(nodeId int) (*grpc.ClientConn, error) {
	if nodeId >= len(ge.nodeAddress) {
		return nil, types.RaiseError(types.RuntimeError, fmt.Sprintf("nodeId %d out of range", nodeId), fmt.Errorf("%d >= %d", nodeId, len(ge.nodeAddress)))
	}
	ge.mu.RLock()
	if conn := ge.connPool[nodeId]; conn != nil {
		if conn.GetState() == connectivity.Ready {
			ge.mu.RUnlock()
			return conn, nil
		} else {
			ge.mu.RUnlock()
			ge.mu.Lock()
			conn.Close()
			ge.connPool[nodeId] = nil
			ge.mu.Unlock()
		}
	}
	if err := ge.connect(nodeId); err != nil {
		return nil, err
	}
	return ge.connPool[nodeId], nil

}

func (ge *GrpcEngine) CloseAll() {
	for _, conn := range ge.connPool {
		if conn == nil {
			continue
		}
		conn.Close()
	}
}

func NewGrpcEngine(config config.GrpcConfig) (*GrpcEngine, error) {
	ge := GrpcEngine{
		GrpcConfig: config,
	}
	return &ge, nil
}

func StartAll(ctx context.Context) {
	grpcEngine, err := NewGrpcEngine(config.GlobalSystemConfig.GrpcConfig)
	if err != nil {
		panic(err)
	}
	grpcEngine.Start(ctx)
}
