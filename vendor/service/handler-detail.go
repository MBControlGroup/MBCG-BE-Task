package service

import (
	"model"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// 查看任务详情, /task/detail/{taskID} [GET]
func detail(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := parse(w, r, true, true, false)
		if err != nil {
			return
		}

		taskInfo, _ := Manager.GetTaskDetail(reqData.TaskID, reqData.AdminID)
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", taskInfo})
	}
}

// 查看参与任务的人员, /task/detail/mem/{taskID} [GET]
func detail_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO：进行权限控制，查看TaskID与AdminID是否对应
		reqData, err := parse(w, r, false, true, false)
		if err != nil {
			return
		}

		offices, orgs, soldiers := Manager.GetAttendMems(reqData.TaskID)
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", detailMems{offices, orgs, soldiers}})
	}
}

type detailMems struct {
	Offices     []model.Office  `json:"offices"`
	Orgs        []model.Org     `json:"orgs"`
	Individuals []model.Soldier `json:"indiv"`
}

// 根据http.Request获取TaskID，AdminID, CountsPerPage, CurPage
func parse(w http.ResponseWriter, r *http.Request, needAdmin, needTask, needPage bool) (*result, error) {
	res := result{}
	// AdminID
	if needAdmin {
		adminID, err := getAdminID(w, r)
		if err != nil {
			return &res, err
		}
		res.AdminID = adminID
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	// TaskID
	reqData := mux.Vars(r)
	if needTask {
		taskIDStr := reqData["taskID"]
		res.TaskID, _ = strconv.Atoi(taskIDStr)
	}
	// CountsPerPage, CurPage
	if needPage {
		countsPerPageStr := reqData["countsPerPage"]
		res.CountsPerPage, _ = strconv.Atoi(countsPerPageStr)
		curPageStr := reqData["curPage"]
		res.CurPage, _ = strconv.Atoi(curPageStr)
	}
	return &res, nil
}

type result struct {
	AdminID       int
	TaskID        int
	CountsPerPage int
	CurPage       int
}
