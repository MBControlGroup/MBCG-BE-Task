package service

import (
	"fmt"
	"net/http"

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
		reqData, err := parse(w, r, true, false, true)
		if err != nil {
			return
		}

		// 获取任务列表
		taskList, err := Manager.GetTaskList(reqData.AdminID, reqData.CountsPerPage, reqData.CountsPerPage*(reqData.CurPage-1), isFinish)
		if err != nil {
			fmt.Println(err)
			formatter.JSON(w, http.StatusInternalServerError, serverErrorMsg{internalServerErrorMsg})
			return
		}

		formatter.JSON(w, http.StatusOK, taskList)
	}
}
