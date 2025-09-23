package types

import (
	"fmt"
)

const (
	CPU_USAGE_THERESHOLD     int64 = 80
	DISK_AVAILABLE_THRESHOLD int64 = 10000000000 // 10GB
)

type NodeStatus struct {
	NodeID          int
	cpuUsage        int64 // we set a threshold, %, todo gpu?
	diskAvailable   int64 // we set a threshold, here usage is an absolute-number like 100000KB
	unprocessedTask int64 // KB, how many slots this node not finish
	speed           int64 // we compute a "speed" of node, we use Aging algorithm
	// todo 这里可以加一个恶意评分，就是在检测出某个节点提交的东西有问题的时候返回，增加这个评分，如果提交正常的slot，慢慢减这个评分,check里再加上一个信誉值
}

func (status *NodeStatus) Speed() int64 {
	return status.speed
}

// Check checks whether a node can undertake more tasks
func (status *NodeStatus) Check() bool {
	return status.cpuUsage < CPU_USAGE_THERESHOLD && status.diskAvailable > DISK_AVAILABLE_THRESHOLD
}

type NodeGrpcAddress struct {
	NodeIPAddress string
	NodeGrpcPort  int
}

func (address *NodeGrpcAddress) String() string {
	return fmt.Sprintf("%s:%d", address.NodeIPAddress, address.NodeGrpcPort)
}
