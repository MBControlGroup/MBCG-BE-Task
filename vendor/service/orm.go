package service

import (
	"errors"

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

func getAdminType(adminID uint) (isOff bool, err error) {
	o := orm.NewOrm()
	var adminType string
	err = o.Raw("SELECT admin_type FROM Admins WHERE admin_id = ?").QueryRow(&adminType)
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

func insertPlace(place *Place) int {
	o := orm.NewOrm()
	o.Begin()
	rawSQL := "INSERT INTO Places(place_name, place_lat, place_lng) VALUES(?,?,?)"
	result, err := o.Raw(rawSQL, place.Name, place.Lat, place.Lng).Exec()
	if err != nil {
		o.Rollback()
	} else {
		o.Commit()
	}
	placeID, _ := result.LastInsertId()
	return int(placeID)
}

func insertTask(task *Task, place *Place) {
	if task.PlaceID == -1 {
		task.PlaceID = insertPlace(place)
	}
	o := orm.NewOrm()
	o.Begin()
	rawSQL := "INSERT INTO Tasks"
	rawSQL += "(title, mem_count, launch_admin_id, gather_datetime, detail, gather_place_id, finish_datetime)"
	rawSQL += "VALUES(?,?,?,?,?,?,?)"
	_, err := o.Raw(rawSQL, task.Title, task.Count, task.AdminID, task.Gather, task.Detail, task.PlaceID, task.Finish).Exec()
	if err != nil {
		o.Rollback()
	} else {
		o.Commit()
	}
}

func insertAcMem(taskID int, acmem *AcMem) {
	o := orm.NewOrm()
	o.Begin()
	taskIDStr := strconv.Itoa(taskID)

	// 批量插入“接受任务的单位”
	rawSQL := "INSERT INTO TaskAcceptOffices(ac_task_id, ac_office_id) VALUES"
	for _, acOffID := range acmem.AcOffIDs {
		rawSQL += "(" + taskIDStr + "," + strconv.Itoa(acOffID) + "),"
	}
	_, err := o.Raw(rawSQL[:len(rawSQL)-1]).Exec()
	if err != nil {
		o.Rollback()
		return
	}
	// 插入“接受任务的组织”
	rawSQL = "INSERT INTO TaskAcceptOrgs(ac_task_id, ac_org_id) VALUES"
	for _, acOrgID := range acmem.AcOrgIDs {
		rawSQL += "(" + taskIDStr + "," + strconv.Itoa(acOrgID) + "),"
	}
	_, err = o.Raw(rawSQL[:len(rawSQL)-1]).Exec()
	if err != nil {
		o.Rollback()
		return
	}

	uniqueSoldrIDs := make(map[int]bool) // 从单位、组织、个人中选取的所有民兵ID，因为人员可能有重复，故用map消重
	var soldierIDs []int
	// 从单位ID获取民兵ID
	for _, acOffID := range acmem.AcOffIDs {
		soldierIDs = getSoldrIDsFromOrg(acOffID)
		for _, soldrID := range soldierIDs {
			uniqueSoldrIDs[soldrID] = true
		}
	}
	// 从组织ID获取民兵ID
	for _, acOrgID := range acmem.AcOrgIDs {
		soldierIDs = getSoldrIDsFromOrg(acOrgID)
		for _, soldrID := range soldierIDs {
			uniqueSoldrIDs[soldrID] = true
		}
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
	_, err = o.Raw(rawSQL[:len(rawSQL)-1]).Exec()
	if err != nil {
		o.Rollback()
	} else {
		o.Commit()
	}

	// TODO: 对所有 uniqueSoldrIDs 进行广告（模板消息、短信）操作
}

// 通过单位ID获取所有民兵ID
func getSoldrIDsFromOffice(officeID int) []int {
	var soldierIDs []int
	rawSQL := "SELECT soldier_id FROM Soldiers WHERE serve_office_id = ?"
	o := orm.NewOrm()
	o.Raw(rawSQL, officeID).QueryRows(&soldierIDs)
	return soldierIDs
}

// 通过组织ID获取所有民兵ID
func getSoldrIDsFromOrg(orgID int) []int {
	var soldierIDs []int
	rawSQL := "SELECT soldier_id FROM OrgSoldierRelationships WHERE serve_org_id = ?"
	o := orm.NewOrm()
	o.Raw(rawSQL, orgID).QueryRows(&soldierIDs)
	return soldierIDs
}
