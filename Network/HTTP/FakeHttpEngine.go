package HTTP

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"
)

type HttpTaskRequest = paradigm.UnprocessedTask

// FakeHttpEngine 定义模拟的 HTTP 引擎
type FakeHttpEngine struct {
	//PendingRequestPool chan HttpTaskRequest          // 给 Scheduler 的请求池，接收来自前端的数据
	initTasks chan paradigm.UnprocessedTask // 给taskManager用于初始化任务的
	config    Config.BHLayer2NodeConfig
	ip        string // IP 地址
	port      int    // 端口
}

// HandleRequest 处理请求
// 模拟从外部收到请求并将其推送到 pendingRequestPool
func (e *FakeHttpEngine) HandleRequest() {
	//for {
	// 模拟生成一个 HTTP 请求
	request := e.generateFakeRequest()

	// 将请求推送到请求池中
	//e.PendingRequestPool <- request
	e.initTasks <- request

	// 模拟请求间隔
	//time.Sleep(10 * time.Second)
	//}
}

// Start 启动 HTTP 引擎
func (e *FakeHttpEngine) Start() {
	address := fmt.Sprintf("%s:%d", e.ip, e.port) // 格式化 HTTP 地址
	LogWriter.Log("INFO", fmt.Sprintf("FakeHttpEngine Starting at %s...", address))

	// 启动请求处理 Goroutine
	go e.HandleRequest()
}

// generateFakeRequest 模拟生成 HTTP 请求格式化后的结果
func (e *FakeHttpEngine) generateFakeRequest() HttpTaskRequest {
	// 模拟生成的请求
	request := HttpTaskRequest{
		Sign:  fmt.Sprintf("FakeSign-%d", time.Now().Unix()),
		Size:  100, // 模拟固定大小
		Model: "FakeModel",
		Params: map[string]interface{}{
			"param1": "value1",
			"param2": "value2",
		},
	}
	LogWriter.Log("DEBUG", fmt.Sprintf("Generated Fake HTTP Request: %+v", request))
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
func NewFakeHttpEngine(initTasks chan paradigm.UnprocessedTask) *FakeHttpEngine {
	return &FakeHttpEngine{
		initTasks: initTasks,
		//PendingRequestPool: PendingRequestPool,
	}
}
