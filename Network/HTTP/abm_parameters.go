package HTTP

import (
	"BHLayer2Node/paradigm"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var abmTunableParamKeys = []string{
	"N_FT",
	"S_FT",
	"N_LMT",
	"ALPHA_L",
	"N_SMT",
	"ALPHA_S",
	"N_NT",
	"MU_L",
	"SIGMA_L",
	"K1",
	"K2",
	"BETA_L",
	"BETA_S",
	"DELTA_NT",
	"THETA",
	"MU",
	"DELTA",
	"RHO",
	"VOLUME",
	"GAMMA",
}

var abmIntParamKeys = map[string]bool{
	"N_FT":   true,
	"S_FT":   true,
	"N_LMT":  true,
	"N_SMT":  true,
	"N_NT":   true,
	"VOLUME": true,
}

var abmParamLabels = map[string]string{
	"N_FT":     "基本面交易者数量",
	"S_FT":     "基本面交易者交易间隔（step 维度）",
	"N_LMT":    "长期动量交易者数量",
	"ALPHA_L":  "长期动量趋势信号衰减/更新权重（EMA 系数）",
	"N_SMT":    "短期动量交易者数量",
	"ALPHA_S":  "短期动量趋势信号衰减/更新权重（EMA 系数）",
	"N_NT":     "噪音交易者数量",
	"MU_L":     "限价订单价格距离对数正态分布均值",
	"SIGMA_L":  "限价订单价格距离对数正态分布标准差",
	"K1":       "基本面交易者线性需求系数",
	"K2":       "基本面交易者非线性需求系数",
	"BETA_L":   "长期动量交易者需求计算系数",
	"BETA_S":   "短期动量交易者需求计算系数",
	"DELTA_NT": "噪音交易者总需求水平参数",
	"THETA":    "提交限价订单概率",
	"MU":       "提交市价订单概率",
	"DELTA":    "限价订单取消概率",
	"RHO":      "市价单与限价单提交概率之比",
	"VOLUME":   "单笔订单体积",
	"GAMMA":    "动量交易者需求计算系数",
}

var abmRuntimeParamDefaults = map[string]interface{}{
	"MU_L":     1.1,
	"SIGMA_L":  0.3,
	"K1":       0.2855,
	"K2":       0.4058,
	"BETA_L":   0.6905,
	"BETA_S":   0.0554,
	"DELTA_NT": 1.0325,
	"THETA":    0,
	"MU":       0,
	"DELTA":    0.005,
	"RHO":      0.2,
	"VOLUME":   100,
	"GAMMA":    10,
}

var stockCodeRegexp = regexp.MustCompile(`\d{6}`)

type abmModelParamsFile struct {
	StructuralParams map[string]interface{} `json:"structural_params"`
	CalibratedParams map[string]interface{} `json:"calibrated_params"`
}

func (e *HttpEngine) buildABMParametersResponse(stockCode string) map[string]interface{} {
	response := cloneABMParameters(e.config.AbmParameters)
	stockCode = normalizeABMStockCode(stockCode)
	tunedParams, hasTunedParams := loadABMStockTunedParams(stockCode, &e.config)

	for _, key := range abmTunableParamKeys {
		spec := ensureABMParamSpec(response, key)
		if value, ok := tunedParams[key]; hasTunedParams && ok {
			spec["default"] = value
			spec["source"] = "tuned"
			continue
		}
		if _, ok := spec["source"]; !ok {
			spec["source"] = "default"
		}
	}

	return response
}

func cloneABMParameters(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		if nested, ok := value.(map[string]interface{}); ok {
			copied := make(map[string]interface{}, len(nested))
			for nestedKey, nestedValue := range nested {
				copied[nestedKey] = nestedValue
			}
			dst[key] = copied
			continue
		}
		dst[key] = value
	}
	return dst
}

func ensureABMParamSpec(parameters map[string]interface{}, key string) map[string]interface{} {
	if existing, ok := parameters[key].(map[string]interface{}); ok {
		return existing
	}

	paramType := "float"
	if abmIntParamKeys[key] {
		paramType = "int"
	}
	spec := map[string]interface{}{
		"label":   abmParamLabels[key],
		"type":    paramType,
		"default": nil,
	}
	if value, ok := abmRuntimeParamDefaults[key]; ok {
		spec["default"] = value
	}
	parameters[key] = spec
	return spec
}

func loadABMStockTunedParams(stockCode string, config *paradigm.BHLayer2NodeConfig) (map[string]interface{}, bool) {
	stockCode = normalizeABMStockCode(stockCode)
	if stockCode == "" {
		return map[string]interface{}{}, false
	}

	paramsDir := abmStockParamDir(config)

	path := filepath.Join(paramsDir, stockCode, "model_params.json")
	file, err := os.Open(path)
	if err != nil {
		return map[string]interface{}{}, false
	}
	defer file.Close()

	allowed := make(map[string]bool, len(abmTunableParamKeys))
	for _, key := range abmTunableParamKeys {
		allowed[key] = true
	}

	var payload abmModelParamsFile
	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return map[string]interface{}{}, false
	}

	result := map[string]interface{}{}
	mergeABMParamValues(result, payload.StructuralParams, allowed)
	mergeABMParamValues(result, payload.CalibratedParams, allowed)
	return result, len(result) > 0
}

func mergeABMParamValues(dst map[string]interface{}, src map[string]interface{}, allowed map[string]bool) {
	for key, raw := range src {
		if !allowed[key] {
			continue
		}
		value, ok := normalizeABMParamValue(key, raw)
		if ok {
			dst[key] = value
		}
	}
}

func normalizeABMParamValue(key string, raw interface{}) (interface{}, bool) {
	value, ok := abmNumericValue(raw)
	if !ok {
		return nil, false
	}
	if abmIntParamKeys[key] {
		return int(value), true
	}
	return value, true
}

func abmNumericValue(raw interface{}) (float64, bool) {
	switch value := raw.(type) {
	case json.Number:
		parsed, err := value.Float64()
		return parsed, err == nil
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case int32:
		return float64(value), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func normalizeABMStockCode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if _, err := strconv.Atoi(raw); err == nil && len(raw) < 6 {
		return strings.Repeat("0", 6-len(raw)) + raw
	}
	match := stockCodeRegexp.FindString(raw)
	return match
}
