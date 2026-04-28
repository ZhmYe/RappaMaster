package HTTP

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const scheduledABMV2Horizon = "1天 (T+1)"

func isScheduledCreateTask(c *gin.Context) bool {
	return strings.EqualFold(c.Query("isScheduled"), "true")
}

func (e *HttpEngine) buildScheduledABMV2RawTasks() ([]map[string]interface{}, error) {
	stockCodes, err := e.listABMSupportedStockCodes()
	if err != nil {
		return nil, err
	}
	if len(stockCodes) == 0 {
		return nil, fmt.Errorf("no supported ABM stocks found")
	}

	tasks := make([]map[string]interface{}, 0, len(stockCodes))
	for _, stockCode := range stockCodes {
		// 定时任务由外部调度方按周期触发；任务内容固定为所有已调参股票的 T+1 推演，
		// 不接受请求体里的额外覆盖参数，避免不同周期任务口径不一致。
		tasks = append(tasks, map[string]interface{}{
			"stockCode": stockCode,
			"stockName": scheduledABMV2StockName(stockCode),
			"horizon":   scheduledABMV2Horizon,
		})
	}
	return tasks, nil
}

func (e *HttpEngine) listABMSupportedStockCodes() ([]string, error) {
	paramsDir := abmStockParamDir(&e.config)
	dataDir := abmStockDataDir(&e.config)

	entries, err := os.ReadDir(paramsDir)
	if err != nil {
		return nil, err
	}

	stockCodes := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		stockCode := normalizeABMStockCode(entry.Name())
		if stockCode == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(paramsDir, stockCode, "model_params.json")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(dataDir, stockCode+".csv")); err != nil {
			continue
		}
		stockCodes = append(stockCodes, stockCode)
	}
	sort.Strings(stockCodes)
	return stockCodes, nil
}

func scheduledABMV2StockName(stockCode string) string {
	// 离线参数文件当前不包含股票简称；定时批量任务先用代码占位，保证查询与调度主键稳定。
	return stockCode
}
