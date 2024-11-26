package Grpc

//
//import (
//	"BHLayer2Node/Config"
//	"BHLayer2Node/LogWriter"
//	"BHLayer2Node/paradigm"
//	"fmt"
//	"time"
//)
//
//type FakeGrpcEngine struct {
//	config Config.BHLayer2NodeConfig
//	port   int
//	ip     string
//	//connectedNodes    map[int]bool                  // 模拟连接的节点，key 为节点 ID
//	//connectionIndexes []int                         // 模拟节点索引
//	PendingSlotPool   chan paradigm.PendingSlotItem // 来自scheduler
//	PendingSlotRecord chan paradigm.SlotRecord      // 发给scheduler
//}
//
//// SendSchedule 模拟发送调度信息到节点
//func (e *FakeGrpcEngine) SendSchedule() {
//	//LogWriter.Log("GRPC", fmt.Sprintf("Sending schedule to connected nodes: %s", schedule))
//	//// 模拟发送的耗时
//	//time.Sleep(500 * time.Millisecond)
//	//LogWriter.Log("GRPC", "Schedule sent successfully")
//}
//func (e *FakeGrpcEngine) Start() {
//	LogWriter.Log("DEBUG", fmt.Sprintf("FakeGrpcEngine start at %s:%d", e.ip, e.port))
//	processPendingSlotPool := func() {
//		for slot := range e.PendingSlotPool {
//			// 这里是收到了scheduler的调度方案，要转发给合成节点
//			totalSize := 0
//			active := make([]int, 0)
//			for _, s := range slot.Schedule {
//				totalSize += s.Size
//				active = append(active, s.NID)
//				LogWriter.Log("DEBUG", fmt.Sprintf("sending Task %s Slot %d Schedule to node %d, size: %d", slot.Sign, slot.Slot, s.NID, s.Size))
//			}
//			go func(size int, active []int, sign string, slot int) {
//				record := paradigm.SlotRecord{
//					Size:   size,
//					Sign:   sign,
//					Slot:   slot,
//					Miss:   make([]int, 0),
//					Active: active,
//				}
//				time.Sleep(2 * time.Second)
//				e.PendingSlotRecord <- record
//			}(totalSize, active, slot.Sign, slot.Slot)
//		}
//	}
//	go processPendingSlotPool()
//
//}
//
//// Setup 初始化 FakeGrpcEngine
//func (e *FakeGrpcEngine) Setup(config Config.BHLayer2NodeConfig) {
//	e.config = config
//	e.port = config.GrpcPort
//	e.ip = "127.0.0.1" // 默认绑定到本地地址
//	//e.connectedNodes = make(map[int]bool)
//	//e.connectionIndexes = []int{}
//
//	//LogWriter.Log("DEBUG", fmt.Sprintf("FakeGrpcEngine setup completed on %s:%d", e.ip, e.port))
//}
//
//// GetConnected 获取当前连接的节点数量
//func (e *FakeGrpcEngine) GetConnected() int {
//	return 5
//}
//
//// GetConnectIndex 获取所有连接的节点索引
//func (e *FakeGrpcEngine) GetConnectIndex() []int {
//	return []int{0, 1, 2, 3, 4}
//}
//
//// NewFakeGrpcEngine 创建一个新的 FakeGrpcEngine 实例
//func NewFakeGrpcEngine(pendingSlotPool chan paradigm.PendingSlotItem, pendingSlotRecord chan paradigm.SlotRecord) *FakeGrpcEngine {
//	return &FakeGrpcEngine{
//		PendingSlotPool:   pendingSlotPool,
//		PendingSlotRecord: pendingSlotRecord,
//		//connectedNodes:    make(map[int]bool),
//		//connectionIndexes: []int{},
//	}
//}
