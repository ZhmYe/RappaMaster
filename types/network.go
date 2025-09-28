package types

//type HttpRequestInterface interface {
//	ToJson() map[interface{}]interface{}
//	ConstructResponse() HttpResponse
//}

type HttpResponse struct {
	Message string      `json:"msg"`    // 消息
	Code    string      `json:"status"` // 状态
	Data    interface{} `json:"data"`   // 数据
}

type BlockedGrpcPayload[MSG any, RES any] struct {
	msg  MSG
	conn chan RES
}

func NewBlockedGrpcPayload[MSG any, RES any](msg MSG) (BlockedGrpcPayload[MSG, RES], chan RES) {
	conn := make(chan RES)
	return BlockedGrpcPayload[MSG, RES]{
		conn: conn,
		msg:  msg,
	}, conn
}
