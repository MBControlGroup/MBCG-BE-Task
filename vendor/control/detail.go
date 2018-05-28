package control

import (
	"model"
	"time"
)

// GetTaskDetail 获取任务详情
func (c Controller) GetTaskDetail(taskID int, watchAdminID int) (*model.TaskInfo, error) {
	task := make([]model.TaskInfo, 1)
	// 任务title, launch_datetime, gather_datetime, place_name等
	task[0] = db.GetTaskDetailFromDB(taskID)
	// 任务的launcher（发起任务的组织/单位名称）
	c.writeOrgOfficeNamesFromAdminIDs(task)
	finishTime, _ := time.Parse("2006-01-02 15:04:05", task[0].FinishTime)
	// 如果任务未完成
	if finishTime.Unix() > time.Now().Unix() {
		// 任务的status, status_detail
		c.calTasksStatus(task)
		// 任务的is_launcher，即判断查看该任务的Admin是否为任务发起者
		if watchAdminID == task[0].AdminID {
			task[0].IsLauncher = true
		}
	}
	return &task[0], nil
}
