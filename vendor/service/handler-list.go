package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// 查看执行中任务的列表, /task/working/{countsPerPage}/{curPage}
func workingList(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析URL参数
		reqData := mux.Vars(r)
		countsPerPage := reqData["countsPerPage"]
		curPage := reqData["curPage"]

		// 获取AdminID及类型
		adminID, isOffice, err := getAdminAndType(w, r)
		if err != nil {
			return
		}

		if isOffice { // Admin是单位类型
DBInfo.
		} else { // Admin是组织类型

		}

	}
}

// 查看已完成任务的列表, /task/done/{countsPerPage}/{curPage}
func doneList(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
