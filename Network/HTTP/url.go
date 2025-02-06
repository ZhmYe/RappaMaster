package HTTP

import (
	"BHLayer2Node/LogWriter"
	"BHLayer2Node/paradigm"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HttpService struct {
	Url     string
	Method  string
	Handler func(c *gin.Context)
}

//func (s *HttpService) HandleRequest(request paradigm.HttpRequestInterface) paradigm.HttpResponse {
//	return s.Handler(request)
//}

type HttpServiceEnum int

const (
	INIT_TASK = iota
	ORACLE_QUERY
)

func (e *HttpEngine) SupportUrl() []HttpServiceEnum {
	return []HttpServiceEnum{INIT_TASK}
}
func (e *HttpEngine) GetHttpService(service HttpServiceEnum) (*HttpService, error) {
	switch service {
	case INIT_TASK:
		// 初始化任务
		httpService := HttpService{
			Url:    "/create",
			Method: "POST",
			Handler: func(c *gin.Context) {
				var requestBody paradigm.HttpInitTaskRequest

				// 解析请求体中的 JSON 数据
				if err := c.ShouldBindJSON(&requestBody); err != nil {
					// 如果解析失败，返回错误信息
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "Invalid JSON data",
						Code:    "ERROR",
						Data:    nil,
					})
					return
				}

				e.taskIDConsumer <- 1        // 获取ID
				taskID := <-e.taskIDProvider // 这里要阻塞
				task := paradigm.NewTask(
					taskID,
					paradigm.NameToModelType(requestBody.Model),
					requestBody.Params,
					requestBody.Size,
					requestBody.IsReliable,
				)
				LogWriter.Log("HTTP", fmt.Sprintf("Receive Init Task Request: %v, Generate New Task: %s", requestBody, task.Sign))
				// 新建任务用于上链，然后直接返回response
				//e.channel.PendingTransactions <- &paradigm.InitTaskTransaction{Task: task}
				//// 返回response
				//response := paradigm.HttpResponse{
				//	Message: fmt.Sprintf("Create New SynthTask Successfully, taskID: %s", task.Sign),
				//	Code:    "OK",
				//	Data:    nil,
				//}
				//c.JSON(http.StatusOK, response)

			},
		}
		return &httpService, nil
	case ORACLE_QUERY:
		// TODO
		return nil, nil
	default:
		paradigm.RaiseError(paradigm.NetworkError, "Unknown HTTP Service", false)
		//LogWriter.Log("ERROR", fmt.Sprintf("%s: %s", paradigm.ErrorToString(paradigm.NetworkError), "Unknown Http Service"))
		return nil, fmt.Errorf("unknown Http Service")
	}
}
