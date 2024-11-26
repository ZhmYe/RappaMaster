package paradigm

type CommitSlotItem struct {
	Nid     int
	Process int
	Sign    string
	Slot    int
}

func (c *CommitSlotItem) Record() ScheduleItem {
	return ScheduleItem{
		Size: c.Process,
		NID:  c.Nid,
	}
}
