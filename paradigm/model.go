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

func NameToModelType(name string) SupportModelType {
	switch name {
	case "CTGAN":
		return CTGAN
	case "AGSS":
		return AGSS
	default:
		panic("Unknown model type!!!")
	}
}

type ModelOutputType int // 模型输出类型
const (
	DATAFRAME ModelOutputType = iota // 表格数据
	NETWORK                          // 图数据
	// todo
)

func ModelOutputTypeToString(t ModelOutputType) string {
	switch t {
	case DATAFRAME:
		return "Dataframe"
	case NETWORK:
		return "Network"
	default:
		panic("Unknown model output type!!!")
	}
}
