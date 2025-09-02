package HTTP

import (
	"RappaMaster/Query"
	"RappaMaster/paradigm"
	"RappaMaster/task"
	"RappaMaster/transaction"
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpService struct {
	Url     string
	Method  string
	Handler func(c *gin.Context)
}

type HttpServiceEnum int

const (
	INIT_TASK = iota
	ORACLE_QUERY
	COLLECT_TASK
	BLOCKCHAIN_QUERY
	DATASYNTH_QUERY
)

func (e *HttpEngine) SupportUrl() []HttpServiceEnum {
	return []HttpServiceEnum{INIT_TASK, ORACLE_QUERY, BLOCKCHAIN_QUERY, DATASYNTH_QUERY, COLLECT_TASK}
}
func (e *HttpEngine) HandleGET(c *gin.Context) {
	var requestBody Query.HttpOracleQueryRequest
	// 解析请求体中的 JSON 数据

	if success, query := requestBody.BuildQueryFromGETRequest(c); success {
		//fmt.Println(query.ToHttpJson())
		e.channel.QueryChannel <- query
		r := query.ReceiveResponse() // 这里会阻塞

		response := paradigm.HttpResponse{
			Message: fmt.Sprintf("Query Data Successfully, query type: %s, query: %v", requestBody.Query, requestBody.Data),
			Code:    "OK",
			Data:    r.ToHttpJson(),
		}
		c.JSON(http.StatusOK, response)
		//fmt.Println(r.ToHttpJson(), r.Error())
	} else {
		c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
			Message: "Invalid Request data",
			Code:    "ERROR",
			Data:    nil,
		})
	}
}

func (e *HttpEngine) HandleDownload(c *gin.Context) {
	var requestBody Query.HttpOracleQueryRequest
	// 解析请求体中的 JSON 数据
	if success, query := requestBody.BuildQueryFromGETRequest(c); success {
		//fmt.Println(query.ToHttpJson())
		e.channel.QueryChannel <- query
		r := query.ReceiveResponse() // 这里会阻塞
		fileJson := r.ToHttpJson()
		data := fileJson["file"].([]byte)
		filename := fileJson["filename"].(string)
		// 将 byte 数组转换为 Reader
		reader := bytes.NewReader(data)
		// 设置响应头
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		// 流式传输数据
		c.DataFromReader(
			http.StatusOK,
			int64(len(data)), // 数据总大小（Content-Length）
			"application/octet-stream",
			reader,
			nil, // 可选的额外 headers
		)
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
				t := task.NewTask(requestBody.Name, paradigm.NameToModelType(requestBody.Model), requestBody.Size)
				paradigm.Log("HTTP", fmt.Sprintf("Receive Init Task Request: %v, Generate New Task: %s", requestBody, t.Sign))
				receipt, err := e.channel.SendWithSync(transaction.NewInitTaskTransaction([]task.Task{*t}))
				if err != nil {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: fmt.Sprintf("create task error: %v", err),
						Code:    "ERROR",
						Data:    nil,
					})
				}
				if err = e.channel.CreateTask(*t, receipt); err != nil {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: fmt.Sprintf("create task error: %v", err),
						Code:    "ERROR",
						Data:    nil,
					})
				}

				response := paradigm.HttpResponse{
					Message: fmt.Sprintf("Create New SynthTask Successfully, task sign: %s", t.Sign()),
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
		httpService := HttpService{
			Url:     "/collect",
			Method:  "GET",
			Handler: e.HandleDownload,
		}
		return &httpService, nil
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
