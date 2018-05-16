package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	host           = "http://localhost"
	tokenValidPath = "/tokenValid"
)

type tokenMessg struct {
	Success bool
	Detail  string
	Id      uint
}

// GetAdmin 输入token字符串, 返回管理员ID
func GetAdmin(token string) (adminID uint, err error) {
	// 访问/tokenValid
	b, _ := json.Marshal(struct {
		Token string `json:"token"`
	}{token})
	resp, err := http.Post(host+tokenValidPath, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Println(err)
		return 0, errors.New("error: unable to validate identiy")
	}
	defer resp.Body.Close()

	// 获取/tokenValid 返回的信息,包括 AdminID、token是否valid等
	b, _ = ioutil.ReadAll(resp.Body)
	var messg tokenMessg
	json.Unmarshal(b, &messg)
	if !messg.Success { // 可能是token不合法
		return 0, errors.New(messg.Detail)
	}
	return messg.Id, nil // token合法，返回 AdminID
}

// GetAdminAndType 输入token字符串, 返回管理员ID和类型(true: 单位, false: 组织)
func GetAdminAndType(token string) (adminID uint, isOff bool, err error) {
	adminID, err = GetAdmin(token)
	if err != nil { // 可能是token不合法
		return adminID, false, err
	}

	isOff, err = getAdminType(adminID)
	if err != nil { // 可能无法通过AdminID找到相应的管理员信息
		return adminID, false, err
	}
	return adminID, isOff, nil
}
