package HTTP

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// HttpEngine 定义模拟的 HTTP 引擎,这里引入gin框架提高编码效率
type HttpEngine struct {
	channel        *paradigm.RappaChannel
	taskIDConsumer chan int // 这里暂时用这个方法获取TaskID
	taskIDProvider chan paradigm.TaskHash
	config         paradigm.BHLayer2NodeConfig
	ip             string // IP 地址
	port           int    // 端口
	// 服务器
	r *gin.Engine
}

// AccumulateTaskID 不断累加即可
func (e *HttpEngine) AccumulateTaskID() {
	taskID := 0
	// 这里是一个额外的协程，在获取ID的时候阻塞
	for {
		<-e.taskIDConsumer
		e.taskIDProvider <- fmt.Sprintf("SynthTask-%d-%d", taskID, time.Now().Unix())
		taskID++
	}

}

// HandleRequest 处理请求
// 模拟从外部收到请求并将其推送到 pendingRequestPool
//func (e *HttpEngine) HandleRequest(c *gin.Context) {
//	var requestBody HttpTaskRequest
//
//	// 解析请求体中的 JSON 数据
//	if err := c.ShouldBindJSON(&requestBody); err != nil {
//		// 如果解析失败，返回错误信息
//		c.JSON(http.StatusBadRequest, HttpResponse{
//			Message: "Invalid JSON data",
//			Code:    "error",
//			Data:    nil,
//		})
//		return
//	}
//
//	task := paradigm.UnprocessedTask{
//		Sign:   requestBody.Sign,
//		Slot:   requestBody.Slot,
//		Size:   requestBody.Size,
//		Model:  paradigm.NameToModelType(requestBody.Model),
//		Params: requestBody.Params,
//	}
//
//	e.channel.InitTasks <- task
//
//	// 构造响应体
//	response := HttpResponse{
//		Message: "Received JSON data successfully",
//		Code:    "success",
//		Data:    requestBody, // 直接将请求体作为数据返回
//	}
//
//	// 返回 JSON 响应
//	c.JSON(http.StatusOK, response)
//}

func (e *HttpEngine) Start() {
	go e.AccumulateTaskID()
	paradigm.Print("INFO", fmt.Sprintf("Http server run on port %s:%d", e.ip, e.port))
	err := e.r.Run(fmt.Sprintf(":%d", e.port))
	if err != nil {
		paradigm.Error(paradigm.NetworkError, "Faild to start http engine because of"+err.Error())
	}
}

// Setup 配置 HTTP 引擎
func (e *HttpEngine) Setup(config paradigm.BHLayer2NodeConfig) {
	e.config = config
	e.port = config.HttpPort
	e.ip = "127.0.0.1" // 默认绑定到本地地址
	if e.config.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		// 设置 Gin 为 release 模式
		gin.SetMode(gin.ReleaseMode)
	}
	e.r = gin.Default()

	// 注册url
	//e.r.POST("/task", e.HandleRequest)
	for _, s := range e.SupportUrl() {
		service, err := e.GetHttpService(s)
		if err != nil {
			paradigm.Error(paradigm.RuntimeError, "url service not impl")
			continue
		}
		if service.Method == "POST" {
			e.r.POST(service.Url, service.Handler)
		} else if service.Method == "GET" {
			e.r.GET(service.Url, service.Handler)
		} else {
			// TODO
		}
	}
}

// NewHttpEngine 创建并返回一个新的 HttpEngine 实例
func NewHttpEngine(channel *paradigm.RappaChannel) *HttpEngine {
	http := HttpEngine{
		//initTasks:          channel.InitTasks,
		//fakeCollectChannel: channel.FakeCollectSignChannel,
		//slotCollectChannel: channel.ToCollectorRequestChannel,
		channel:        channel,
		taskIDProvider: make(chan paradigm.TaskHash, 100), // TODO @SD 加到config
		taskIDConsumer: make(chan int, 100),               // 这个数字和上面统一 TODO
		//PendingRequestPool: PendingRequestPool,
	}
	http.Setup(*channel.Config)
	return &http
}
