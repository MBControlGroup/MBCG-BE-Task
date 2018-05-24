package model

import (
	"errors"
	"fmt"

	"strconv"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.Debug = true
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", "mbcsdev:mbcsdev2018@tcp(222.200.180.59:9000)/MBDB?charset=utf8")
	//orm.RegisterModelWithPrefix("mb", new(Task), new(Place))
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

// 通过单位ID获取所有民兵ID
func (db DBManager) getSoldrIDsFromOfficeIDs(officeIDs arrayInt) []int {
	var soldierIDs []int
	rawSQL := "SELECT DISTINCT(soldier_id) FROM Soldiers WHERE serve_office_id IN "
	//rawSQL := "SELECT soldier_id FROM Soldiers WHERE serve_office_id IN "
	rawSQL += fmt.Sprint(officeIDs)
	o := orm.NewOrm()
	o.Raw(rawSQL).QueryRows(&soldierIDs)
	return soldierIDs
}

// 通过组织ID获取所有民兵ID
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

// GetOfficesAndMemsFromAdminID 根据AdminID获取单位、下属单位及成员
func (db DBManager) GetOfficesAndMemsFromAdminID(adminID int) (*OfficeInfo, error) {
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
