package paradigm

type SupportModelType int

const (
	CTGAN SupportModelType = iota
	BAED
	FINKAN
	ABM
	// TODO 后续有支持的新模型在这里加上
)

func ModelTypeToString(t SupportModelType) string {
	switch t {
	case CTGAN:
		return "CTGAN"
	case BAED:
		return "BAED"
	case FINKAN:
		return "FINKAN"
	case ABM:
		return "ABM"
	default:
		panic("Unknown model type!!!")
	}
}

func NameToModelType(name string) SupportModelType {
	switch name {
	case "CTGAN":
		return CTGAN
	case "BAED":
		return BAED
	case "FINKAN":
		return FINKAN
	case "ABM":
		return ABM
	default:
		e := Error(RuntimeError, "Unknown model type")
		panic(e.Error())
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
		e := Error(RuntimeError, "Unknown output type")
		panic(e.Error())
	}
}
func ModelOutputTypeToFileExt(t ModelOutputType) string {
	switch t {
	case DATAFRAME:
		return "csv"
	case NETWORK:
		return "json"
	default:
		return "txt"
	}
}

type AnalysisType int

const (
	OrderDynamics AnalysisType = iota
	PriceSynthesis
	CrashRisk
	InvestorComposition
	PerformanceComparison
)

func (t AnalysisType) String() string {
	switch t {
	case OrderDynamics:
		return "order_dynamics"
	case PriceSynthesis:
		return "price_synthesis"
	case CrashRisk:
		return "crash_risk"
	case InvestorComposition:
		return "investor_composition"
	case PerformanceComparison:
		return "performance_comparison"
	default:
		return "unknown"
	}
}
