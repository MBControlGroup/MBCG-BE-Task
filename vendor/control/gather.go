package control

import "math"

type GatherDetail struct {
	CheckCount int `json:"check_count"`
}

// GetTaskGather 获取任务的集合情况：签到人数
func (c Controller) GetTaskGather(taskID int) *GatherDetail {
	gather := GatherDetail{}
	gather.CheckCount = db.GetCheckCountsFromTask(taskID)
	return &gather
}

// GetTaskGatherMems 获取任务的集合人员列表
func (c Controller) GetTaskGatherMems(taskID, offset, count int) *MemList {
	memList := MemList{}
	memList.MemCount = db.GetTaskAcceptCount(taskID)
	memList.PageCount = int(math.Ceil(float64(memList.MemCount) / float64(count)))
	memList.Members = db.GetTaskGatherMems(taskID, offset, count)
	return &memList
}
