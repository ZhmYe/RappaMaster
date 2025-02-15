package paradigm

//type HttpRequestInterface interface {
//	ToJson() map[interface{}]interface{}
//	ConstructResponse() HttpResponse
//}

type HttpResponse struct {
	Message string      `json:"msg"`    // 消息
	Code    string      `json:"status"` // 状态
	Data    interface{} `json:"data"`   // 数据
}
