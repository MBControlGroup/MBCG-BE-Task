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
		adminID, taskID, err := getTaskAndAdminID(w, r)
		if err != nil {
			return
		}

		taskInfo, _ := Manager.GetTaskDetail(taskID, adminID)
		formatter.JSON(w, http.StatusOK, taskInfo)
	}
}

// 查看参与任务的人员, /task/detail/mem/{taskID} [GET]
func detail_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO：进行权限控制，查看TaskID与AdminID是否对应
		_, taskID, err := getTaskAndAdminID(w, r)
		if err != nil {
			return
		}

		offices, orgs, soldiers := Manager.GetAttendMems(taskID)
		formatter.JSON(w, http.StatusOK, detailMems{offices, orgs, soldiers})
	}
}

type detailMems struct {
	Offices     []model.Office  `json:"offices"`
	Orgs        []model.Org     `json:"orgs"`
	Individuals []model.Soldier `json:"indiv"`
}

// 根据http.Request获取TaskID，AdminID
func getTaskAndAdminID(w http.ResponseWriter, r *http.Request) (adminID, taskID int, err error) {
	adminID, err = getAdminID(w, r)
	if err != nil {
		return 0, 0, err
	}

	reqData := mux.Vars(r)
	taskIDStr := reqData["taskID"]
	taskID, _ = strconv.Atoi(taskIDStr)

	return adminID, taskID, nil
}
