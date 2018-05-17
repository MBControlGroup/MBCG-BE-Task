package service

import (
	"errors"

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
	result, err := o.Raw("INSERT INTO Places(place_name, place_lat, place_lng) VALUES(?,?,?)", place.Name, place.Lat, place.Lng).Exec()
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

func insertNotifications(taskID int, acmem *AcMem) {

}
