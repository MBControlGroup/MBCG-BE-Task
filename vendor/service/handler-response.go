package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/unrolled/render"
)

// 任务响应情况 [/task/response/{task_id}] [GET]
func response(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := parse(w, r, false, true, false)
		if err != nil {
			return
		}
		response := Manager.GetTaskResponse(reqData.TaskID)
		formatter.JSON(w, http.StatusOK, response)
	}
}

// 任务的响应人员列表 [/task/response/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
func response_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := parse(w, r, false, true, true)
		if err != nil {
			return
		}
		b, _ := json.Marshal(reqData)
		fmt.Println(string(b))
		responseMem := Manager.GetTaskRespMems(reqData.TaskID, reqData.CountsPerPage*(reqData.CurPage-1), reqData.CountsPerPage)
		formatter.JSON(w, http.StatusOK, responseMem)
	}
}
