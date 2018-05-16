package task

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	orm.RegisterDataBase("default", "mysql", "mbcsdev:mbcsdev2018@tcp(222.200.180.59:9000)/MBDB?charset=utf8")
}
