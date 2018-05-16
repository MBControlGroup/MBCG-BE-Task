package service

import (
	"net/http"

	"task"

	"github.com/unrolled/render"
)

const (
	tokenCookieName = "token"
	reLoginMsg      = "登录超时，请重新登录"
	loginPath       = "/signin"
)

type redirectMsg struct {
	Msg string `json:"cnmsg"`
	URL string `json:"url"`
}

// 结束任务, /task [PUT]
func endTask(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// 获取基本信息, /task [GET]
func basicInfo(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// 发布任务, /task [POST]
func createTask(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(tokenCookieName) // 获取cookie中的token
		if err != nil || c.Value == "" {    // 用户可能登录超时，需重新登录
			formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
			panic(err)
		}

		token := c.Value
		adminID, err := task.GetAdmin(token)
		if err != nil {
			formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
			panic(err)
		}
	}
}

// 获取所有下属组织及人员, /task/orgs [GET]
func orgs(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// 获取所有下属单位及人员, /task/offices [GET]
func offices(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
