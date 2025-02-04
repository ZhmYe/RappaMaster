package paradigm

type Query interface {
	GenerateResponse(data interface{}) map[interface{}]interface{}
	FromHttpJson(rawData map[interface{}]interface{}) bool
}
