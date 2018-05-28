package control

import (
	"math"
	"model"
	"sync"
	"time"
)

// GetTaskList 获取任务列表(执行中, 已完成). 区分单位/组织, 每页显示数目, 当前页
func (c Controller) GetTaskList(adminID, countsPerPage, offset int, isFinish bool) (*model.List, error) {
	tasklist := model.List{Tasks: make([]model.TaskInfo, 0)}
	adminIDs := []int{adminID}

	isOffice, _ := db.GetAdminType(adminID)
	// 获取Amin下属组织/单位的AdminID
	if isOffice { // Admin类型为单位
		officeID := db.GetOfficeIDFromAdminID(adminID)
		orgIDs := db.GetOrgIDsFromOffices([]int{officeID})
		if len(orgIDs) > 0 {
			subAdminIDs := db.GetAdminIDsFromOrgs(orgIDs)
			adminIDs = append(adminIDs, subAdminIDs...)
		}
		c.getAdminIDsInAllLowerOfficesAndOrgs([]int{officeID}, adminIDs)
	} else { // Admin类型为组织
		orgID := db.GetOrgIDFromAdminID(adminID)
		c.getAdminIDsInAllLowerOrgs([]int{orgID}, adminIDs)
	}
	// 从AdminIDs获取Tasks和Tasks总数
	tasks, taskCount := c.getTasksAndCountsFromAdmin(adminIDs, isFinish, offset, countsPerPage)

	tasklist.TaskCount = taskCount
	tasklist.PageCount = int(math.Ceil(float64(taskCount) / float64(countsPerPage)))
	tasklist.Tasks = tasks
	return &tasklist, nil
}

// 从单位获取其下属单位和组织(除去其本身和其所含组织)的AdminIDs
// 递归
func (c Controller) getAdminIDsInAllLowerOfficesAndOrgs(officeIDs []int, adminIDs []int) {
	// 获取目前单位的所有下属单位、AdminIDs
	lowerOffIDs := db.GetLowerOfficeIDsFromOffices(officeIDs)

	if len(lowerOffIDs) > 0 {
		adminIDsFromOffices := db.GetAdminIDsFromOffices(lowerOffIDs)
		adminIDs = append(adminIDs, adminIDsFromOffices...)

		// 从所有下属单位获取他们所含的组织、AdminIDs
		lowerOrgIDs := db.GetOrgIDsFromOffices(lowerOffIDs)
		if len(lowerOrgIDs) > 0 {
			adminIDsFromOrgs := db.GetAdminIDsFromOrgs(lowerOrgIDs)
			adminIDs = append(adminIDs, adminIDsFromOrgs...)
		}
		c.getAdminIDsInAllLowerOfficesAndOrgs(lowerOffIDs, adminIDs)
	}
}

// 给定highOrgIDs, 获取其所有下属(除去其本身)组织的AdminIDs
// 对组织的遍历, BFS
func (c Controller) getAdminIDsInAllLowerOrgs(highOrgIDs []int, adminIDs []int) {
	// 从该OrgID获取下属OrgIDs
	lowerOrgIDs := db.GetLowerOrgIDsFromOrgIDs(highOrgIDs)
	if len(lowerOrgIDs) > 0 {
		// 从下属Organizations获取所有AdminIDs
		subAdminIDs := db.GetAdminIDsFromOrgs(lowerOrgIDs)
		adminIDs = append(adminIDs, subAdminIDs...)
		// 从下属OrgIDs获取所有AdminIDs
		c.getAdminIDsInAllLowerOrgs(lowerOrgIDs, adminIDs)
	}
}

// 从AdminIDs获取他们发布的任务[]Tasklist(根据偏移量和所需数量), 他们发布任务的总数int
func (c Controller) getTasksAndCountsFromAdmin(adminIDs []int, isFinish bool, offset, countsPerPage int) ([]model.TaskInfo, int) {
	tasks := make([]model.TaskInfo, 0)
	// 获取Tasks的数量
	taskCount := db.GetTaskCountFromAdminIDs(adminIDs, isFinish)
	if taskCount > 0 {
		var waitGroup sync.WaitGroup

		// 获取所需Tasks
		tasks = db.GetTasksFromAdminIDs(adminIDs, isFinish, offset, countsPerPage)

		// 根据AdminID获取所在org/office名称
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			c.writeOrgOfficeNamesFromAdminIDs(tasks)
		}()

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			if !isFinish { // 执行中任务， 计算status, detail信息
				c.calTasksStatus(tasks)
			} else { // 已完成任务, 计算response_count, accept_count, check_count信息
				c.calTasksFinished(tasks)
			}
		}()
		waitGroup.Wait()
	}
	return tasks, taskCount
}

// 根据AdminID写入其所在org/office名称
func (c Controller) writeOrgOfficeNamesFromAdminIDs(tasks []model.TaskInfo) {
	var waitGroup sync.WaitGroup

	for i := range tasks {
		waitGroup.Add(1)

		go func(i int) {
			defer waitGroup.Done()

			// 根据AdminID获取Admin类型
			isOffice, _ := db.GetAdminType(tasks[i].AdminID)

			// 根据AdminID、Admin类型获取所在组织/单位名称
			if isOffice {
				tasks[i].Launcher = db.GetOfficeOrgNameFromAdmin(tasks[i].AdminID, true)
			} else {
				tasks[i].Launcher = db.GetOfficeOrgNameFromAdmin(tasks[i].AdminID, false)
			}
		}(i)
	}
	waitGroup.Wait()
}

// 计算进行中任务的状态, 状态详情
func (c Controller) calTasksStatus(tasks []model.TaskInfo) {
	// waitGroup 等待下面所有goroutine执行完毕
	var waitGroup sync.WaitGroup

	for i := range tasks {
		waitGroup.Add(1)

		go func(i int) {
			defer waitGroup.Done()
			// 获取该任务的接受人数
			acceptCount := db.GetAcceptCountsFromTask(tasks[i].ID)
			// 把该任务的gather_time转为time.Time
			gatherTime, _ := time.Parse("2006-01-02 15:04:05", tasks[i].GatherTime)

			// 判断任务的状态, 征集、集合、执行
			if gatherTime.Unix() <= time.Now().Unix() { // 执行中,已经过了集合时间, 即 集合时间 < NOW()
				tasks[i].Status = "zx"
			} else { // 可知以下情况都是 还未到集合时间, 即 集合时间 > NOW()
				if acceptCount < tasks[i].MemCount { // 征集中"zj", 接受人数 < 目标人数 && 集合时间 > NOW()
					tasks[i].Status = "zj"
					tasks[i].StatusDetail = float32(acceptCount) / float32(tasks[i].MemCount)
				} else { // 集合中, 接受人数 > 目标人数
					tasks[i].Status = "jh"
					checkCount := db.GetCheckCountsFromTask(tasks[i].ID)
					tasks[i].StatusDetail = float32(checkCount) / float32(acceptCount)
				}
			}
		}(i)
	}

	waitGroup.Wait()
}

// 计算已完成任务的response_count, accept_count, check_count
func (c Controller) calTasksFinished(tasks []model.TaskInfo) {
	// waitGroup 等待下面所有goroutine执行完毕
	var waitGroup sync.WaitGroup

	for i := range tasks {
		waitGroup.Add(1)
		go func(i int) {
			defer waitGroup.Done()
			tasks[i].AcCount = db.GetAcceptCountsFromTask(tasks[i].ID)
			tasks[i].CheckCount = db.GetCheckCountsFromTask(tasks[i].ID)
			refuseCount := db.GetRefuseCountsFromTask(tasks[i].ID)
			tasks[i].RespCount = tasks[i].AcCount + refuseCount
		}(i)
	}

	waitGroup.Wait()
}
