package Grpc

import (
	"RappaMaster/helper"
	pb "RappaMaster/pb/service"
	"context"
)

func (ge *GrpcEngine) CommitSlot(ctx context.Context, req *pb.SlotCommitRequest) (*pb.SlotCommitResponse, error) {
	err := helper.GlobalServiceHelper.DB.CommitSlot(req)
	if err != nil {
		return &pb.SlotCommitResponse{Accept: false}, err
	}
	helper.GlobalServiceHelper.UpdateEpochTree(req) // 非阻塞，这里只要数据库成功写入了就行
	return &pb.SlotCommitResponse{Accept: true}, nil
}
