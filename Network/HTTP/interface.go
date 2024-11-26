package HTTP

import "BHLayer2Node/Config"

type HttpInterface interface {
	HandleRequest()
	Start()
	Setup(config Config.BHLayer2NodeConfig)
}
