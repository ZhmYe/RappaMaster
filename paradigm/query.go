package paradigm

import (
	"encoding/json"
	"fmt"
	"time"
)

type DevQueryType int

const (
	EpochQuery          = iota // 查询一个epoch内发生了什么
	TaskSlotQuery              // 查询一个task在某个slot内的完成情况
	EpochRangeQuery            // 查询epoch_i ~ epoch_j
	TaskSlotRangeQuery         // 查询某个task在slot_i ~ slot_j内的完成情况
	TxReceiptQuery             // 查询某个tx的receipt
	BatchTxReceiptQuery        // 查询某些tx的receipt
	// todo
	BlockInfoQuery // 查询区块信息
	TxInfoQuery    // 查询交易信息
	EpochNumQuery  // 查询区块链区块数量
	TxNumQuery     // 查询区块链交易数量

)

// 查询Response 响应结构
type QueryResponse struct {
	Code    int         `json:"code"`
	Result  interface{} `json:"result"`
	Message string      `json:"message"`
	Type    int         `json:"type"`
}

// NewSuccessResponse 成功响应构造
func NewSuccessResponse(result interface{}, queryType int) *QueryResponse {
	return &QueryResponse{
		Code:    0,
		Result:  result,
		Message: "success",
		Type:    queryType,
	}
}

// NewErrorResponse 错误响应构造（含状态码）
func NewErrorResponse(code int, message string, queryType int) *QueryResponse {
	return &QueryResponse{
		Code:    code,
		Result:  nil,
		Message: message,
		Type:    queryType,
	}
}

// QueryRequest 定义统一的查询请求结构，所有查询请求必须包含公共属性
type QueryRequest struct {
	// QueryType 标识查询类型
	QueryType DevQueryType `json:"queryType"`
	// Params 存放各查询类型需要的参数
	Params map[string]interface{} `json:"params"`
	// RequestID 可选的请求标识
	RequestID string `json:"requestID,omitempty"`
	// Timestamp 请求时间戳（Unix 秒）
	Timestamp int64 `json:"timestamp,omitempty"`
}

// NewQueryRequest 创建基础请求对象（自动填充时间戳）
func NewQueryRequest(queryType DevQueryType, params map[string]interface{}) *QueryRequest {
	return &QueryRequest{
		QueryType: queryType,
		Params:    params,
		Timestamp: time.Now().Unix(),
	}
}

// WithRequestID 链式调用设置 RequestID
func (q *QueryRequest) WithRequestID(id string) *QueryRequest {
	q.RequestID = id
	return q
}

// AddParam 安全添加参数（处理 nil map）
func (q *QueryRequest) AddParam(key string, value interface{}) {
	if q.Params == nil {
		q.Params = make(map[string]interface{})
	}
	q.Params[key] = value
}

// ValidateParams 校验参数是否包含必需字段
func (q *QueryRequest) ValidateParams(requiredKeys []string) error {
	missing := make([]string, 0)
	for _, key := range requiredKeys {
		if _, exists := q.Params[key]; !exists {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required params: %v", missing)
	}
	return nil
}

// 辅助函数：从 JSON 字符串反序列化
func DeserializeData(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
