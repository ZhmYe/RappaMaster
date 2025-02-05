package paradigm

import (
	"BHLayer2Node/LogWriter"
	"fmt"
)

type ErrorEnum int

const (
	RuntimeError = iota
	NetworkError
	ValueError
)

func ErrorToString(error ErrorEnum) string {
	switch error {
	case RuntimeError:
		return "RuntimeError"
	case NetworkError:
		return "NetworkError"
	case ValueError:
		return "ValueError"
	default:
		return "Unknown Error"
	}
}
func RaiseError(errorType ErrorEnum, errorMessage string, isPanic bool) {
	LogWriter.Log("ERROR", fmt.Sprintf("%s: %s", ErrorToString(errorType), errorMessage))
	if isPanic {
		panic(fmt.Sprintf("%s: %s", ErrorToString(errorType), errorMessage))
	}
}
