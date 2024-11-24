package Grpc

import "BHLayer2node/Config"

type GrpcInterface interface {
	GetConnected() int                      // 获取当前连接数，在获取心跳的时候记录
	GetConnectIndex() []int                 // 获取有哪些节点是可以访问到的
	SendSchedule()                          // 向合成节点发送任务
	Setup(config Config.BHLayer2NodeConfig) // 读取配置
	Start()
}
