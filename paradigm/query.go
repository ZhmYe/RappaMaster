package paradigm

import "fmt"

type Query interface {
	GenerateResponse(data interface{}) Response
	ParseRawDataFromHttpEngine(rawData map[interface{}]interface{}) bool
	SendResponse(response Response)
	ReceiveResponse() Response
	ToHttpJson() map[string]interface{}
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
