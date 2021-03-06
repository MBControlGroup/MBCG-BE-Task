package control

import (
	"fmt"
	"math"
	"model"
	"sync"
	"time"
)

// GetTaskResponse 通过TaskID获取任务响应情况
func (c Controller) GetTaskResponse(taskID int) *TaskResp {
	resp := TaskResp{}
	var wg sync.WaitGroup
	// 任务目标征集人数 memCount
	memCount, launchDateStr := db.GetTaskMemCountLaunchDate(taskID)
	resp.MemCount = memCount
	// 通知人数 notifyCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp.NotifyCount = db.GetTaskNotifyCount(taskID)
	}()
	// 响应人数 respCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp.RespCount = db.GetTaskResponseCount(taskID)
	}()
	// 接受人数 acceptCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp.AcceptCount = db.GetTaskAcceptCount(taskID)
	}()
	// 平均响应时间 avgRespTime
	wg.Add(1)
	go func() {
		defer wg.Done()
		avgRespTimeStr := db.GetTaskAvgRespTime(taskID)
		if len(avgRespTimeStr) == 0 {
			resp.AvgRespTime = ""
		} else {
			avgRespTime, _ := time.Parse("2006-01-02 15:04:05", avgRespTimeStr)
			launchDate, _ := time.Parse("2006-01-02 15:04:05", launchDateStr)
			delta, _ := time.Parse("2006-01-02 15:04:05", "0000-00-00 00:00:00")
			avgDelta := delta.Add(time.Duration(avgRespTime.Unix()-launchDate.Unix()) * time.Second)
			fmt.Println(avgDelta)
			resp.AvgRespTime = avgDelta.String()[11:19] // 01:12:54
		}
	}()
	wg.Wait()
	return &resp
}

type TaskResp struct {
	MemCount    int    `json:"mem_count" orm:"column(mem_count)"`
	NotifyCount int    `json:"notify_count" orm:"column(notify_count)"`
	RespCount   int    `json:"response_count" orm:"resp_count"`
	AcceptCount int    `json:"accept_count" orm:"ac_count"`
	AvgRespTime string `json:"avg_resp_time,omitempty" orm:"avg_resp_time"`
}

// GetTaskRespMems 任务的响应人员列表
func (c Controller) GetTaskRespMems(taskID, offset, countsPerPage int) *MemList {
	respMemList := MemList{}
	respCount := db.GetTaskResponseCount(taskID)
	respMemList.MemCount = respCount
	respMemList.PageCount = int(math.Ceil(float64(respCount) / float64(countsPerPage)))
	respMemList.Members = db.GetTaskResponseMems(taskID, offset, countsPerPage)
	return &respMemList
}

type MemList struct {
	PageCount int             `json:"total_pages"`
	MemCount  int             `json:"total_mem"`
	Members   []model.Soldier `json:"mem_list"`
}
