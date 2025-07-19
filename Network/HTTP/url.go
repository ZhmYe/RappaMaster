package HTTP

import (
	"BHLayer2Node/Collector"
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
	COMMIT_PROOF
)

func (e *HttpEngine) SupportUrl() []HttpServiceEnum {
	return []HttpServiceEnum{INIT_TASK, ORACLE_QUERY, BLOCKCHAIN_QUERY, DATASYNTH_QUERY, COLLECT_TASK, COMMIT_PROOF}
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

func (e *HttpEngine) HandleCommitProof(c *gin.Context) {
	var req Query.CommitProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body: " + err.Error()})
		return
	}

	proofBytes, err := base64.StdEncoding.DecodeString(req.Proof)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid base64 proof data"})
		return
	}

	paradigm.Log("HTTP", fmt.Sprintf("Received proof for slot: %s", req.SlotHash))

	proofReceipt := paradigm.ProofReceipt{
		SlotHash: req.SlotHash,
		Proof:    proofBytes,
	}

	select {
	case e.channel.ProofReceivedChannel <- proofReceipt:
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Proof received and queued for processing."})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "Server is busy, please try again later."})
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
					requestBody.Name,
					paradigm.NameToModelType(requestBody.Model),
					requestBody.Params,
					requestBody.Size,
					requestBody.IsReliable,
				)
				task.SetCollector(Collector.NewCollector(task.Sign, task.OutputType, e.channel))
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
	case COMMIT_PROOF:
		httpService := HttpService{
			Url:     "/commit_proof",
			Method:  "POST",
			Handler: e.HandleCommitProof,
		}
		return &httpService, nil
	default:
		paradigm.Error(paradigm.NetworkError, "Unknown HTTP Service")
		//LogWriter.Log("ERROR", fmt.Sprintf("%s: %s", paradigm.ErrorToString(paradigm.NetworkError), "Unknown Http Service"))
		return nil, fmt.Errorf("unknown Http Service")
	}
}
