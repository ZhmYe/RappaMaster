package paradigm

import (
	"fmt"
)

type Query interface {
	GenerateResponse(data interface{}) Response
	ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool
	SendResponse(response Response)
	ReceiveResponse() Response
	ToHttpJson() map[string]interface{}
}

type BasicChannelQuery struct {
	responseChannel chan Response
	// 这里没有别的参数
}

func (q *BasicChannelQuery) SendResponse(response Response) {
	q.responseChannel <- response
	close(q.responseChannel)
}
func (q *BasicChannelQuery) ReceiveResponse() Response {
	return <-q.responseChannel
}
func NewBasicChannelQuery() BasicChannelQuery {
	return BasicChannelQuery{responseChannel: make(chan Response, 1)}
}

// DoubleChannelQuery 需要和链交互，因此有一个给client传递消息的channel
type DoubleChannelQuery struct {
	BasicChannelQuery
	infoChannel chan interface{}
}

func (q *DoubleChannelQuery) SendInfo(info interface{}) {
	q.infoChannel <- info
	close(q.infoChannel)
}
func (q *DoubleChannelQuery) ReceiveInfo() interface{} {
	return <-q.infoChannel
}
func NewDoubleChannelQuery() DoubleChannelQuery {
	return DoubleChannelQuery{
		BasicChannelQuery: NewBasicChannelQuery(),
		infoChannel:       make(chan interface{}),
	}
}

type Response interface {
	ToHttpJson() map[string]interface{}
	Error() string
}

type SuccessResponse struct {
	rawData map[string]interface{}
}

func NewSuccessResponse(data map[string]interface{}) *SuccessResponse {
	return &SuccessResponse{
		rawData: data,
	}
}

func (r *SuccessResponse) ToHttpJson() map[string]interface{} {
	return r.rawData
}
func (r *SuccessResponse) Error() string {
	return ""
}

type ErrorResponse struct {
	errorType    ErrorEnum
	errorMessage string
}

func (e *ErrorResponse) ToHttpJson() map[string]interface{} {
	return map[string]interface{}{"error": ErrorToString(e.errorType), "errorMessage": e.errorMessage}
}
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", ErrorToString(e.errorType), e.errorMessage)
}
func NewErrorResponse(errorType ErrorEnum, errorMessage string) *ErrorResponse {
	return &ErrorResponse{
		errorType:    errorType,
		errorMessage: errorMessage,
	}
}

type LatestBlockchainInfo struct {
	LatestTxs     []*PackedTransaction
	LatestEpoch   []*DevEpoch
	NbFinalized   int32
	SynthData     int32
	NbEpoch       int32
	NbBlock       int32
	NbTransaction int32
}
