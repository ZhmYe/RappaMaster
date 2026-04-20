package HTTP

import (
	"BHLayer2Node/paradigm"
	"testing"
)

func TestEncodeAnalysisTypeRequest(t *testing.T) {
	cases := []struct {
		name     string
		analType paradigm.AnalysisType
		options  map[string]string
		expected string
	}{
		{
			name:     "order dynamics with date",
			analType: paradigm.OrderDynamics,
			options:  map[string]string{"date": "2026-04-19"},
			expected: "order_dynamics?date=2026-04-19",
		},
		{
			name:     "performance comparison with selected model",
			analType: paradigm.PerformanceComparison,
			options:  map[string]string{"selectedModel": "GBM"},
			expected: "performance_comparison?selectedModel=GBM",
		},
		{
			name:     "empty options",
			analType: paradigm.OrderDynamics,
			options:  map[string]string{},
			expected: "order_dynamics",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := encodeAnalysisTypeRequest(tc.analType, tc.options)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestBuildCrashRiskTopRiskList(t *testing.T) {
	items := []analyticsQueryItem{
		{
			StockCode: "600000",
			StockName: "浦发银行",
			Data: map[string]interface{}{
				"forecastSeries": []interface{}{
					map[string]interface{}{"crashProb": 0.12},
					map[string]interface{}{"crashProb": 0.35},
				},
			},
		},
		{
			StockCode: "000001",
			StockName: "平安银行",
			Data: map[string]interface{}{
				"forecastSeries": []interface{}{
					map[string]interface{}{"crashProb": 0.28},
				},
			},
		},
		{
			StockCode: "600519",
			StockName: "贵州茅台",
			Data: map[string]interface{}{
				"summary": map[string]interface{}{"predictionProbability": 0.18},
			},
		},
	}

	got := buildCrashRiskTopRiskList(items)
	if len(got) != 3 {
		t.Fatalf("expected 3 ranked items, got %d", len(got))
	}

	if got[0]["code"] != "600000" || got[0]["probability"] != 0.35 {
		t.Fatalf("unexpected rank1 item: %#v", got[0])
	}
	if got[1]["code"] != "000001" || got[1]["probability"] != 0.28 {
		t.Fatalf("unexpected rank2 item: %#v", got[1])
	}
	if got[2]["code"] != "600519" || got[2]["probability"] != 0.18 {
		t.Fatalf("unexpected rank3 item: %#v", got[2])
	}
}

func TestAttachCrashRiskTopRiskListToItems(t *testing.T) {
	items := []analyticsQueryItem{
		{
			StockCode: "600000",
			StockName: "浦发银行",
			Data: map[string]interface{}{
				"summary":        map[string]interface{}{"predictionProbability": 0.22},
				"forecastSeries": []interface{}{map[string]interface{}{"crashProb": 0.22}},
				"topRiskList":    []interface{}{},
			},
		},
		{
			StockCode: "000001",
			StockName: "平安银行",
			Data: map[string]interface{}{
				"summary":        map[string]interface{}{"predictionProbability": 0.31},
				"forecastSeries": []interface{}{map[string]interface{}{"crashProb": 0.31}},
				"topRiskList":    []interface{}{},
			},
		},
	}

	got := attachCrashRiskTopRiskListToItems(items)
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}

	firstPayload := got[0].Data.(map[string]interface{})
	list, ok := firstPayload["topRiskList"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected injected topRiskList slice, got %#v", firstPayload["topRiskList"])
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 ranked entries, got %d", len(list))
	}
	if list[0]["code"] != "000001" {
		t.Fatalf("expected highest risk stock 000001, got %#v", list[0])
	}
}
