package service

import (
	"net/http"

	"github.com/unrolled/render"
)

// /task/response/{task_id} [GET]
func response(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, taskID, err := getTaskAndAdminID(w, r)
		if err != nil {
			return
		}
		response := Manager.GetTaskResponse(taskID)
		formatter.JSON(w, http.StatusOK, response)
	}
}

func response_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
