package control

import (
	"math"
	"model"
	"sync"
	"time"
)

// Controller 全局控制层管理器
type Controller struct{}

var db model.DBManager

func (c Controller) CreateTask(task *model.Task, place *model.Place, acmem *model.AcMem) error {
	err := db.CreateTask(task, place, acmem)
	return err
}

func (c Controller) EndTask(taskID, adminID int) error {
	err := db.EndTask(taskID, adminID)
	return err
}

// GetCommonPlaces 根据AdminID、Admin类型获取其常用地点
func (c Controller) GetCommonPlaces(adminID int) ([]model.Place, bool, error) {
	isOffice, err := db.GetAdminType(adminID)
	if err != nil {
		return nil, false, err
	}
	places := db.GetCommonPlaces(adminID, isOffice)
	return places, isOffice, nil
}

// GetOfficeInfoAndMems 根据AdminID获取单位、下属单位及成员
func (c Controller) GetOfficeInfoAndMems(adminID int) (*model.OfficeInfo, error) {
	officeID := db.GetOfficeIDFromAdminID(adminID)

	var officeInfo model.OfficeInfo
	officeDetail, memCounts, err := c.getOfficeDetail(officeID)
	if err != nil {
		return nil, err
	}

	officeInfo.OfficeDetail = officeDetail
	officeInfo.TotalMems = memCounts
	return &officeInfo, err
}

// 递归, 获取下属单位、人员及人数
func (c Controller) getOfficeDetail(officeID int) (model.Office, int, error) {
	office := model.Office{ID: officeID, LowerOffs: make([]model.Office, 0)}

	// 根据OfficeID获取单位名称
	office.Name = db.GetOfficeName(officeID)
	// 根据OfficeID获取所含民兵及人数
	soldiers, memCounts := db.GetOfficeMems(officeID)
	office.Members = soldiers

	// 获取该单位的下属单位
	lowerOffIDs := db.GetLowerOfficeIDsFromOffices([]int{officeID})
	for _, lowerOffID := range lowerOffIDs {
		lowerOffice, counts, err := c.getOfficeDetail(lowerOffID)
		if err != nil {
			return office, 0, err
		}

		office.LowerOffs = append(office.LowerOffs, lowerOffice)
		memCounts += int64(counts)
	}

	return office, int(memCounts), nil
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
			c.WriteOrgOfficeNamesFromAdminIDs(tasks)
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

// GetTaskDetail 获取任务详情
func (c Controller) GetTaskDetail(taskID int, watchAdminID int) (*model.TaskInfo, error) {
	task := make([]model.TaskInfo, 1)
	// 任务title, launch_datetime, gather_datetime, place_name等
	task[0] = db.GetTaskDetailFromDB(taskID)
	// 任务的launcher（发起任务的组织/单位名称）
	c.WriteOrgOfficeNamesFromAdminIDs(task)
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

// 根据AdminID写入其所在org/office名称
func (c Controller) WriteOrgOfficeNamesFromAdminIDs(tasks []model.TaskInfo) {
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

// GetOrgInfoAndMems 获取下属组织及成员
func (c Controller) GetOrgInfoAndMems(adminID int) (*model.OrgInfo, error) {
	orgDetail := model.OrgDetail{Orgs: make([]model.Org, 0), LowerOffices: make([]model.OrgDetail, 0)}
	orgInfo := model.OrgInfo{}

	isOffice, _ := db.GetAdminType(adminID)
	if isOffice { // 单位
		// 获取OfficeID
		officeID := db.GetOfficeIDFromAdminID(adminID)
		// 获取OrgDetail
		orgDetail, uniqueSoldiers := c.getOrgDetail(officeID)
		// 获取Total Members
		orgInfo.TotalMems = len(uniqueSoldiers)
		orgInfo.Orgdetail = orgDetail
	} else { // 组织
		// 通过AdminID获取其所在Office的名称
		orgDetail.OfficeName = db.GetOfficeOrgNameFromAdmin(adminID, false)
		// 通过AdminID获取其所在Org名称
		orgID := db.GetOrgIDFromAdminID(adminID)
		// 通过OrgID获取组织、下属组织的信息，如OrgName，成员，成员数量
		orgs, memCount := c.getOrgAndAllLowerOrgs(orgID)
		orgDetail.Orgs = orgs

		orgInfo.Orgdetail = orgDetail
		orgInfo.TotalMems = memCount
	}
	return &orgInfo, nil
}

type safeMap struct {
	lock           sync.Mutex
	uniqueSoldrIDs map[int]bool
}

// 根据OfficeID获取OrgInfo.OrgDetail
func (c Controller) getOrgDetail(officeID int) (model.OrgDetail, map[int]bool) {
	uniqueSoldiers := safeMap{uniqueSoldrIDs: make(map[int]bool)}
	var waitGroup sync.WaitGroup
	orgDetail := model.OrgDetail{Orgs: make([]model.Org, 0), LowerOffices: make([]model.OrgDetail, 0)}

	// 获取officeName
	orgDetail.OfficeName = db.GetOfficeName(officeID)
	// 获取Orgs，Orgs的不重复SoldierIDs
	orgIDs := db.GetOrgIDsFromOffices([]int{officeID})
	if len(orgIDs) > 0 {
		orgs, subUniqueSoldrs := c.getOrgs(orgIDs)
		orgDetail.Orgs = orgs
		// 插入Unique SoldierIDs
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			uniqueSoldiers.lock.Lock()
			defer uniqueSoldiers.lock.Unlock()

			for soldierID := range subUniqueSoldrs {
				uniqueSoldiers.uniqueSoldrIDs[soldierID] = true
			}
		}()
	}
	// 获取LowerOffices
	lowerOfficeIDs := db.GetLowerOfficeIDsFromOffices([]int{officeID})
	if len(lowerOfficeIDs) > 0 {
		var lowerOfficesLock sync.Mutex
		for _, lowerOfficeID := range lowerOfficeIDs {
			waitGroup.Add(1)
			go func(officeID int) {
				defer waitGroup.Done()

				// 根据LowerOfficeIDs获取对应OrgDetails（OrgDetail.LowerOffices）
				lowerOrgDetail, subUniqueSoldrs := c.getOrgDetail(officeID)
				// 插入lowerOffices
				waitGroup.Add(1)
				go func() {
					defer waitGroup.Done()
					lowerOfficesLock.Lock()
					defer lowerOfficesLock.Unlock()

					orgDetail.LowerOffices = append(orgDetail.LowerOffices, lowerOrgDetail)
				}()
				// 插入UniqueSoldiers
				waitGroup.Add(1)
				go func() {
					defer waitGroup.Done()
					uniqueSoldiers.lock.Lock()
					defer uniqueSoldiers.lock.Unlock()

					for soldierID := range subUniqueSoldrs {
						uniqueSoldiers.uniqueSoldrIDs[soldierID] = true
					}
				}()
			}(lowerOfficeID)
		}
	}
	waitGroup.Wait()
	return orgDetail, uniqueSoldiers.uniqueSoldrIDs
}

// 通过OrgIDs获取Orgs（name, members, lowerOrgIDs）
func (c Controller) getOrgs(orgIDs []int) ([]model.Org, map[int]bool) {
	orgs := make([]model.Org, len(orgIDs))
	uniqueSoldiers := safeMap{uniqueSoldrIDs: make(map[int]bool)}
	var waitGroup sync.WaitGroup

	for i := range orgs {
		waitGroup.Add(1)
		go func(i int) {
			defer waitGroup.Done()
			// 根据OrgID获取组织信息
			orgs[i] = c.getOrg(orgIDs[i])
			// 获取不重复的SoldierID
			uniqueSoldiers.lock.Lock()
			defer uniqueSoldiers.lock.Unlock()
			for _, soldier := range orgs[i].Members {
				uniqueSoldiers.uniqueSoldrIDs[soldier.ID] = true
			}
		}(i)
	}
	waitGroup.Wait()
	return orgs, uniqueSoldiers.uniqueSoldrIDs
}

// 通过OrgID获取Org（name, 成员, 下属OrgIDs）
func (c Controller) getOrg(orgID int) model.Org {
	org := model.Org{ID: orgID}
	org.Name = db.GetOrgName(orgID)
	org.Members = c.getOrgMemsAndAdmins(orgID)
	org.LowerOrgIDs = db.GetLowerOrgIDsFromOrgIDs([]int{orgID})
	return org
}

// 通过OrgID获取组织及其所有下属的名称、成员
// 返回该组织及所有下属组织（的信息），成员数量（成员不重复）
func (c Controller) getOrgAndAllLowerOrgs(orgID int) ([]model.Org, int) {
	orgs := make([]model.Org, 0)
	// 记录该组织及其下属组织的成员数量，人员不重复
	uniqueSoldiers := safeMap{uniqueSoldrIDs: make(map[int]bool)}
	var waitGroup sync.WaitGroup

	// 使用queue找到orgID的所有下属Org，及其name, members, lowerOrgIDs
	queue := make([]int, 1)
	queue[0] = orgID
	for len(queue) != 0 {
		orgID := queue[0]
		org := c.getOrg(orgID)
		orgs = append(orgs, org)

		queue = append(queue, org.LowerOrgIDs...)
		queue = queue[1:] // queue.pop()

		// 记录不重复的民兵
		waitGroup.Add(1)
		go func(members []model.Soldier) {
			defer waitGroup.Done()
			uniqueSoldiers.lock.Lock()
			defer uniqueSoldiers.lock.Unlock()
			for _, soldier := range members {
				uniqueSoldiers.uniqueSoldrIDs[soldier.ID] = true
			}
		}(org.Members)
	}
	waitGroup.Wait()
	memCount := len(uniqueSoldiers.uniqueSoldrIDs)
	return orgs, memCount
}

// 通过OrgID获取其成员
func (c Controller) getOrgMemsAndAdmins(orgID int) []model.Soldier {
	// 获取是管理员的民兵ID
	isSoldierAdmin := make(map[int]bool) // map[soldierID]isAdmin
	soldierIDs := db.GetSoldrIDsWhoAreAdmins(orgID)
	for _, soldierID := range soldierIDs {
		isSoldierAdmin[soldierID] = true
	}

	soldiers := db.GetOrgMems(orgID)
	// 判断每个Soldier是否为Admin
	for i := range soldiers {
		if isSoldierAdmin[soldiers[i].ID] {
			soldiers[i].IsAdmin = true
		}
	}
	return soldiers
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
