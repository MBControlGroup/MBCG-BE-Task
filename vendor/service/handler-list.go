package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// 查看执行中任务的列表, /task/working/{countsPerPage}/{curPage}
func workingList(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getTaskList(formatter, false)(w, r)
	}
}

// 查看已完成任务的列表, /task/done/{countsPerPage}/{curPage}
func doneList(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getTaskList(formatter, true)(w, r)
	}
}

func getTaskList(formatter *render.Render, isFinish bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析URL参数
		reqData := mux.Vars(r)
		countsPerPageStr := reqData["countsPerPage"]
		countsPerPage, _ := strconv.Atoi(countsPerPageStr)
		curPageStr := reqData["curPage"]
		curPage, _ := strconv.Atoi(curPageStr)

		// 获取AdminID及类型
		adminID, err := getAdminID(w, r)
		if err != nil {
			return
		}

		// 获取任务列表
		taskList, err := Manager.GetTaskList(adminID, countsPerPage, countsPerPage*(curPage-1), isFinish)
		if err != nil {
			fmt.Println(err)
			formatter.JSON(w, http.StatusInternalServerError, serverErrorMsg{internalServerErrorMsg})
			return
		}

		formatter.JSON(w, http.StatusOK, taskList)
	}
}
