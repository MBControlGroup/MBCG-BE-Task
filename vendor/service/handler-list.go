package service

import (
	"net/http"

	"github.com/unrolled/render"
)

// 查看执行中任务的列表, /task/working/{countsPerPage}/{curPage}
func workingList(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// 查看已完成任务的列表, /task/done/{countsPerPage}/{curPage}
func doneList(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
