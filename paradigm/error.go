package paradigm

import (
	"fmt"
)

// RappaError gives an unified format of error in rappa
type RappaError struct {
	errorType    ErrorEnum
	errorMessage string
	err          error
}

func (e *RappaError) Error() string {
	return fmt.Sprintf("%s: %s, error: %v", ErrorToString(e.errorType), e.errorMessage, e.err)
}
func NewRappaError(errorType ErrorEnum, message string, err error) RappaError {
	return RappaError{
		errorType:    errorType,
		errorMessage: message,
		err:          err,
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
	DatabaseError
	FileError
	UpchainError
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
	case DatabaseError:
		return "DatabaseError"
	case FileError:
		return "FileError"
	case UpchainError:
		return "UpchainError"
	default:
		return "Unknown Error"
	}
}

func RaiseError(errorType ErrorEnum, message string, err error) error {
	return &RappaError{
		errorType:    errorType,
		errorMessage: message,
		err:          err,
	}
}
