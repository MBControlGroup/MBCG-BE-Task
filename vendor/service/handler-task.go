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
	tokenCookieName        = "token"
	reLoginMsg             = "登录超时，请重新登录"
	internalServerErrorMsg = "很抱歉，服务器出错了"
	loginPath              = "/signin"
)

type redirectMsg struct {
	Msg string `json:"cnmsg"`
	URL string `json:"url"`
}

type serverErrorMsg struct {
	Msg string `json:"cnmsg"`
}

type taskID struct {
	ID int `json:"task_id"`
}

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
		err := EndTask(taskid.ID, 123456) // 将来可能会进行权限控制. 非该任务的发起管理员都不能结束任务
		if err != nil {                   // DB UPDATE 出错
			formatter.JSON(w, http.StatusInternalServerError, serverErrorMsg{internalServerErrorMsg})
		} else { // 成功结束任务
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

type BasicInfo struct {
	IsOff  bool               `json:"is_office"`
	Places []PlaceInBasicInfo `json:"places"`
}

type PlaceInBasicInfo struct {
	ID   int     `json:"place_id" orm:"column(place_id)"`
	Name string  `json:"place_name" orm:"column(place_name)"`
	Lat  float64 `json:"place_lat" orm:"column(place_lat)"`
	Lng  float64 `json:"place_lng" orm:"column(place_lng)"`
}

// 获取基本信息, /task [GET]
func basicInfo(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取管理员ID和类型
		adminID, isOffice, err := GetAdminAndType(w, r)
		if err != nil { // 若出错, 则GetAdminAndType函数已经对ResponseWriter进行写入, 可直接返回
			return
		}

		// 获取Admin所在组织/单位的常用地点
		places, err := GetCommonPlaces(adminID, isOffice)
		if err != nil { // 查询出错
			formatter.JSON(w, http.StatusInternalServerError, serverErrorMsg{internalServerErrorMsg})
			return
		}
		info := BasicInfo{IsOff: isOffice, Places: places}
		formatter.JSON(w, http.StatusOK, info)
	}
}

// 发布任务, /task [POST]
func createTask(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取AdminID
		adminID, err := GetAdminID(w, r)
		if err != nil { // 若出错, GetAdminID 已经对 ResponseWriter 写入信息, 故可直接 return
			return
		}

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
		reqTask.AdminID = adminID
		json.Unmarshal([]byte(reqBytes), &reqTask)  // 从json中解析Task的内容
		json.Unmarshal([]byte(reqBytes), &reqPlace) // 从json中解析Place的内容
		json.Unmarshal([]byte(reqBytes), &reqAcMem) // 从json中解析AcMem接收集合通知的成员

		fmt.Println("[/task POST] Request Body:")
		fmt.Println(reqTask)
		fmt.Println(reqPlace)
		fmt.Println(reqAcMem)

		err = CreateTask(&reqTask, &reqPlace, &reqAcMem)
		if err != nil {
			formatter.JSON(w, http.StatusInternalServerError, serverErrorMsg{internalServerErrorMsg})
		} else {
			w.WriteHeader(http.StatusCreated)
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
		// 获取AdminID
		adminID, err := GetAdminID(w, r)
		if err != nil {
			return
		}

		officeInfo := GetOfficesAndMemsFromAdminID(adminID)
	}
}

// OfficeInfo 获取下属单位及人员
type OfficeInfo struct {
	TotalMems    int    `json:"total_mems"`
	OfficeDetail Office `json:"office_detail"`
}

// Office 目前主要针对"获取下属单位及人员"设计
type Office struct {
	ID        int       `json:"office_id"`
	Name      string    `json:"name"`
	Members   []Soldier `json:"members"`
	LowerOffs []Office  `json:"lower_offices"`
}

// Soldier 用于所有JSON数据的传输
type Soldier struct {
	ID          int    `json:"soldier_id" orm:"column(soldier_id)"`
	Name        string `json:"name" orm:"column(name)"`
	Phone       int64  `json:"phone,omitempty" orm:"column(phone_num)"`
	IMUserID    int    `json:"im_user_id,omitempty" orm:"column(im_user_id)"`
	IsAdmin     bool   `json:"is_admin,omitempty"`
	ServeOffice string `json:"serve_office,omitempty"`
	Status      string `json:"status,omitempty"`
	RespTime    string `json:"resp_time,omitempty"`
}
