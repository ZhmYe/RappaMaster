package test

import (
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/paradigm"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestABMParametersEndpointUsesTunedParamsAndDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ABM_STOCK_PARAM_DIR", dir)

	content := []byte(`
N_FT = 123
K1 = 1.9855
MU_L = -1.6
IGNORED = 999
`)
	if err := os.WriteFile(filepath.Join(dir, "config_600000.py"), content, 0o644); err != nil {
		t.Fatalf("write tuned params: %v", err)
	}

	service := newABMParametersService(t)

	defaultData := callABMParameters(t, service, "")
	assertParam(t, defaultData, "N_FT", 300, "default")
	assertParam(t, defaultData, "K1", 0.2855, "default")

	tunedData := callABMParameters(t, service, "stockCode=600000")
	assertParam(t, tunedData, "N_FT", float64(123), "tuned")
	assertParam(t, tunedData, "K1", 1.9855, "tuned")
	assertParam(t, tunedData, "MU_L", -1.6, "tuned")
	if _, exists := tunedData["IGNORED"]; exists {
		t.Fatalf("unexpected non-tunable key in response: %#v", tunedData["IGNORED"])
	}

	missingData := callABMParameters(t, service, "stockCode=999999")
	assertParam(t, missingData, "N_FT", 300, "default")
	assertParam(t, missingData, "K1", 0.2855, "default")
}

func newABMParametersService(t *testing.T) *HTTP.HttpService {
	t.Helper()
	gin.SetMode(gin.TestMode)

	engine := &HTTP.HttpEngine{}
	engine.Setup(paradigm.BHLayer2NodeConfig{
		AbmParameters: map[string]interface{}{
			"N_FT": map[string]interface{}{
				"label":   "基本面交易者数量",
				"type":    "int",
				"default": 300,
				"min":     10,
				"max":     8000,
			},
		},
	})

	service, err := engine.GetHttpService(HTTP.ABM_PARAMETERS)
	if err != nil {
		t.Fatalf("get ABM parameters service: %v", err)
	}
	return service
}

func callABMParameters(t *testing.T, service *HTTP.HttpService, rawQuery string) map[string]interface{} {
	t.Helper()

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, "/simulation/abm_parameters?"+rawQuery, nil)
	context.Request = request

	service.Handler(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response paradigm.HttpResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected response data object, got %#v", response.Data)
	}
	return data
}

func assertParam(t *testing.T, data map[string]interface{}, key string, defaultValue interface{}, source string) {
	t.Helper()

	rawSpec, ok := data[key]
	if !ok {
		t.Fatalf("missing param %s in %#v", key, data)
	}
	spec, ok := rawSpec.(map[string]interface{})
	if !ok {
		t.Fatalf("expected %s spec object, got %#v", key, rawSpec)
	}
	if !sameJSONValue(spec["default"], defaultValue) {
		t.Fatalf("expected %s default=%#v, got %#v", key, defaultValue, spec["default"])
	}
	if spec["source"] != source {
		t.Fatalf("expected %s source=%q, got %#v", key, source, spec["source"])
	}
}

func sameJSONValue(actual interface{}, expected interface{}) bool {
	actualFloat, actualNumeric := numericValue(actual)
	expectedFloat, expectedNumeric := numericValue(expected)
	if actualNumeric && expectedNumeric {
		return actualFloat == expectedFloat
	}
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

func numericValue(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
