package HTTP

import (
	"BHLayer2Node/Config"
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HttpResponse struct {
	Message string      `json:"msg"`    // 消息
	Code    string      `json:"status"` // 状态
	Data    interface{} `json:"data"`   // 数据
}

type HttpTaskRequest struct {
	Sign   string                 // task sign
	Slot   int32                  // slot index
	Size   int32                  // data size
	Model  string                 // 模型名称
	Params map[string]interface{} // 不确定的模型参数
}

// HttpEngine 定义模拟的 HTTP 引擎,这里引入gin框架提高编码效率
type HttpEngine struct {
	//PendingRequestPool chan HttpTaskRequest          // 给 Scheduler 的请求池，接收来自前端的数据
	//initTasks          chan paradigm.UnprocessedTask // 给taskManager用于初始化任务的
	//fakeCollectChannel chan [2]interface{}
	//slotCollectChannel chan paradigm.CollectRequest
	channel *paradigm.RappaChannel
	config  Config.BHLayer2NodeConfig
	ip      string // IP 地址
	port    int    // 端口
	// 服务器
	r *gin.Engine
}

// HandleRequest 处理请求
// 模拟从外部收到请求并将其推送到 pendingRequestPool
func (e *HttpEngine) HandleRequest(c *gin.Context) {
	var requestBody HttpTaskRequest

	// 解析请求体中的 JSON 数据
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		// 如果解析失败，返回错误信息
		c.JSON(http.StatusBadRequest, HttpResponse{
			Message: "Invalid JSON data",
			Code:    "error",
			Data:    nil,
		})
		return
	}

	task := paradigm.UnprocessedTask{
		Sign:   requestBody.Sign,
		Slot:   requestBody.Slot,
		Size:   requestBody.Size,
		Model:  paradigm.NameToModelType(requestBody.Model),
		Params: requestBody.Params,
	}

	e.channel.InitTasks <- task

	// 构造响应体
	response := HttpResponse{
		Message: "Received JSON data successfully",
		Code:    "success",
		Data:    requestBody, // 直接将请求体作为数据返回
	}

	// 返回 JSON 响应
	c.JSON(http.StatusOK, response)
}

func (e *HttpEngine) Start() {
	LogWriter.Log("INFO", fmt.Sprintf("Http server run on port %s:%d", e.ip, e.port))
	err := e.r.Run(fmt.Sprintf(":%d", e.port))
	if err != nil {
		LogWriter.Log("ERROR", "Faild to start http engine because of"+err.Error())
	}
}

// Setup 配置 HTTP 引擎
func (e *HttpEngine) Setup(config Config.BHLayer2NodeConfig) {
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
	e.r.POST("/task", e.HandleRequest)
}

// NewFakeHttpEngine 创建并返回一个新的 FakeHttpEngine 实例
func NewHttpEngine(channel *paradigm.RappaChannel) *HttpEngine {
	return &HttpEngine{
		//initTasks:          channel.InitTasks,
		//fakeCollectChannel: channel.FakeCollectSignChannel,
		//slotCollectChannel: channel.ToCollectorRequestChannel,
		channel: channel,
		//PendingRequestPool: PendingRequestPool,
	}
}
