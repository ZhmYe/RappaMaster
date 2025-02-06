package paradigm

type HttpRequestInterface interface {
	ToJson() map[interface{}]interface{}
	ConstructResponse() HttpResponse
}

type HttpResponse struct {
	Message string      `json:"msg"`    // 消息
	Code    string      `json:"status"` // 状态
	Data    interface{} `json:"data"`   // 数据
}

type HttpInitTaskRequest struct {
	Sign string // task sign
	//Slot       int32                  // slot index
	Size       int32                  // data size
	Model      string                 // 模型名称
	Params     map[string]interface{} // 不确定的模型参数
	IsReliable bool                   // 是否需要可信证明
}

func (r *HttpInitTaskRequest) ToJson() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"sign": r.Sign,
		//"slot":       r.Slot,
		"size":       r.Size,
		"Model":      r.Model,
		"Params":     r.Params,
		"isReliable": r.IsReliable,
	}
}
func (r *HttpInitTaskRequest) ConstructResponse() HttpResponse {
	return HttpResponse{
		Message: "Invalid JSON data",
		Code:    "error",
		Data:    nil,
	}
}
