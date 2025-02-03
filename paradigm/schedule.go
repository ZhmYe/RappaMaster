package paradigm

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type SynthTaskSchedule struct {
	TaskID     TaskHash         // 任务标识
	ScheduleID ScheduleHash     // 第几次调度
	Size       int32            // 调度总量
	Model      SupportModelType // 模型名称
	Params     map[string]interface{}
	Slots      map[int]*Slot // 调度的Slot
}

func (s *SynthTaskSchedule) Print() {
	fmt.Printf("TaskID: %s\n", s.TaskID)
	fmt.Printf("ScheduleID: %s\n", s.ScheduleID)
	fmt.Printf("Size: %d KB\n", s.Size)
	fmt.Printf("Model: %s\n", s.Model)
	fmt.Println("Params:")
	for key, value := range s.Params {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 打印 Slot 信息，使用表格格式
	fmt.Println("\nSlot Information:")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "SlotID\tScheduleSize\tStatus\tError")
	for _, slot := range s.Slots {
		status := ""
		switch slot.Status {
		case Finished:
			status = "Finished"
		case Processing:
			status = "Processing"
		case Failed:
			status = "Failed"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", slot.SlotID, slot.ScheduleSize, status, slot.err)
	}
	w.Flush() // 刷新输出
}

//// TaskSchedule 描述任务分配到的节点
//type TaskSchedule struct {
//	Sign      string           // 任务标识
//	Slot      int32            // 第几次被调用，这里会出现一次调度不被接受，因此需要多次调度的，称为slot
//	Size      int32            // 数据总量
//	Model     SupportModelType // 模型名称
//	Params    map[string]interface{}
//	Schedules []ScheduleItem
//}
//type ScheduleItem struct {
//	//Sign   string
//	//Slot   int
//	Size       int32
//	NID        int
//	Commitment []byte
//	Hash       SlotHash
//	//Model  string
//	//Params map[string]interface{}
//}
