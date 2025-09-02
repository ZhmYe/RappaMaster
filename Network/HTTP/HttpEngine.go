package HTTP

import (
	"RappaMaster/channel"
	"RappaMaster/config"
	"RappaMaster/paradigm"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// HttpEngine uses gin framework to process http requests
type HttpEngine struct {
	config.HTTPConfig
	channel channel.RappaChannel
	// 服务器
	r *gin.Engine
}

func (e *HttpEngine) Start() {
	paradigm.Print("INFO", fmt.Sprintf("Http server run on port %s:%d", e.IP, e.Port))
	err := e.r.Run(fmt.Sprintf(":%d", e.Port))
	if err != nil {
		panic(paradigm.RaiseError(paradigm.NetworkError, "Faild to start http engine", err))
	}
}

// Setup 配置 HTTP 引擎
func (e *HttpEngine) Setup() error {
	gin.SetMode(gin.ReleaseMode)
	e.r = gin.Default()
	e.r.Use(cors.Default())
	for _, s := range e.SupportUrl() {
		service, err := e.GetHttpService(s)
		if err != nil {
			return paradigm.RaiseError(paradigm.RuntimeError, "url service not impl", err)
		}
		if service.Method == "POST" {
			e.r.POST(service.Url, service.Handler)
		} else if service.Method == "GET" {
			e.r.GET(service.Url, service.Handler)
		} else {
			// TODO
		}
	}
	return nil
}

// NewHttpEngine 创建并返回一个新的 HttpEngine 实例
func NewHttpEngine(channel channel.RappaChannel, config config.HTTPConfig) (*HttpEngine, error) {
	http := HttpEngine{
		channel:    channel,
		HTTPConfig: config,
	}
	if err := http.Setup(); err != nil {
		return nil, err
	}
	return &http, nil
}
