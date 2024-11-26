package HTTP

import "BHCoordinator/Config"

type HttpInterface interface {
	HandleRequest()
	Start()
	Setup(config Config.BHCoordinatorConfig)
}
