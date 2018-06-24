package service

import (
	"bytes"
	"control"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"model"
	"net/http"
	"net/url"

	"github.com/unrolled/render"
)

// Manager 全局的控制层管理器
var Manager control.Controller

type tokenMessg struct {
	Success bool   `json:"success"`
	Detail  string `json:"detail"`
	Id      int    `json:"id"`
}

const (
	tokenCookieName = "token"
	reLoginMsg      = "登录超时，请重新登录"

	host           = "http://222.200.180.59"
	tokenValidPort = ":9200"
	tokenValidPath = "/tokenValid"
	loginPath      = "/signin"
)

/*
type redirectMsg struct {
	Msg string `json:"cnmsg"`
	URL string `json:"url"`
}
*/

type taskID struct {
	ID int `json:"task_id"`
}

type returnMessg struct {
	Code  int         `json:"code"`
	Enmsg string      `json:"enmsg"` // 一般为ok
	Cnmsg string      `json:"cnmsg"` // "成功"
	Data  interface{} `json:"data,omitempty"`
}

var internalServerErrorMsg = returnMessg{http.StatusInternalServerError, "error", "很抱歉，服务器出错了", nil}
var redirectMsg = returnMessg{http.StatusFound, "error", "登录超时，请重新登录", nil}

// 结束任务, /task [PUT]
func endTask(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取http.Request中的Body
		reqBody, _ := ioutil.ReadAll(r.Body) // 读取http.Request的Body
		defer r.Body.Close()

		// 从Request Body中获取taskID
		var taskid taskID
		json.Unmarshal(reqBody, &taskid)

		// TODO: 获取管理员ID

		// 结束任务
		err := Manager.EndTask(taskid.ID, 123456) // 将来可能会进行权限控制. 非该任务的发起管理员都不能结束任务
		if err != nil {                           // DB UPDATE 出错
			formatter.JSON(w, http.StatusInternalServerError, internalServerErrorMsg)
		} else { // 成功结束任务
			formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", nil})
		}
	}
}

// BasicInfo 用于 获取基本信息
type BasicInfo struct {
	IsOff  bool          `json:"is_office"`
	Places []model.Place `json:"places"`
}

// 获取基本信息, /task [GET]
func basicInfo(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取管理员ID和类型
		adminID, err := getAdminID(w, r)
		if err != nil { // 若出错, 则GetAdminAndType函数已经对ResponseWriter进行写入, 可直接返回
			return
		}

		// 获取Admin所在组织/单位的常用地点
		places, isOffice, err := Manager.GetCommonPlaces(adminID)
		if err != nil { // 查询出错
			formatter.JSON(w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
		info := BasicInfo{IsOff: isOffice, Places: places}
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", info})
	}
}

// 发布任务, /task [POST]
func createTask(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取AdminID
		adminID, err := getAdminID(w, r)
		if err != nil { // 若出错, getAdminID 已经对 ResponseWriter 写入信息, 故可直接 return
			return
		}

		// 获取http.Request中的Body
		reqBody, _ := ioutil.ReadAll(r.Body) // 读取http.Request的Body
		defer r.Body.Close()
		reqBytes, _ := url.QueryUnescape(string(reqBody)) // 把Body转为bytes

		// 解析Request.Body中的JSON数据
		var (
			reqTask  model.Task
			reqPlace model.Place
			reqAcMem model.AcMem
		)
		reqTask.AdminID = adminID
		json.Unmarshal([]byte(reqBytes), &reqTask)  // 从json中解析Task的内容
		json.Unmarshal([]byte(reqBytes), &reqPlace) // 从json中解析Place的内容
		json.Unmarshal([]byte(reqBytes), &reqAcMem) // 从json中解析AcMem接收集合通知的成员

		fmt.Println("[/task POST] Request Body:")
		fmt.Println(reqTask)
		fmt.Println(reqPlace)
		fmt.Println(reqAcMem)

		uniqueSoldierIDs, err := Manager.CreateTask(&reqTask, &reqPlace, &reqAcMem)
		if err != nil {
			formatter.JSON(w, http.StatusInternalServerError, internalServerErrorMsg)
		} else {
			formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "任务发布成功", nil})
			go Manager.SendMessgs(&reqTask, uniqueSoldierIDs) // 发送短信、语音
		}
	}
}

// 获取所有下属组织及人员, /task/orgs [GET]
func orgs(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID, err := getAdminID(w, r)
		if err != nil {
			return
		}

		orgInfo, err := Manager.GetOrgInfoAndMems(adminID)
		if err != nil {
			fmt.Println(err)
			formatter.JSON(w, http.StatusInternalServerError, internalServerErrorMsg)
		}
		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", orgInfo})
	}
}

// 获取所有下属单位及人员, /task/offices [GET]
func offices(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取AdminID
		adminID, err := getAdminID(w, r)
		if err != nil {
			return
		}

		// 从AdminID获取Offices和成员
		officeInfo, err := Manager.GetOfficeInfoAndMems(adminID)
		if err != nil {
			formatter.JSON(w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}

		formatter.JSON(w, http.StatusOK, returnMessg{http.StatusOK, "ok", "成功", officeInfo})
	}
}

// getAdminID 读取Request中的Cookie, 获取并解析token, 返回管理员ID
// 若出错，会对ResponseWriter进行写入
func getAdminID(w http.ResponseWriter, r *http.Request) (adminID int, err error) {
	//return 1, nil
	formatter := render.New(render.Options{IndentJSON: true})
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取cookie中的token
	c, err := r.Cookie(tokenCookieName)
	if err != nil || c.Value == "" { // 用户可能登录超时，需重新登录
		formatter.JSON(w, http.StatusFound, redirectMsg)
		log.Println(err)
		return 0, err
	}
	token := c.Value

	// 访问 /tokenValid
	reqBody, _ := json.Marshal(struct {
		Token string `json:"token"`
	}{token})
	resp, err := http.Post(host+tokenValidPort+tokenValidPath, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		formatter.JSON(w, http.StatusFound, redirectMsg)
		log.Println(err)
		return 0, errors.New("error: unable to validate identiy")
	}
	defer resp.Body.Close()

	// 获取/tokenValid 返回的信息,包括 AdminID、token是否valid等
	reqBody, _ = ioutil.ReadAll(resp.Body)
	var messg tokenMessg
	json.Unmarshal(reqBody, &messg)
	fmt.Println(messg)
	if !messg.Success { // 可能是token不合法
		formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg)
		return 0, errors.New(messg.Detail)
	}
	fmt.Println("adminid:", messg.Id)
	return messg.Id, nil // token合法，返回 AdminID
}
