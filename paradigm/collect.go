package paradigm

// CollectSlotItem 这里的Slot已经经过了finalized，无需记录其他的状态
// 考虑到用户可能不是一次性下载所有数据，更常见的应该是download多少数据
// 所以要做的其实是按序遍历下来，要注意存储有序
type CollectSlotItem struct {
	Sign string   // 这里其实可以不记录sign
	Hash SlotHash // 主要是以这个作为标识
	Size int32    // 表示这个slot包含了多少的数据
}

// CollectSlotInstance 这里需要在真的开始collect的，有一个结构体来承载所有数据
type CollectSlotInstance struct {
	// todo
}
