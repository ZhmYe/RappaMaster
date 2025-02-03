package HTTP

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"
)

type FakeHttpTaskRequest = paradigm.UnprocessedTask

// FakeHttpEngine 定义模拟的 HTTP 引擎
type FakeHttpEngine struct {
	//PendingRequestPool chan HttpTaskRequest          // 给 Scheduler 的请求池，接收来自前端的数据
	//initTasks          chan paradigm.UnprocessedTask // 给taskManager用于初始化任务的
	//fakeCollectChannel chan [2]interface{}
	//slotCollectChannel chan paradigm.CollectRequest
	channel *paradigm.RappaChannel
	config  Config.BHLayer2NodeConfig
	ip      string // IP 地址
	port    int    // 端口
}

// HandleRequest 处理请求
// 模拟从外部收到请求并将其推送到 pendingRequestPool
func (e *FakeHttpEngine) HandleRequest() {
	//for {
	// 模拟生成一个 HTTP 请求
	request := e.generateFakeRequest()

	// 将请求推送到请求池中
	//e.PendingRequestPool <- request
	e.channel.InitTasks <- request

	// 模拟请求间隔
	//time.Sleep(10 * time.Second)
	//}
}
func (e *FakeHttpEngine) HandleCollect() {
	// 当收到一个task完成的消息后，生成两个不同的collect请求
	idx := 0
	for fakeSign := range e.channel.FakeCollectSignChannel {
		sign := fakeSign[0].(string)
		size := fakeSign[1].(int32)
		generate_mission := func(sign string, size int32) string {
			return fmt.Sprintf("%s_%d_%d", sign, size, idx)
		}
		mission := generate_mission(sign, size)
		request := e.generateFakeCollectRequest(sign, size, mission)
		go func(collectChannel chan interface{}) {
			e.channel.ToCollectorRequestChannel <- request
			result := make([]interface{}, 0)
			for slotRecoverData := range collectChannel {
				result = append(result, slotRecoverData)
			}
			// 等待channel关闭
			LogWriter.Log("DEBUG", fmt.Sprintf("Mission %s Collect Finish, Size: %d", mission, size))
			fmt.Println(result, len(result))
		}(request.TransferChannel)
	}
}

// Start 启动 HTTP 引擎
func (e *FakeHttpEngine) Start() {
	address := fmt.Sprintf("%s:%d", e.ip, e.port) // 格式化 HTTP 地址
	LogWriter.Log("INFO", fmt.Sprintf("FakeHttpEngine Starting at %s...", address))

	// 启动请求处理 Goroutine
	go e.HandleRequest()
	go e.HandleCollect()
}

// generateFakeRequest 模拟生成 HTTP 请求格式化后的结果
func (e *FakeHttpEngine) generateFakeRequest() FakeHttpTaskRequest {
	// 模拟生成的请求
	request := FakeHttpTaskRequest{
		Sign:  fmt.Sprintf("FakeSign-%d", time.Now().Unix()),
		Size:  40, // 模拟固定大小
		Model: paradigm.CTGAN,
		Params: map[string]interface{}{
			"condition_column": "native-country",
			"condition_value":  "United-States",
		},
	}
	LogWriter.Log("DEBUG", fmt.Sprintf("Generated Fake HTTP Request: %+v", request))
	return request
}
func (e *FakeHttpEngine) generateFakeCollectRequest(sign string, size int32, mission string) paradigm.CollectRequest {
	request := paradigm.CollectRequest{
		Sign:            sign,
		Mission:         mission,
		Size:            size,
		TransferChannel: make(chan interface{}),
	}
	LogWriter.Log("DEBUG", fmt.Sprintf("Generate Fake Collect Request: %+v", request))
	return request
}

// Setup 配置 HTTP 引擎
func (e *FakeHttpEngine) Setup(config Config.BHLayer2NodeConfig) {
	e.config = config
	e.port = config.HttpPort
	e.ip = "127.0.0.1" // 默认绑定到本地地址
	// 初始化请求池
	//e.PendingRequestPool = make(chan UnprocessedTask, config.MaxHttpRequestPoolSize)
}

// NewFakeHttpEngine 创建并返回一个新的 FakeHttpEngine 实例
func NewFakeHttpEngine(channel *paradigm.RappaChannel) *FakeHttpEngine {
	return &FakeHttpEngine{
		//initTasks:          channel.InitTasks,
		//fakeCollectChannel: channel.FakeCollectSignChannel,
		//slotCollectChannel: channel.ToCollectorRequestChannel,
		channel: channel,
		//PendingRequestPool: PendingRequestPool,
	}
}
