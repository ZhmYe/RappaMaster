package types

type HttpInitTaskRequest struct {
	Sign string // task sign
	//Slot       int32                  // slot index
	Name       string                 // 任务名称
	Size       int32                  // data size
	Model      string                 // 模型名称
	Params     map[string]interface{} // 不确定的模型参数
	IsReliable bool                   // 是否需要可信证明
}
