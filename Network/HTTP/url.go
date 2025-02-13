package HTTP

import (
	"BHLayer2Node/Query"
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
	COLLECT_TASK
	BLOCKCHAIN_QUERY
	DATASYNTH_QUERY
)

func (e *HttpEngine) SupportUrl() []HttpServiceEnum {
	return []HttpServiceEnum{INIT_TASK, ORACLE_QUERY, BLOCKCHAIN_QUERY, DATASYNTH_QUERY}
}
func (e *HttpEngine) HandleGET(c *gin.Context) {
	var requestBody Query.HttpOracleQueryRequest
	// 解析请求体中的 JSON 数据

	if success, query := requestBody.BuildQueryFromGETRequest(c); success {
		fmt.Println(query.ToHttpJson())
		e.channel.QueryChannel <- query
		r := query.ReceiveResponse() // 这里会阻塞
		//fmt.Println(r.ToHttpJson(), r.Error())
		response := paradigm.HttpResponse{
			Message: fmt.Sprintf("Query Data Successfully, query type: %s, query: %v", requestBody.Query, requestBody.Data),
			Code:    "OK",
			Data:    r.ToHttpJson(),
		}
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
			Message: "Invalid Request data",
			Code:    "ERROR",
			Data:    nil,
		})
	}
}
func (e *HttpEngine) GetHttpService(service HttpServiceEnum) (*HttpService, error) {
	switch service {
	case INIT_TASK:
		// 初始化任务
		httpService := HttpService{
			Url:    "/create",
			Method: "POST",
			Handler: func(c *gin.Context) {
				var requestBody Query.HttpInitTaskRequest

				// 解析请求体中的 JSON 数据
				if err := c.ShouldBindJSON(&requestBody); err != nil {
					// 如果解析失败，返回错误信息
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "Invalid Request data",
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
				paradigm.Log("HTTP", fmt.Sprintf("Receive Init Task Request: %v, Generate New Task: %s", requestBody, task.Sign))
				// 新建任务用于上链，然后直接返回response
				e.channel.PendingTransactions <- &paradigm.InitTaskTransaction{Task: task}
				// 返回response
				response := paradigm.HttpResponse{
					Message: fmt.Sprintf("Create New SynthTask Successfully, taskID: %s", task.Sign),
					Code:    "OK",
					Data:    nil,
				}
				c.JSON(http.StatusOK, response)

			},
		}
		return &httpService, nil
	case ORACLE_QUERY:
		httpService := HttpService{
			Url:     "/oracle",
			Method:  "GET",
			Handler: e.HandleGET,
		}
		return &httpService, nil
	case COLLECT_TASK:
		// TODO
		return nil, nil
	case BLOCKCHAIN_QUERY:
		httpService := HttpService{
			Url:     "/blockchain",
			Method:  "GET",
			Handler: e.HandleGET,
		}
		return &httpService, nil
	case DATASYNTH_QUERY:
		httpService := HttpService{
			Url:     "/dataSynth",
			Method:  "GET",
			Handler: e.HandleGET,
		}
		return &httpService, nil
	default:
		paradigm.Error(paradigm.NetworkError, "Unknown HTTP Service")
		//LogWriter.Log("ERROR", fmt.Sprintf("%s: %s", paradigm.ErrorToString(paradigm.NetworkError), "Unknown Http Service"))
		return nil, fmt.Errorf("unknown Http Service")
	}
}
