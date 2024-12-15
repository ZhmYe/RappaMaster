package paradigm

import "BHLayer2Node/pb/service"

type HeartBeat struct {
	Commits   []*service.JustifiedSlot
	Finalizes []*service.JustifiedSlot
	Tasks     map[string]int32
	Epoch     int
}
