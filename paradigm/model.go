package paradigm

type SupportModelType int

const (
	CTGAN SupportModelType = iota
	AGSS
	// TODO 后续有支持的新模型在这里加上
)

func ModelTypeToString(t SupportModelType) string {
	switch t {
	case CTGAN:
		return "CTGAN"
	case AGSS:
		return "AGSS"
	default:
		panic("Unknown model type!!!")
	}
}
