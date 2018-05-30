package control

// GetTaskGather 获取任务的集合情况：签到人数
func (c Controller) GetTaskGather(taskID int) int {
	checkCount := db.GetCheckCountsFromTask(taskID)
	return checkCount
}

func (c Controller) GetTaskGatherMems(taskID, offset, counts int) *MemList {
	memList := MemList{}
	db.GetTaskAcceptCount(taskID)
}
