package service

import (
	"net/http"

	"github.com/unrolled/render"
)

// 任务集合情况 [/task/gather/{task_id}] [GET]
func gather(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := parse(w, r, false, true, false)
		if err != nil {
			return
		}
		checkCount := Manager.GetTaskGather(reqData.TaskID)
		formatter.JSON(w, http.StatusOK, struct {
			CheckCount int `json:"check_counts"`
		}{checkCount})
	}
}

// 任务的集合人员列表 [/task/working/gather/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
func gather_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
