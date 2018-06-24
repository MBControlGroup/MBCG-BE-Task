package service

import (
	"control"
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
		err = getPageFromReqBody(w, r, reqData)
		if err != nil {
			return
		}

		gatherDetail := Manager.GetTaskGather(reqData.TaskID)
		memList := Manager.GetTaskGatherMems(reqData.TaskID, reqData.CountsPerPage*(reqData.CurPage-1), reqData.CountsPerPage)
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", struct {
			*control.GatherDetail
			*control.MemList
		}{gatherDetail, memList}})
	}
}

// 任务的集合人员列表 [/task/working/gather/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
/*func gather_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := parse(w, r, false, true, true)
		if err != nil {
			return
		}
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", memList})
	}
}*/
