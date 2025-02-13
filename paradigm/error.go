package paradigm

import (
	"fmt"
)

type RappaError struct {
	errorType    ErrorEnum // 错误类型
	errorMessage string    // 错误信息
}

func (e *RappaError) Error() string {
	return fmt.Sprintf("Error %s: %s", ErrorToString(e.errorType), e.errorMessage)
}
func NewRappaError(errorType ErrorEnum, message string) RappaError {
	return RappaError{
		errorType:    errorType,
		errorMessage: message,
	}
}

type ErrorEnum int

const (
	RuntimeError = iota
	NetworkError
	ValueError
	ChunkRecoverError
	DataTransformError
	NotImplError
	ExecutorError
	SlotLifeError
)

func ErrorToString(error ErrorEnum) string {
	switch error {
	case RuntimeError:
		return "RuntimeError"
	case NetworkError:
		return "NetworkError"
	case ValueError:
		return "ValueError"
	case ChunkRecoverError:
		return "ChunkRecoverError"
	case DataTransformError:
		return "DataTransformError"
	case NotImplError:
		return "NotImplError"
	case ExecutorError:
		return "ExecutorError"
	case SlotLifeError:
		return "SlotLifeError"
	default:
		return "Unknown Error"
	}
}

//func RaiseError(errorType ErrorEnum, errorMessage string, isPanic bool) {
//	Log("ERROR", fmt.Sprintf("%s: %s", ErrorToString(errorType), errorMessage))
//	if isPanic {
//		panic(fmt.Sprintf("%s: %s", ErrorToString(errorType), errorMessage))
//	}
//}
