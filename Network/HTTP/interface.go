package HTTP

import "BHLayer2node/Config"

type HttpInterface interface {
	HandleRequest()
	Start()
	Setup(config Config.BHLayer2NodeConfig)
}
