package paradigm

type SupportModelType int

const (
	CTGAN SupportModelType = iota
	// TODO 后续有支持的新模型在这里加上
)

func ModelTypeToString(t SupportModelType) string {
	switch t {
	case CTGAN:
		return "CTGAN"
	default:
		panic("Unknown model type!!!")
	}
}
