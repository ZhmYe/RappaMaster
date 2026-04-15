package HTTP

import (
	"BHLayer2Node/Collector"
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"fmt"
	"io"
	"net/http"
	"time"

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
	UPLOAD_TASK
	BLOCKCHAIN_QUERY
	DATASYNTH_QUERY
	EXECUTION_LOG
	CREATE_SIM_TASK
	ANALYZED_STOCKS
	ABM_PARAMETERS
	ORDER_DYNAMICS
	PRICE_SYNTH_DOWNLOAD
	PRICE_SYNTH
	CRASH_RISK
	INVESTOR_COMP
	PERF_COMPARISON
)

func (e *HttpEngine) SupportUrl() []HttpServiceEnum {
	return []HttpServiceEnum{INIT_TASK, ORACLE_QUERY, BLOCKCHAIN_QUERY, DATASYNTH_QUERY, COLLECT_TASK, EXECUTION_LOG, CREATE_SIM_TASK, ANALYZED_STOCKS, ABM_PARAMETERS, ORDER_DYNAMICS, PRICE_SYNTH_DOWNLOAD, PRICE_SYNTH, CRASH_RISK, INVESTOR_COMP, PERF_COMPARISON}
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

// HandleSimulationDownload 复用已有下载功能的逻辑
func (e *HttpEngine) HandleSimulationDownload(c *gin.Context) {
	taskId := c.Query("taskId")
	stockId := c.Query("stockId")
	if taskId == "" || stockId == "" {
		c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId and stockId required", Code: "E100005", Data: nil})
		return
	}

	// 映射到系统的 Sign (SubTask-taskId-stockId)
	sign := fmt.Sprintf("SubTask-%s-%s", taskId, stockId)

	// 获取任务信息来确定Size
	task, err := e.dbService.GetTaskByID(sign)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, paradigm.HttpResponse{Message: "任务未找到", Code: "E100006", Data: nil})
		return
	}

	// 构造 CollectTaskQuery 并直接调用 HandleDownload 的核心逻辑
	query := Query.NewCollectTaskQuery(map[interface{}]interface{}{
		"taskID": sign,
		"size":   int(task.Size),
	})

	e.channel.QueryChannel <- query
	r := query.ReceiveResponse() // 会阻塞
	fileJson := r.ToHttpJson()
	interfaceReader, ok := fileJson["fileReader"]
	if !ok || interfaceReader == nil {
		c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{Message: "结果文件尚未准备好", Code: "E100007", Data: nil})
		return
	}

	reader := interfaceReader.(io.Reader)
	filename := fileJson["filename"].(string)

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (e *HttpEngine) HandleDownload(c *gin.Context) {
	var requestBody Query.HttpOracleQueryRequest
	// 解析请求体中的 JSON 数据
	if success, query := requestBody.BuildQueryFromGETRequest(c); success {
		//fmt.Println(query.ToHttpJson())
		e.channel.QueryChannel <- query
		r := query.ReceiveResponse() // 这里会阻塞
		fileJson := r.ToHttpJson()
		reader := fileJson["fileReader"].(*io.PipeReader) // 这里改成流式reader
		filename := fileJson["filename"].(string)
		// 将 byte 数组转换为 Reader
		//reader := bytes.NewReader(data)
		// 设置响应头
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		// 流式传输数据
		c.DataFromReader(
			http.StatusOK,
			-1, // 内容长度未知，使用 -1
			"application/octet-stream",
			reader, // 流传输
			nil,    // 可选的额外 headers
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

				//如果没有指定，设置默认值
				if requestBody.SlotSize == 0 {
					requestBody.SlotSize = 3000
				}

				task := paradigm.NewTask(
					taskID,
					requestBody.Name,
					paradigm.NameToModelType(requestBody.Model),
					requestBody.SlotSize,
					requestBody.Params,
					requestBody.Size,
					requestBody.IsReliable,
				)
				task.SetCollector(Collector.NewCollector(task.Sign, task.OutputType, e.channel, e.pkiManager))
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
	case UPLOAD_TASK:
		httpService := HttpService{
			Url:     "/upload",
			Method:  "GET",
			Handler: e.HandleGET,
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
	case EXECUTION_LOG:
		httpService := HttpService{
			Url:    "/dashboard/execution_log",
			Method: "GET",
			Handler: func(c *gin.Context) {
				tasks, err := e.dbService.GetAllPlatformTasks()
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取执行日志失败",
						Code:    "E100001",
						Data:    nil,
					})
					return
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    tasks,
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case CREATE_SIM_TASK:
		httpService := HttpService{
			Url:    "/simulation/create-task",
			Method: "POST",
			Handler: func(c *gin.Context) {
				var rawTasks []map[string]interface{}
				if err := c.ShouldBindJSON(&rawTasks); err != nil {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "参数格式错误",
						Code:    "E100002",
						Data:    false,
					})
					return
				}

				isScheduled := c.Query("isScheduled") == "true"

				// 生成ID
				taskID, _ := e.dbService.GetNextPlatformTaskID()

				var subTasks []paradigm.Task
				for _, raw := range rawTasks {
					// 调度逻辑：为每个子任务分配一个节点
					nodeID := e.monitor.SelectLeastLoadedNode()

					// 复用 Task 结构
					task := paradigm.Task{
						Sign:      fmt.Sprintf("SubTask-%s-%s", taskID, raw["stockCode"]),
						Name:      fmt.Sprintf("%s推演", raw["stockName"]),
						Params:    make(map[string]interface{}),
						Status:    paradigm.Processing,
						StartTime: time.Now(),
					}

					// 将参数存到 params 字段里，保留横线
					for k, v := range raw {
						task.Params[k] = v
					}

					// 记录分配的节点 (在 Schedule 层级处理，或者临时记在某处)
					// 这里先简单记录到 Params 里或者 Task 的某个字段
					task.Params["assigned_node_id"] = nodeID

					subTasks = append(subTasks, task)
				}

				platformTask := &paradigm.PlatformTask{
					ID:          taskID,
					TaskName:    "平台任务申报",
					SubTasks:    subTasks,
					IsScheduled: isScheduled,
					Status:      "running",
					CreatedAt:   time.Now(),
				}

				if isScheduled {
					platformTask.ExecutionType = "定时任务"
				} else {
					platformTask.ExecutionType = "即时任务"
				}

				// 汇总参数描述
				if len(subTasks) > 0 {
					platformTask.Parameters = fmt.Sprintf("Stock: %s (%s); Horizon: %v", subTasks[0].Params["stockName"], subTasks[0].Params["stockCode"], subTasks[0].Params["horizon"])
					if len(subTasks) > 1 {
						platformTask.Parameters += fmt.Sprintf("... (共%d个子任务)", len(subTasks))
					}
				}

				err := e.dbService.SetPlatformTask(platformTask)
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "任务持久化失败",
						Code:    "E100003",
						Data:    false,
					})
					return
				}

				// 平台任务的子任务只调度一次：直接推入调度队列，不经过区块链/Tracker
				for _, subTask := range subTasks {
					t := subTask // capture loop variable
					// 从请求参数中读取数据量，未提供则使用默认 SlotSize
					taskSize := int32(3000)
					if v, ok := t.Params["size"]; ok {
						switch s := v.(type) {
						case float64:
							taskSize = int32(s)
						case int:
							taskSize = int32(s)
						case int32:
							taskSize = s
						}
					}
					e.channel.UnprocessedTasks <- paradigm.UnprocessedTask{
						TaskID:   t.Sign,
						SlotSize: taskSize,
						Size:     taskSize,
						Model:    paradigm.ABM,
						Params:   t.Params,
					}
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    true,
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case ANALYZED_STOCKS:
		httpService := HttpService{
			Url:    "/market/analyzed-stocks",
			Method: "GET",
			Handler: func(c *gin.Context) {
				tasks, err := e.dbService.GetFinishedTasks()
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取分析已完成任务失败",
						Code:    "E100004",
						Data:    nil,
					})
					return
				}

				result := make([]map[string]interface{}, 0)
				for _, t := range tasks {
					p := t.Params
					stock := map[string]interface{}{
						"stockId":   p["stockId"],
						"stockCode": p["stockCode"],
						"stockName": p["stockName"],
						"label":     fmt.Sprintf("%v %v", p["stockCode"], p["stockName"]),
						"taskId":    t.PlatformTaskID,
						"taskName":  fmt.Sprintf("%s 风险监测", p["stockName"]), // 如果没有platform task name就拼一下
					}

					// 如果有关联的平台任务，可以尝试获取它的名字
					if t.PlatformTaskID != nil {
						pt, err := e.dbService.GetPlatformTaskByID(*t.PlatformTaskID)
						if err == nil && pt != nil {
							stock["taskName"] = pt.TaskName
						}
					}

					result = append(result, stock)
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    result,
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case ABM_PARAMETERS:
		httpService := HttpService{
			Url:    "/simulation/abm-parameters",
			Method: "GET",
			Handler: func(c *gin.Context) {
				// 获取预定义的 ABM 模型结构参数 (从配置加载)
				parameters := e.config.AbmParameters

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    parameters,
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case ORDER_DYNAMICS:
		httpService := HttpService{
			Url:    "/dashboard/order_dynamics",
			Method: "GET",
			Handler: func(c *gin.Context) {
				taskId := c.Query("taskId")
				stockId := c.Query("stockId")
				if taskId == "" || stockId == "" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId and stockId required", Code: "E100008", Data: nil})
					return
				}

				// 从节点获取实时数据
				data, err := e.FetchNodeAnalytics(taskId, stockId, paradigm.OrderDynamics)
				if err != nil {
					paradigm.Log("ERROR", fmt.Sprintf("Failed to fetch live dynamics: %v", err))
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取实时订单态势失败: " + err.Error(),
						Code:    "E100009",
						Data:    nil,
					})
					return
				}
				c.JSON(http.StatusOK, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
			},
		}
		return &httpService, nil
	case PRICE_SYNTH_DOWNLOAD:
		httpService := HttpService{
			Url:     "/dashboard/price_synthesis/download",
			Method:  "GET",
			Handler: e.HandleSimulationDownload,
		}
		return &httpService, nil
	case PRICE_SYNTH:
		httpService := HttpService{
			Url:    "/dashboard/price_synthesis",
			Method: "GET",
			Handler: func(c *gin.Context) {
				taskId := c.Query("taskId")
				stockId := c.Query("stockId")
				if taskId == "" || stockId == "" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId and stockId required", Code: "E100008", Data: nil})
					return
				}

				data, err := e.FetchNodeAnalytics(taskId, stockId, paradigm.PriceSynthesis)
				if err != nil {
					paradigm.Log("ERROR", fmt.Sprintf("Failed to fetch live price synthesis: %v", err))
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取价格合成分析失败: " + err.Error(),
						Code:    "E100010",
						Data:    nil,
					})
					return
				}
				c.JSON(http.StatusOK, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
			},
		}
		return &httpService, nil
	case CRASH_RISK:
		httpService := HttpService{
			Url:    "/dashboard/crash_risk_warning",
			Method: "GET",
			Handler: func(c *gin.Context) {
				taskId := c.Query("taskId")
				stockId := c.Query("stockId")
				if taskId == "" || stockId == "" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId and stockId required", Code: "E100008", Data: nil})
					return
				}

				data, err := e.FetchNodeAnalytics(taskId, stockId, paradigm.CrashRisk)
				if err != nil {
					paradigm.Log("ERROR", fmt.Sprintf("Failed to fetch live crash risk: %v", err))
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取崩盘风险预警失败: " + err.Error(),
						Code:    "E100011",
						Data:    nil,
					})
					return
				}
				c.JSON(http.StatusOK, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
			},
		}
		return &httpService, nil
	case INVESTOR_COMP:
		httpService := HttpService{
			Url:    "/dashboard/investor_composition",
			Method: "GET",
			Handler: func(c *gin.Context) {
				taskId := c.Query("taskId")
				stockId := c.Query("stockId")
				if taskId == "" || stockId == "" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId and stockId required", Code: "E100008", Data: nil})
					return
				}

				data, err := e.FetchNodeAnalytics(taskId, stockId, paradigm.InvestorComposition)
				if err != nil {
					paradigm.Log("ERROR", fmt.Sprintf("Failed to fetch live investor composition: %v", err))
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取投资者构成失败: " + err.Error(),
						Code:    "E100012",
						Data:    nil,
					})
					return
				}
				c.JSON(http.StatusOK, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
			},
		}
		return &httpService, nil
	case PERF_COMPARISON:
		httpService := HttpService{
			Url:    "/dashboard/performance_comparison",
			Method: "GET",
			Handler: func(c *gin.Context) {
				taskId := c.Query("taskId")
				stockId := c.Query("stockId")
				if taskId == "" || stockId == "" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId and stockId required", Code: "E100008", Data: nil})
					return
				}

				data, err := e.FetchNodeAnalytics(taskId, stockId, paradigm.PerformanceComparison)
				if err != nil {
					paradigm.Log("ERROR", fmt.Sprintf("Failed to fetch live performance comparison: %v", err))
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取模型评估失败: " + err.Error(),
						Code:    "E100013",
						Data:    nil,
					})
					return
				}
				c.JSON(http.StatusOK, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
			},
		}
		return &httpService, nil
	default:
		paradigm.Error(paradigm.NetworkError, "Unknown HTTP Service")
		//LogWriter.Log("ERROR", fmt.Sprintf("%s: %s", paradigm.ErrorToString(paradigm.NetworkError), "Unknown Http Service"))
		return nil, fmt.Errorf("unknown Http Service")
	}
}
