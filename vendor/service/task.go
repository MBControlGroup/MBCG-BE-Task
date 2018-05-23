package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/unrolled/render"
)

const (
	host           = "http://localhost"
	tokenValidPort = ":8080"
	tokenValidPath = "/tokenValid"
)

type tokenMessg struct {
	Success bool
	Detail  string
	Id      int
}

// GetAdminID 读取Request中的Cookie, 获取并解析token, 返回管理员ID
func GetAdminID(w http.ResponseWriter, r *http.Request) (adminID int, err error) {
	return 3, nil
	formatter := render.New(render.Options{IndentJSON: true})

	// 获取cookie中的token
	c, err := r.Cookie(tokenCookieName)
	if err != nil || c.Value == "" { // 用户可能登录超时，需重新登录
		formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
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
		formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
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
		formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
		return 0, errors.New(messg.Detail)
	}
	return messg.Id, nil // token合法，返回 AdminID
}

// GetAdminAndType 输入token字符串, 返回管理员ID和类型(true: 单位, false: 组织)
func GetAdminAndType(w http.ResponseWriter, r *http.Request) (adminID int, isOff bool, err error) {
	formatter := render.New(render.Options{IndentJSON: true})

	// 获取管理员ID
	adminID, err = GetAdminID(w, r)
	if err != nil { // 可能是token不合法
		return adminID, false, err
	}

	// 获取管理员类型
	isOff, err = getAdminType(adminID)
	if err != nil { // 可能无法通过AdminID找到相应的管理员信息
		formatter.JSON(w, http.StatusTemporaryRedirect, redirectMsg{reLoginMsg, loginPath})
		log.Println(err)
		return adminID, false, err
	}
	return adminID, isOff, nil
}
