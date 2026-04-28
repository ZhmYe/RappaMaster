package HTTP

import (
	"BHLayer2Node/paradigm"
	"os"
	"path/filepath"
	"strings"
)

const defaultABMStockDataDir = "/root/rappa/stockdata"

func abmStockDataDir(config *paradigm.BHLayer2NodeConfig) string {
	if value := strings.TrimSpace(os.Getenv("ABM_STOCK_DATA_DIR")); value != "" {
		return value
	}
	if config != nil {
		if value := strings.TrimSpace(config.ABMStockDataDir); value != "" {
			return value
		}
	}
	return defaultABMStockDataDir
}

func abmStockParamDir(config *paradigm.BHLayer2NodeConfig) string {
	if value := strings.TrimSpace(os.Getenv("ABM_STOCK_PARAM_DIR")); value != "" {
		return value
	}
	if config != nil {
		if value := strings.TrimSpace(config.ABMStockParamDir); value != "" {
			return value
		}
	}
	return filepath.Join(abmStockDataDir(config), "params")
}
