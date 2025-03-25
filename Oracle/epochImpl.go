package Oracle

import "BHLayer2Node/paradigm"

// 记录epoch
func (o *PersistedOracle) setEpoch(epochRecord *paradigm.DevEpoch) {
	o.db.Create(epochRecord)
}
