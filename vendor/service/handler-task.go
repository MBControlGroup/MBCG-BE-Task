package service

import (
	"encoding/json"
	"fmt"
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
		// 获取cookie中的token
		/*c, err := r.Cookie(tokenCookieName)
		if err != nil || c.Value == "" { // 用户可能登录超时，需重新登录
			formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
			panic(err)
		}

		// 根据token获取AdminID
		token := c.Value
		adminID, err := GetAdmin(token)
		if err != nil {
			formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
			panic(err)
		}*/

		// 获取http.Request中的Body
		reqBody, _ := ioutil.ReadAll(r.Body)              // 读取http.Request的Body
		reqBytes, _ := url.QueryUnescape(string(reqBody)) // 把Body转为bytes
		defer r.Body.Close()

		// 解析Request.Body中的JSON数据
		var (
			reqTask  Task
			reqPlace Place
			reqAcMem AcMem
		)
		var adminID uint = 3
		reqTask.AdminID = adminID
		json.Unmarshal([]byte(reqBytes), &reqTask)  // 从json中解析Task的内容
		json.Unmarshal([]byte(reqBytes), &reqPlace) // 从json中解析Place的内容
		json.Unmarshal([]byte(reqBytes), &reqAcMem) // 从json中解析AcMem接收集合通知的成员
		/*testtask, _ := json.Marshal(&reqTask)
		testplace, _ := json.Marshal(&reqPlace)
		testacmem, _ := json.Marshal(&reqAcMem)*/
		fmt.Println(reqTask)
		fmt.Println(reqPlace)
		fmt.Println(reqAcMem)

		CreateTask(&reqTask, &reqPlace, &reqAcMem)

		w.WriteHeader(http.StatusCreated)
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
