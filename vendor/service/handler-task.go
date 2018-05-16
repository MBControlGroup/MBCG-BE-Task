package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

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

		// 根据token获取AdminID
		token := c.Value
		adminID, err := GetAdmin(token)
		if err != nil {
			formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
			panic(err)
		}

		// 获取post信息
		sbody, _ := ioutil.ReadAll(r.Body) // 读取http的json参数
		body, _ := url.QueryUnescape(string(sbody))
		defer r.Body.Close()

		var t Task
		json.Unmarshal([]byte(body), &t) // 从json中解析Task的内容
		CreateTask(adminID, t)
	}
}

type Task struct {
	Title     string  `json:"title"`
	Count     int     `json:"mem_count"`
	Launch    string  `json:"launch_datetime"`
	Gather    string  `json:"gather_datetime"`
	Detail    string  `json:"detail"`
	PlaceID   int     `json:"gather_place_id"`
	PlaceName string  `json:"gather_place_name"`
	PlaceLat  float64 `json:"gather_place_lat"`
	PlaceLng  float64 `json:"gather_place_lng"`
	Finish    string  `json:"finish_datetime"`
	AcOrgIDs  []int   `json:"accept_org_ids"`
	AcOffIDs  []int   `json:"accept_office_ids"`
	AcSoldIDs []int   `json:"accept_soldr_ids"`
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
