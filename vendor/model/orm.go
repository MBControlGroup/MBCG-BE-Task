package model

import (
	"errors"
	"fmt"
	"strconv"
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
		officeID := db.GetOfficeIDFromAdminID(task.AdminID) // 通过adminID获取officeID
		rawSQL = "INSERT INTO OfficePlaces (office_id, place_id) VALUES (?, ?)"
		(*o).Raw(rawSQL, officeID, placeID).Exec()
	} else { // 插入OrgPlaces
		orgID := db.GetOrgIDFromAdminID(task.AdminID) // 通过adminID获取orgID
		rawSQL = "INSERT INTO OrgPlaces (org_id, place_id) VALUES (?, ?)"
		(*o).Raw(rawSQL, orgID, placeID).Exec()
	}

	return int(placeID), nil
}

func (db DBManager) GetOrgIDFromAdminID(adminID int) int {
	var orgID int
	o := orm.NewOrm()
	o.Raw("SELECT org_id FROM OrgAdminRelationships WHERE admin_id = ?", adminID).QueryRow(&orgID)
	return orgID
}

// GetOfficeIDFromAdminID 通过AdminID获取其所在OfficeID
func (db DBManager) GetOfficeIDFromAdminID(adminID int) int {
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
	rawSQL := ""

	// 批量插入“接受任务的单位”
	if len(acmem.AcOffIDs) > 0 {
		rawSQL = "INSERT INTO TaskAcceptOffices(ac_task_id, ac_office_id) VALUES "
		for _, acOffID := range acmem.AcOffIDs {
			rawSQL += "(" + taskIDStr + "," + strconv.Itoa(acOffID) + "),"
		}
		_, err := (*o).Raw(rawSQL[:len(rawSQL)-1]).Exec()
		if err != nil {
			return err
		}
	}
	// 插入“接受任务的组织”
	if len(acmem.AcOrgIDs) > 0 {
		rawSQL = "INSERT INTO TaskAcceptOrgs(ac_task_id, ac_org_id) VALUES"
		for _, acOrgID := range acmem.AcOrgIDs {
			rawSQL += "(" + taskIDStr + "," + strconv.Itoa(acOrgID) + "),"
		}
		_, err := (*o).Raw(rawSQL[:len(rawSQL)-1]).Exec()
		if err != nil {
			return err
		}
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
	if len(uniqueSoldrIDs) > 0 {
		rawSQL = "INSERT INTO GatherNotifications(gather_task_id, recv_soldier_id, read_status) VALUES "
		for soldierID, ok := range uniqueSoldrIDs {
			if ok {
				rawSQL += "(" + taskIDStr + "," + strconv.Itoa(soldierID) + ",'UR')," // readStatus: UR(未读状态)
			}
		}
		_, err := (*o).Raw(rawSQL[:len(rawSQL)-1]).Exec()
		if err != nil {
			return err
		}
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
func (db DBManager) GetCommonPlaces(adminID int, isOffice bool) []Place {
	places := make([]Place, 0)
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
	o.Raw(rawSQL, adminID).QueryRows(&places)
	return places
}

// GetOfficeMems 通过OfficeID获取成员
func (db DBManager) GetOfficeMems(officeID int, needIMUser bool) ([]Soldier, int64) {
	soldiers := make([]Soldier, 0)
	o := orm.NewOrm()
	rawSQL := "SELECT soldier_id, name "
	if needIMUser {
		rawSQL += ", im_user_id "
	}
	rawSQL += "FROM Soldiers WHERE serve_office_id = ?"
	memCounts, _ := o.Raw(rawSQL, officeID).QueryRows(&soldiers)
	return soldiers, memCounts
}

// GetOrgIDsFromOffices 获取从属于该Office的Organizations ID
func (db DBManager) GetOrgIDsFromOffices(officeIDs arrayInt) []int {
	var orgIDs arrayInt
	o := orm.NewOrm()
	rawSQL := "SELECT org_id FROM Organizations WHERE serve_office_id IN " + fmt.Sprint(officeIDs)
	o.Raw(rawSQL).QueryRows(&orgIDs)
	return orgIDs
}

// 获取这些单位的下属单位
func (db DBManager) GetLowerOfficeIDsFromOffices(officeIDs arrayInt) []int {
	var lowerOfficeIDs arrayInt
	o := orm.NewOrm()
	rawSQL := "SELECT lower_office_id FROM OfficeRelationships WHERE higher_office_id IN " + fmt.Sprint(officeIDs)
	o.Raw(rawSQL).QueryRows(&lowerOfficeIDs)
	return lowerOfficeIDs
}

// 通过OfficeIDs查找出其对应的AdminIDs
func (db DBManager) GetAdminIDsFromOffices(officeIDs arrayInt) []int {
	var adminIDs []int
	o := orm.NewOrm()
	rawSQL := "SELECT admin_id FROM OfficeAdminRelationships WHERE office_id IN " + fmt.Sprint(officeIDs)
	o.Raw(rawSQL).QueryRows(&adminIDs)
	return adminIDs
}

// 通过AdminIDs获取他们发布过的任务数量, 分类为"执行中"和"已完成"
func (db DBManager) GetTaskCountFromAdminIDs(adminIDs arrayInt, isFinish bool) int {
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

// GetTasksFromAdminIDs 通过AdminIDs获取他们发布过的任务, 分类为"执行中"和"已完成"
// 获取“执行中”“已完成”Tasks所需信息的交集， 但不包括发起单位、组织， 集合地点名称的信息
func (db DBManager) GetTasksFromAdminIDs(adminIDs arrayInt, isFinish bool, offset, countsPerPage int) []TaskInfo {
	var tasks []TaskInfo

	o := orm.NewOrm()
	rawSQL := "SELECT t.task_id task_id, t.title title, t.mem_count mem_count, t.launch_admin_id launch_admin_id, "
	rawSQL += "t.launch_datetime launch_datetime, p.place_name place_name "
	if !isFinish { // 执行中的任务需要有gather_datetime. 已完成任务就不需要
		rawSQL += ", gather_datetime "
	}
	rawSQL += "FROM Tasks t, Places p "
	rawSQL += "WHERE launch_admin_id IN " + fmt.Sprint(adminIDs)
	rawSQL += " AND p.place_id = t.gather_place_id"
	if isFinish {
		rawSQL += " AND finish_datetime <= NOW() "
	} else {
		rawSQL += " AND finish_datetime > NOW() "
	}
	rawSQL += "LIMIT ? OFFSET ?"
	o.Raw(rawSQL, countsPerPage, offset).QueryRows(&tasks)

	return tasks
}

// 根据[]Tasklist中的每个placeID获取placeName
func (db DBManager) writePlaceNamesFromPlaceIDs(tasks []TaskInfo, needLatLng bool) {
	o := orm.NewOrm()
	rawSQL := ""
	if needLatLng {
		rawSQL = "SELECT place_name, place_lat, place_lng "
	} else {
		rawSQL = "SELECT place_name "
	}
	rawSQL += "FROM Places WHERE place_id = ?"
	for i := range tasks {
		o.Raw(rawSQL, tasks[i].PlaceID).QueryRow(&(tasks[i]))
	}
}

func (db DBManager) GetCheckCountsFromTask(taskID int) int {
	var count int
	o := orm.NewOrm()
	o.Raw("SELECT COUNT(*) FROM TaskCheckNames WHERE check_task_id = ?", taskID).QueryRow(&count)
	return count
}

// GetTaskAcceptCount 获取任务的接受人数
func (db DBManager) GetTaskAcceptCount(taskID int) int {
	var count int
	o := orm.NewOrm()
	rawSQL := "SELECT COUNT(*) FROM GatherNotifications WHERE gather_task_id = ? AND read_status = 'AC'"
	o.Raw(rawSQL, taskID).QueryRow(&count)
	return count
}

func (db DBManager) GetRefuseCountsFromTask(taskID int) int {
	var count int
	o := orm.NewOrm()
	rawSQL := "SELECT COUNT(*) FROM GatherNotifications WHERE read_status = 'RF' AND gather_task_id = ?"
	o.Raw(rawSQL, taskID).QueryRow(&count)
	return count
}

// 从OrgIDs获取其所有AdminIDs
func (db DBManager) GetAdminIDsFromOrgs(orgIDs arrayInt) []int {
	var adminIDs []int
	o := orm.NewOrm()
	rawSQL := "SELECT admin_id FROM OrgAdminRelationships WHERE org_id IN " + fmt.Sprint(orgIDs)
	o.Raw(rawSQL).QueryRows(&adminIDs)
	return adminIDs
}

// 从OrgIDs获取下属OrgIDs
func (db DBManager) GetLowerOrgIDsFromOrgIDs(orgIDs arrayInt) []int {
	lowerOrgIDs := make([]int, 0)
	o := orm.NewOrm()
	rawSQL := "SELECT lower_org_id FROM OrgRelationships WHERE higher_org_id IN " + fmt.Sprint(orgIDs)
	o.Raw(rawSQL).QueryRows(&lowerOrgIDs)
	return lowerOrgIDs
}

// 根据TaskID从数据库选出该任务的详情，但不包括发起单位、组织
func (db DBManager) GetTaskDetailFromDB(taskID int) TaskInfo {
	task := TaskInfo{}
	o := orm.NewOrm()
	rawSQL := "SELECT t.task_id task_id, t.title title, t.launch_admin_id launch_admin_id, "
	rawSQL += "t.launch_datetime launch_datetime, t.gather_datetime gather_datetime, "
	rawSQL += "t.finish_datetime finish_datetime, t.mem_count mem_count, "
	rawSQL += "p.place_name place_name, p.place_lat place_lat, p.place_lng place_lng "
	rawSQL += "FROM Tasks t, Places p "
	rawSQL += "WHERE t.task_id = ? AND t.gather_place_id = p.place_id"
	o.Raw(rawSQL, taskID).QueryRow(&task)
	return task
}

// GetOfficeName 通过OfficeID获取Office名称
func (db DBManager) GetOfficeName(officeID int) string {
	officeName := ""
	o := orm.NewOrm()
	o.Raw("SELECT name FROM Offices WHERE office_id = ?", officeID).QueryRow(&officeName)
	return officeName
}

// 通过OrgID获取成员的soldier_id, name
func (db DBManager) GetOrgMems(orgID int, needIMUser bool) []Soldier {
	var soldiers []Soldier
	o := orm.NewOrm()
	rawSQL := "SELECT soldier_id, name"
	if needIMUser {
		rawSQL += ", im_user_id "
	}
	rawSQL += "FROM Soldiers WHERE soldier_id IN ("
	rawSQL += "SELECT soldier_id FROM OrgSoldierRelationships WHERE serve_org_id = ?)"
	o.Raw(rawSQL, orgID).QueryRows(&soldiers)
	return soldiers
}

// 通过OrgID找到所有是Admin的民兵ID
func (db DBManager) GetSoldrIDsWhoAreAdmins(orgID int) []int {
	var soldierIDs []int
	o := orm.NewOrm()
	o.Raw("SELECT leader_sid FROM OrgAdminRelationships WHERE org_id = ?", orgID).QueryRows(&soldierIDs)
	return soldierIDs
}

// 通过OrgID获取Org名称
func (db DBManager) GetOrgName(orgID int) string {
	orgName := ""
	o := orm.NewOrm()
	o.Raw("SELECT name FROM Organizations WHERE org_id = ?", orgID).QueryRow(&orgName)
	return orgName
}

// 通过组织的AdminID获取Admin所属单位名称
func (db DBManager) GetOfficeOrgNameFromAdmin(adminID int, isOffice bool) string {
	officeName := ""
	rawSQL := ""
	o := orm.NewOrm()
	if isOffice {
		rawSQL = "SELECT name FROM Offices WHERE office_id = ("
		rawSQL += "SELECT office_id FROM OfficeAdminRelationships WHERE admin_id = ?)"
	} else {
		rawSQL = "SELECT name FROM Organizations WHERE org_id = ("
		rawSQL += "SELECT org_id FROM OrgAdminRelationships WHERE admin_id = ?)"
	}
	o.Raw(rawSQL, adminID).QueryRow(&officeName)
	return officeName
}

// GetAttendOrgs 通过TaskID获取接受该任务的Orgs（不包括成员）
func (db DBManager) GetAttendOrgs(taskID int) []Org {
	orgs := make([]Org, 0)
	o := orm.NewOrm()
	rawSQL := "SELECT org.org_id org_id, org.name name, off.office_level org_level "
	rawSQL += "FROM TaskAcceptOrgs tao, Organizations org, Offices off "
	rawSQL += "WHERE org.org_id = tao.ac_org_id "
	rawSQL += "AND org.serve_office_id = off.office_id AND tao.ac_task_id = ?"
	o.Raw(rawSQL, taskID).QueryRows(&orgs)
	return orgs
}

// GetAttendOffices 通过TaskID获取接受该任务的Offices（不包括成员）
func (db DBManager) GetAttendOffices(taskID int) []Office {
	offices := make([]Office, 0)
	o := orm.NewOrm()
	rawSQL := "SELECT of.office_id office_id, of.name name, of.office_level office_level "
	rawSQL += "FROM TaskAcceptOffices tof, Offices of "
	rawSQL += "WHERE tof.ac_office_id = of.office_id AND tof.ac_task_id = ?"
	o.Raw(rawSQL, taskID).QueryRows(&offices)
	return offices
}

// GetSoldiersExclude 通过TaskID获取接受该任务的民兵，除去UniqueSoldierIDs里的IDs
func (db DBManager) GetSoldiersExclude(taskID int, uniqueSoldierIDs mapInt) []Soldier {
	soldiers := make([]Soldier, 0)
	o := orm.NewOrm()
	rawSQL := "SELECT s.soldier_id soldier_id, s.name name, o.name serve_office, s.im_user_id "
	rawSQL += "FROM GatherNotifications g, Soldiers s, Offices o "
	rawSQL += "WHERE g.gather_task_id = ? AND s.soldier_id = g.recv_soldier_id "
	rawSQL += "AND o.office_id = s.serve_office_id "
	if len(uniqueSoldierIDs) > 0 {
		rawSQL += "AND g.recv_soldier_id NOT IN " + fmt.Sprint(uniqueSoldierIDs)
	}
	o.Raw(rawSQL, taskID).QueryRows(&soldiers)
	return soldiers
}

type mapInt map[int]bool

func (m mapInt) String() string {
	str := "("
	if len(m) >= 1 {
		for key := range m {
			str += strconv.Itoa(key) + ","
		}
		str = str[:len(str)-1]
	}
	str += ")"
	return str
}

// GetTaskMemCountLaunchDate 通过TaskID获取该任务的MemCount
func (db DBManager) GetTaskMemCountLaunchDate(taskID int) (int, string) {
	var (
		memCount   int
		launchDate string
	)
	o := orm.NewOrm()
	rawSQL := "SELECT mem_count, launch_datetime FROM Tasks WHERE task_id = ?"
	o.Raw(rawSQL, taskID).QueryRow(&memCount, &launchDate)
	return memCount, launchDate
}

// GetTaskNotifyCount 通过TaskID获取通知人数
func (db DBManager) GetTaskNotifyCount(taskID int) int {
	var notifyCount int
	o := orm.NewOrm()
	rawSQL := "SELECT COUNT(*) FROM GatherNotifications WHERE gather_task_id = ?"
	o.Raw(rawSQL, taskID).QueryRow(&notifyCount)
	return notifyCount
}

// GetTaskResponseCount 获取任务的响应人数
func (db DBManager) GetTaskResponseCount(taskID int) int {
	var respCount int
	o := orm.NewOrm()
	rawSQL := "SELECT COUNT(*) FROM GatherNotifications "
	rawSQL += "WHERE gather_task_id = ? AND res_datetime IS NOT NULL"
	o.Raw(rawSQL, taskID).QueryRow(&respCount)
	return respCount
}

// GetTaskAvgRespTime 获取任务的平均响应时间
func (db DBManager) GetTaskAvgRespTime(taskID int) string {
	var avg float64
	o := orm.NewOrm()
	rawSQL := "SELECT avg(res_datetime) FROM GatherNotifications "
	rawSQL += "WHERE gather_task_id = ? AND res_datetime IS NOT NULL"
	o.Raw(rawSQL, taskID).QueryRow(&avg)
	if avg == 0 {
		return ""
	}

	avgTime, _ := time.Parse("20060102150405", strconv.Itoa(int(avg)))
	return avgTime.String()[:19] // 2018-05-29 23:23:20，刚好19个字符
}

// GetTaskResponseMems 获取任务的响应人员列表
func (db DBManager) GetTaskResponseMems(taskID, offset, count int) []Soldier {
	soldiers := make([]Soldier, 0)
	o := orm.NewOrm()
	rawSQL := "SELECT s.soldier_id soldier_id, s.name name, s.im_user_id im_user_id, "
	rawSQL += "s.phone_num phone_num, off.name serve_office, g.read_status status, "
	rawSQL += "g.res_datetime res_datetime "
	rawSQL += "FROM Soldiers s, Offices off, GatherNotifications g "
	rawSQL += "WHERE s.serve_office_id = off.office_id AND g.gather_task_id = ? "
	rawSQL += "AND g.recv_soldier_id = s.soldier_id "
	rawSQL += "LIMIT ? OFFSET ?"
	o.Raw(rawSQL, taskID, count, offset).QueryRows(&soldiers)
	return soldiers
}
