package test

import (
	"BHLayer2node/Config"
	"BHLayer2node/Network/HTTP"
	"testing"
	"time"
)

func TestFakeHttpEngine(t *testing.T) {
	// 初始化配置
	config := Config.LoadBHLayer2NodeConfig("")
	pool := make(chan HTTP.HttpTaskRequest, config.MaxHttpRequestPoolSize)
	// 创建并设置 FakeHttpEngine
	httpEngine := HTTP.NewFakeHttpEngine(pool)
	httpEngine.Setup(*config)

	// 启动 FakeHttpEngine
	httpEngine.Start()

	// 模拟请求的接收过程
	timeout := time.After(20 * time.Second) // 总测试时间
	tick := time.Tick(2 * time.Second)      // 检查间隔
	receivedRequests := 0                   // 记录收到的请求数量

	for {
		select {
		case <-timeout:
			// 超时退出测试
			t.Logf("Test completed. Total received requests: %d", receivedRequests)
			if receivedRequests == 0 {
				t.Error("No requests were received in the test duration")
			}
			return
		case <-tick:
			// 检查是否有请求进入请求池
			select {
			case request := <-httpEngine.PendingRequestPool:
				receivedRequests++
				t.Logf("Received request: %+v", request)
			default:
				t.Log("No requests received in this interval")
			}
		}
	}
}
