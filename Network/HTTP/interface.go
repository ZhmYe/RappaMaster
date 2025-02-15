package HTTP

import (
	"BHLayer2Node/paradigm"
)

type HttpInterface interface {
	HandleRequest()
	Start()
	Setup(config paradigm.BHLayer2NodeConfig)
}
