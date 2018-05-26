package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.Debug = true
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", "mbcsdev:mbcsdev2018@tcp(222.200.180.59:9000)/MBDB?charset=utf8")
	//orm.RegisterDataBase("default", "mysql", "root:root@tcp(127.0.0.1:3306)/mb?charset=utf8")
}

// DBManager 数据层的管理器
type DBManager struct{}

func (db DBManager) GetAdminType(adminID int) (isOff bool, err error) {
	o := orm.NewOrm()
	var adminType string
	err = o.Raw("SELECT admin_type FROM Admins WHERE admin_id = ?", adminID).QueryRow(&adminType)
	if err != nil {
		return false, err
	}

	switch adminType {
	case "OF": // office
		return true, nil
	case "OR": // organization
		return false, nil
	default:
		return false, errors.New("error: unknown type of Admin in table 'Admins'")
	}
}

// CreateTask 在表Tasks创建新任务，并在表GatherNotifications创建新的集合通知
func (db DBManager) CreateTask(task *Task, place *Place, acmem *AcMem) error {
	o := orm.NewOrm()
	o.Begin()

	// 创建新任务Tasks，若地点是新的，则创建新地点Places
	err := db.insertTask(&o, task, place)
	if err != nil {
		o.Rollback()
		return err
	}
	// 创建新的AcceptMembers，即插入TaskAcceptOffices, TaskAcceptOrgs, GatherNotifications
	err = db.insertAcMem(&o, task.ID, acmem)
	if err != nil {
		o.Rollback()
		return err
	}
	o.Commit()

	// TODO: 广播：模板消息、短信

	return nil
}

func (db DBManager) insertPlace(o *orm.Ormer, task *Task, isOffice bool, place *Place) (int, error) {
	// 插入Places
	rawSQL := "INSERT INTO Places(place_name, place_lat, place_lng) VALUES(?,?,?)"
	result, err := (*o).Raw(rawSQL, place.Name, place.Lat, place.Lng).Exec()
	if err != nil {
		return -1, err
	}
	placeID, _ := result.LastInsertId()

	// 关联 常用地点 与 组织/单位
	if isOffice { // 插入OfficePlaces
		officeID := db.getOfficeIDFromAdminID(task.AdminID) // 通过adminID获取officeID
		rawSQL = "INSERT INTO OfficePlaces (office_id, place_id) VALUES (?, ?)"
		(*o).Raw(rawSQL, officeID, placeID).Exec()
	} else { // 插入OrgPlaces
		orgID := db.getOrgIDFromAdminID(task.AdminID) // 通过adminID获取orgID
		rawSQL = "INSERT INTO OrgPlaces (org_id, place_id) VALUES (?, ?)"
		(*o).Raw(rawSQL, orgID, placeID).Exec()
	}

	return int(placeID), nil
}

func (db DBManager) getOrgIDFromAdminID(adminID int) int {
	var orgID int
	o := orm.NewOrm()
	o.Raw("SELECT org_id FROM OrgAdminRelationships WHERE admin_id = ?", adminID).QueryRow(&orgID)
	return orgID
}

func (db DBManager) getOfficeIDFromAdminID(adminID int) int {
	var officeID int
	o := orm.NewOrm()
	o.Raw("SELECT office_id FROM OfficeAdminRelationships WHERE admin_id = ?", adminID).QueryRow(&officeID)
	return officeID
}

func (db DBManager) insertTask(o *orm.Ormer, task *Task, place *Place) error {
	var err error
	if task.PlaceID == -1 { // 需要新建单位/组织的常用地点
		isOffice, err := db.GetAdminType(task.AdminID)
		if err != nil {
			return err
		}
		task.PlaceID, err = db.insertPlace(o, task, isOffice, place)
		if err != nil {
			return err
		}
	}

	rawSQL := "INSERT INTO Tasks"
	rawSQL += "(title, mem_count, launch_admin_id, gather_datetime, detail, gather_place_id, finish_datetime)"
	rawSQL += "VALUES(?,?,?,?,?,?,?)"
	result, err := (*o).Raw(rawSQL, task.Title, task.Count, task.AdminID, task.Gather, task.Detail, task.PlaceID, task.Finish).Exec()
	if err != nil {
		return err
	}
	taskID, _ := result.LastInsertId()
	task.ID = int(taskID)
	return nil
}

func (db DBManager) insertAcMem(o *orm.Ormer, taskID int, acmem *AcMem) error {
	taskIDStr := strconv.Itoa(taskID)

	// 批量插入“接受任务的单位”
	rawSQL := "INSERT INTO TaskAcceptOffices(ac_task_id, ac_office_id) VALUES"
	for _, acOffID := range acmem.AcOffIDs {
		rawSQL += "(" + taskIDStr + "," + strconv.Itoa(acOffID) + "),"
	}
	_, err := (*o).Raw(rawSQL[:len(rawSQL)-1]).Exec()
	if err != nil {
		return err
	}
	// 插入“接受任务的组织”
	rawSQL = "INSERT INTO TaskAcceptOrgs(ac_task_id, ac_org_id) VALUES"
	for _, acOrgID := range acmem.AcOrgIDs {
		rawSQL += "(" + taskIDStr + "," + strconv.Itoa(acOrgID) + "),"
	}
	_, err = (*o).Raw(rawSQL[:len(rawSQL)-1]).Exec()
	if err != nil {
		return err
	}

	uniqueSoldrIDs := make(map[int]bool) // 从单位、组织、个人中选取的所有民兵ID，因为人员可能有重复，故用map消重
	var soldierIDs []int
	// 从单位ID获取民兵ID
	soldierIDs = db.getSoldrIDsFromOfficeIDs(acmem.AcOffIDs)
	for _, soldrID := range soldierIDs {
		uniqueSoldrIDs[soldrID] = true
	}

	// 从组织ID获取民兵ID
	soldierIDs = db.getSoldrIDsFromOrgIDs(acmem.AcOrgIDs)
	for _, soldrID := range soldierIDs {
		uniqueSoldrIDs[soldrID] = true
	}

	// 获取单独被选取的民兵ID
	for _, soldrID := range acmem.AcSoldIDs {
		uniqueSoldrIDs[soldrID] = true
	}

	// 批量插入 GatherNotifications
	rawSQL = "INSERT INTO GatherNotifications(gather_task_id, recv_soldier_id, read_status) VALUES"
	for soldierID, ok := range uniqueSoldrIDs {
		if ok {
			rawSQL += "(" + taskIDStr + "," + strconv.Itoa(soldierID) + ",'UR')," // readStatus: UR(未读状态)
		}
	}
	_, err = (*o).Raw(rawSQL[:len(rawSQL)-1]).Exec()
	if err != nil {
		return err
	}
	return nil

	// TODO: 对所有 uniqueSoldrIDs 进行广告（模板消息、短信）操作
}

// 通过单位IDs获取所有民兵ID
func (db DBManager) getSoldrIDsFromOfficeIDs(officeIDs arrayInt) []int {
	var soldierIDs []int
	rawSQL := "SELECT DISTINCT(soldier_id) FROM Soldiers WHERE serve_office_id IN "
	//rawSQL := "SELECT soldier_id FROM Soldiers WHERE serve_office_id IN "
	rawSQL += fmt.Sprint(officeIDs)
	o := orm.NewOrm()
	o.Raw(rawSQL).QueryRows(&soldierIDs)
	return soldierIDs
}

// 通过组织IDs获取所有民兵ID
func (db DBManager) getSoldrIDsFromOrgIDs(orgIDs arrayInt) []int {
	var soldierIDs []int
	rawSQL := "SELECT DISTINCT(soldier_id) FROM OrgSoldierRelationships WHERE serve_org_id IN "
	//rawSQL := "SELECT soldier_id FROM OrgSoldierRelationships WHERE serve_org_id IN "
	rawSQL += fmt.Sprint(orgIDs)
	o := orm.NewOrm()
	o.Raw(rawSQL).QueryRows(&soldierIDs)
	return soldierIDs
}

type arrayInt []int

// 格式: (1,2,3,...,n)
func (num arrayInt) String() string {
	arrayToStr := "("
	arrayLen := len(num)
	for i, n := range num {
		if i == arrayLen-1 {
			arrayToStr += strconv.Itoa(n)
		} else {
			arrayToStr += strconv.Itoa(n) + ","
		}
	}
	arrayToStr += ")"
	return arrayToStr
}

// EndTask 结束任务, 把任务结束时间设为当前时间
func (db DBManager) EndTask(taskID, adminID int) error {
	o := orm.NewOrm()
	_, err := o.Raw("UPDATE Tasks SET finish_datetime = NOW() WHERE task_id = ?", taskID).Exec()
	if err != nil {
		return err
	}
	return nil
}

// GetCommonPlaces 根据AdminID与Admin类型isOffice, 查找Admin对应的组织/单位的常用地点.
// 查找顺序: admin_id -> org_id/office_id -> place_id -> all places
func (db DBManager) GetCommonPlaces(adminID int, isOffice bool) ([]Place, error) {
	var places []Place
	o := orm.NewOrm()
	rawSQL := "SELECT * FROM Places "
	rawSQL += "WHERE place_id IN ( "
	rawSQL += "SELECT place_id "
	if isOffice { // 如果Admin类型是Office, 则从AdminID找出Office
		rawSQL += "FROM OfficePlaces "
		rawSQL += "WHERE office_id IN ( "
		rawSQL += "SELECT office_id FROM OfficeAdminRelationships "
	} else { // 如果Admin类型是Org, 则从AdminID找出Org
		rawSQL += "FROM OrgPlaces "
		rawSQL += "WHERE org_id IN ( "
		rawSQL += "SELECT org_id FROM OrgAdminRelationships "
	}
	rawSQL += "WHERE admin_id = ?)"
	rawSQL += ")"
	_, err := o.Raw(rawSQL, adminID).QueryRows(&places)
	fmt.Println(places)
	return places, err
}

// GetOfficeInfoAndMems 根据AdminID获取单位、下属单位及成员
func (db DBManager) GetOfficeInfoAndMems(adminID int) (*OfficeInfo, error) {
	officeID := db.getOfficeIDFromAdminID(adminID)

	var officeInfo OfficeInfo
	officeDetail, memCounts, err := db.getOfficeDetail(officeID)
	if err != nil {
		return nil, err
	}

	officeInfo.OfficeDetail = officeDetail
	officeInfo.TotalMems = memCounts
	return &officeInfo, err
}

// 递归, 获取下属单位、人员及人数
func (db DBManager) getOfficeDetail(officeID int) (Office, int, error) {
	office := Office{ID: officeID, LowerOffs: make([]Office, 0)}

	// 根据OfficeID获取单位名称
	o := orm.NewOrm()
	o.Raw("SELECT name FROM Offices WHERE office_id = ?", officeID).QueryRow(&(office.Name))

	// 根据OfficeID获取所含民兵及人数
	rawSQL := "SELECT soldier_id, name FROM Soldiers WHERE serve_office_id = ?"
	memCounts, _ := o.Raw(rawSQL, officeID).QueryRows(&(office.Members))

	// 获取该单位的下属单位
	var lowerOffIDs []int
	rawSQL = "SELECT lower_office_id FROM OfficeRelationships WHERE higher_office_id = ?"
	o.Raw(rawSQL, officeID).QueryRows(&lowerOffIDs)
	for _, lowerOffID := range lowerOffIDs {
		lowerOffice, counts, err := db.getOfficeDetail(lowerOffID)
		if err != nil {
			return office, 0, err
		}

		office.LowerOffs = append(office.LowerOffs, lowerOffice)
		memCounts += int64(counts)
	}

	return office, int(memCounts), nil
}

// 获取从属于该Office的Organizations ID
func (db DBManager) getOrgIDsFromOffices(officeIDs arrayInt) arrayInt {
	var orgIDs arrayInt
	o := orm.NewOrm()
	rawSQL := "SELECT org_id FROM Organizations WHERE serve_office_id IN " + fmt.Sprint(officeIDs)
	o.Raw(rawSQL).QueryRows(&orgIDs)
	return orgIDs
}

// 获取这些单位的下属单位
func (db DBManager) getLowerOfficeIDsFromOffices(officeIDs arrayInt) arrayInt {
	var lowerOfficeIDs arrayInt
	o := orm.NewOrm()
	rawSQL := "SELECT lower_office_id FROM OfficeRelationships WHERE higher_office_id IN " + fmt.Sprint(officeIDs)
	o.Raw(rawSQL).QueryRows(&lowerOfficeIDs)
	return lowerOfficeIDs
}

// 通过OfficeIDs查找出其对应的AdminIDs
func (db DBManager) getAdminIDsFromOffices(officeIDs arrayInt) []int {
	var adminIDs []int
	o := orm.NewOrm()
	rawSQL := "SELECT admin_id FROM OfficeAdminRelationships WHERE office_id IN " + fmt.Sprint(officeIDs)
	o.Raw(rawSQL).QueryRows(&adminIDs)
	return adminIDs
}

// 从单位获取其下属单位和组织(除去其本身和其所含组织)的AdminIDs
// 递归
func (db DBManager) getAdminIDsInAllLowerOfficesAndOrgs(officeIDs arrayInt, adminIDs []int) {
	// 获取目前单位的所有下属单位、AdminIDs
	lowerOffIDs := db.getLowerOfficeIDsFromOffices(officeIDs)

	if len(lowerOffIDs) >= 0 {
		adminIDsFromOffices := db.getAdminIDsFromOffices(lowerOffIDs)

		// 从所有下属单位获取他们所含的组织、AdminIDs
		lowerOrgIDs := db.getOrgIDsFromOffices(lowerOffIDs)
		adminIDsFromOrgs := db.getAdminIDsFromOrgs(lowerOrgIDs)

		adminIDs = append(adminIDs, adminIDsFromOffices...)
		adminIDs = append(adminIDs, adminIDsFromOrgs...)

		db.getAdminIDsInAllLowerOfficesAndOrgs(lowerOffIDs, adminIDs)
	}
}

// GetTaskList 获取任务列表(执行中, 已完成). 区分单位/组织, 每页显示数目, 当前页
func (db DBManager) GetTaskList(adminID, countsPerPage, offset int, isFinish, isOffice bool) (*List, error) {
	tasklist := List{}
	adminIDs := arrayInt{adminID}

	// 获取Amin下属组织/单位的AdminID
	if isOffice { // Admin类型为单位
		officeID := db.getOfficeIDFromAdminID(adminID)
		orgIDs := db.getOrgIDsFromOffices(arrayInt{officeID})
		subAdminIDs := db.getAdminIDsFromOrgs(orgIDs)
		adminIDs = append(adminIDs, subAdminIDs...)
		db.getAdminIDsInAllLowerOfficesAndOrgs(arrayInt{officeID}, adminIDs)
	} else { // Admin类型为组织
		orgID := db.getOrgIDFromAdminID(adminID)
		db.getAdminIDsInAllLowerOrgs(arrayInt{orgID}, adminIDs)
	}
	// 从AdminIDs获取Tasks和Tasks总数
	tasks, taskCount := db.getTasksAndCountsFromAdmin(adminIDs, isFinish, offset, countsPerPage)

	tasklist.TaskCount = taskCount
	tasklist.PageCount = int(math.Ceil(float64(taskCount) / float64(countsPerPage)))
	tasklist.Tasks = tasks
	return &tasklist, nil
}

// 通过AdminIDs获取他们发布过的任务数量, 分类为"执行中"和"已完成"
func (db DBManager) getTaskCountFromAdminIDs(adminIDs arrayInt, isFinish bool) int {
	var taskCount int
	o := orm.NewOrm()
	rawSQL := "SELECT COUNT(*) FROM Tasks WHERE launch_admin_id IN " + fmt.Sprint(adminIDs)
	if isFinish {
		rawSQL += " AND finish_datetime <= NOW()"
	} else {
		rawSQL += " AND finish_datetime > NOW()"
	}
	o.Raw(rawSQL).QueryRow(&taskCount)
	return taskCount
}

// getTasksFromAdminIDs 通过AdminIDs获取他们发布过的任务, 分类为"执行中"和"已完成"
// 获取“执行中”“已完成”Tasks所需信息的交集， 但不包括发起单位、组织， 集合地点名称的信息
func (db DBManager) getTasksFromAdminIDs(adminIDs arrayInt, isFinish bool, offset, countsPerPage int) []Tasklist {
	var tasks []Tasklist

	o := orm.NewOrm()
	rawSQL := "SELECT task_id, title, mem_count, launch_admin_id, launch_datetime, gather_place_id "
	if !isFinish { // 执行中的任务需要有gather_datetime. 已完成任务就不需要
		rawSQL += ", gather_datetime "
	}
	rawSQL += "FROM Tasks "
	rawSQL += "WHERE launch_admin_id IN " + fmt.Sprint(adminIDs)
	if isFinish {
		rawSQL += " AND finish_datitime <= NOW() "
	} else {
		rawSQL += " AND finish_datetime > NOW() "
	}
	rawSQL += "LIMIT ? OFFSET ?"
	o.Raw(rawSQL, countsPerPage, offset).QueryRows(&tasks)

	return tasks
}

// 根据[]Tasklist中的每个placeID获取placeName
func (db DBManager) writePlaceNamesFromPlaceIDs(tasks []Tasklist) {
	o := orm.NewOrm()
	rawSQL := "SELECT place_name FROM Places WHERE place_id = ?"
	for i := range tasks {
		o.Raw(rawSQL, tasks[i].PlaceID).QueryRow(&(tasks[i].Place))
	}
}

// 根据AdminID写入其所在org/office名称
func (db DBManager) writeOrgOfficeNamesFromAdminIDs(tasks []Tasklist) {
	rawSQL := ""
	adminType := ""
	var waitGroup sync.WaitGroup

	for i := range tasks {
		waitGroup.Add(1)

		go func(i int) {
			defer waitGroup.Done()

			// 根据AdminID获取Admin类型
			o := orm.NewOrm()
			rawSQL = "SELECT admin_type FROM Admins WHERE admin_id = ?"
			o.Raw(rawSQL, tasks[i].AdminID).QueryRow(&adminType)

			// 根据AdminID、Admin类型获取所在组织/单位名称
			if adminType == "OF" {
				rawSQL = "SELECT name FROM Offices WHERE office_id = ("
				rawSQL += "SELECT office_id FROM OfficeAdminRelationships WHERE admin_id = ?)"
			} else {
				rawSQL = "SELECT name FROM Organizations WHERE org_id = ("
				rawSQL += "SELECT org_id FROM OrgAdminRelationships WHERE admin_id = ?)"
			}
			o.Raw(rawSQL, tasks[i].AdminID).QueryRow(&(tasks[i].Launcher))
		}(i)
	}
	waitGroup.Wait()
}

// 从AdminIDs获取他们发布的任务[]Tasklist(根据偏移量和所需数量), 他们发布任务的总数int
func (db DBManager) getTasksAndCountsFromAdmin(adminIDs arrayInt, isFinish bool, offset, countsPerPage int) ([]Tasklist, int) {
	// 获取Tasks的数量
	taskCount := db.getTaskCountFromAdminIDs(adminIDs, isFinish)
	// 获取所需Tasks
	tasks := db.getTasksFromAdminIDs(adminIDs, isFinish, offset, countsPerPage)
	// 根据[]Tasklist中的每个placeID获取placeName
	db.writePlaceNamesFromPlaceIDs(tasks)
	// 根据AdminID获取所在org/office名称
	db.writeOrgOfficeNamesFromAdminIDs(tasks)

	if !isFinish { // 执行中任务， 计算status, detail信息
		db.calTasksStatus(tasks)
	} else { // 已完成任务, 计算response_count, accept_count, check_count信息
		db.calTasksFinished(tasks)
	}

	return tasks, taskCount
}

// 计算已完成任务的response_count, accept_count, check_count
func (db DBManager) calTasksFinished(tasks []Tasklist) {
	// waitGroup 等待下面所有goroutine执行完毕
	var waitGroup sync.WaitGroup

	for i := range tasks {
		waitGroup.Add(1)
		go func(i int) {
			defer waitGroup.Done()
			tasks[i].AcCount = db.getAcceptCountsFromTask(tasks[i].ID)
			tasks[i].CheckCount = db.getCheckCountsFromTask(tasks[i].ID)
			refuseCount := db.getRefuseCountsFromTask(tasks[i].ID)
			tasks[i].RespCount = tasks[i].AcCount + refuseCount
		}(i)
	}

	waitGroup.Wait()
}

// 计算进行中任务的状态, 状态详情
func (db DBManager) calTasksStatus(tasks []Tasklist) {
	// waitGroup 等待下面所有goroutine执行完毕
	var waitGroup sync.WaitGroup

	for i := range tasks {
		waitGroup.Add(1)

		go func(i int) {
			defer waitGroup.Done()
			// 获取该任务的接受人数
			acceptCount := db.getAcceptCountsFromTask(tasks[i].ID)
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
					checkCount := db.getCheckCountsFromTask(tasks[i].ID)
					tasks[i].StatusDetail = float32(checkCount) / float32(acceptCount)
				}
			}
		}(i)
	}

	waitGroup.Wait()
}

func (db DBManager) getCheckCountsFromTask(taskID int) int {
	var count int
	o := orm.NewOrm()
	o.Raw("SELECT COUNT(*) FROM TaskCheckNames WHERE check_task_id = ?", taskID).QueryRow(&count)
	return count
}

func (db DBManager) getAcceptCountsFromTask(taskID int) int {
	var count int
	o := orm.NewOrm()
	rawSQL := "SELECT COUNT(*) FROM GatherNotifications WHERE gather_task_id = ? AND read_status = 'AC'"
	o.Raw(rawSQL, taskID).QueryRow(&count)
	return count
}

func (db DBManager) getRefuseCountsFromTask(taskID int) int {
	var count int
	o := orm.NewOrm()
	rawSQL := "SELECT COUNT(*) FROM GatherNotifications WHERE read_status = 'RF' AND gather_task_id = ?"
	o.Raw(rawSQL, taskID).QueryRow(&count)
	return count
}

// 给定highOrgIDs, 获取其所有下属(除去其本身)组织的AdminIDs
// 对组织的遍历, BFS
func (db DBManager) getAdminIDsInAllLowerOrgs(highOrgIDs arrayInt, adminIDs []int) {
	// 从该OrgID获取下属OrgIDs
	lowerOrgIDs := db.getLowerOrgIDsFromOrgIDs(highOrgIDs)
	// 从下属Organizations获取所有AdminIDs
	subAdminIDs := db.getAdminIDsFromOrgs(lowerOrgIDs)

	if len(subAdminIDs) > 0 {
		adminIDs = append(adminIDs, subAdminIDs...)
		// 从下属OrgIDs获取所有AdminIDs
		db.getAdminIDsInAllLowerOrgs(lowerOrgIDs, adminIDs)
	}
}

// 从OrgIDs获取其所有AdminIDs
func (db DBManager) getAdminIDsFromOrgs(orgIDs arrayInt) []int {
	var adminIDs []int
	o := orm.NewOrm()
	rawSQL := "SELECT admin_id FROM OrgAdminRelationships WHERE org_id IN " + fmt.Sprint(orgIDs)
	o.Raw(rawSQL).QueryRows(&adminIDs)
	return adminIDs
}

// 从OrgIDs获取下属OrgIDs
func (db DBManager) getLowerOrgIDsFromOrgIDs(orgIDs arrayInt) []int {
	var lowerOrgIDs []int
	o := orm.NewOrm()
	rawSQL := "SELECT lower_org_id FROM OrgRelationships WHERE higher_org_id IN " + fmt.Sprint(orgIDs)
	o.Raw(rawSQL).QueryRows(&lowerOrgIDs)
	return lowerOrgIDs
}
