package HTTP

import (
	"RappaMaster/config"
	"context"
	"testing"
)

var httpEngine *HttpEngine

func TestHttpEngine(t *testing.T) {
	var err error
	httpEngine, err = NewHttpEngine(config.GlobalSystemConfig.HttpConfig)
	if err != nil {
		panic(err)
	}
	httpEngine.Start(context.Background())
}
