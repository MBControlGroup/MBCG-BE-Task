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
	orm.RegisterModel(new(Task), new(Place))
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

func createPlace(place *Place) (placeID int64) {
	o := orm.NewOrm()
	prepareInsert, _ := o.QueryTable("Places").PrepareInsert()
	placeID, _ = prepareInsert.Insert(place)
	return placeID
}

func createTaskDB(task *Task, place *Place) {
	if task.PlaceID == -1 {
		task.PlaceID = int(createPlace(place))
	}
	o := orm.NewOrm()
	prepareInsert, _ := o.QueryTable("Tasks").PrepareInsert()
	prepareInsert.Insert(task)
}

func createNotifications(t *Task) {

}
