package HTTP

import (
	"RappaMaster/config"
	"RappaMaster/paradigm"
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// HttpEngine uses gin framework to process http requests
type HttpEngine struct {
	config.HTTPConfig
	// 服务器
	r *gin.Engine
}

func (e *HttpEngine) Start(ctx context.Context) {
	//paradigm.Print("INFO", fmt.Sprintf("Http server run on port %s:%d", e.IP, e.Port))
	fmt.Println(fmt.Sprintf("Http server run on port %s:%d", e.IP, e.Port))
	go func() {
		err := e.r.Run(fmt.Sprintf(":%d", e.Port))
		if err != nil {
			panic(paradigm.RaiseError(paradigm.NetworkError, "Faild to start http engine", err))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

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

func NewHttpEngine(config config.HTTPConfig) (*HttpEngine, error) {
	http := HttpEngine{
		HTTPConfig: config,
	}
	if err := http.Setup(); err != nil {
		return nil, err
	}
	return &http, nil
}

func StartAll(ctx context.Context) {
	httpEngine, err := NewHttpEngine(config.GlobalSystemConfig.HttpConfig)
	if err != nil {
		panic(err)
	}
	httpEngine.Start(ctx)
}
