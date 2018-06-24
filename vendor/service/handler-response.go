package service

import (
	"control"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/unrolled/render"
)

// 任务响应情况 [/task/response/{task_id}] [GET]
func response(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := parse(w, r, false, true, false)
		if err != nil {
			return
		}
		err = getPageFromReqBody(w, r, reqData)
		if err != nil {
			return
		}

		response := Manager.GetTaskResponse(reqData.TaskID)
		responseMem := Manager.GetTaskRespMems(reqData.TaskID, reqData.CountsPerPage*(reqData.CurPage-1), reqData.CountsPerPage)
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", struct {
			*control.TaskResp
			*control.MemList
		}{response, responseMem}})
	}
}

// 任务的响应人员列表 [/task/response/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
/*func response_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := parse(w, r, false, true, true)
		if err != nil {
			return
		}
		b, _ := json.Marshal(reqData)
		fmt.Println(string(b))
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", responseMem})
	}
}*/

func getPageFromReqBody(w http.ResponseWriter, r *http.Request, reqData *result) error {
	formatter := render.New(render.Options{IndentJSON: true})

	// 获取http.Request中的Body
	reqBody, _ := ioutil.ReadAll(r.Body) // 读取http.Request的Body
	defer r.Body.Close()
	reqBytes, _ := url.QueryUnescape(string(reqBody)) // 把Body转为bytes

	// 解析Request.Body中的JSON数据
	err := json.Unmarshal([]byte(reqBytes), reqData) // 从json中解析cur_page, count_per_page
	if err != nil {
		formatter.JSON(w, http.StatusInternalServerError, internalServerErrorMsg)
		return err
	}
	fmt.Println(reqData)
	return nil
}
