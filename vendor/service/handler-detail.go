package service

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// 查看任务详情, /task/detail/{taskID} [GET]
func detail(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID, err := getAdminID(w, r)
		if err != nil {
			return
		}

		reqData := mux.Vars(r)
		taskIDStr := reqData["taskID"]
		taskID, _ := strconv.Atoi(taskIDStr)

		taskInfo, _ := Manager.GetTaskDetail(taskID, adminID)
		formatter.JSON(w, http.StatusOK, taskInfo)
	}
}

// 查看参与任务的人员, /task/detail/mem/{taskID} [GET]
func detail_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
