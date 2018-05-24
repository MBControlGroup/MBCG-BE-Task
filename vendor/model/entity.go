package model

type Task struct {
	ID      int    `json:"task_id"`
	Title   string `json:"title"`
	Count   int    `json:"mem_count"`
	AdminID int    `json:"launch_admin_id"`
	Launch  string `json:"launch_datetime"`
	Gather  string `json:"gather_datetime"`
	Detail  string `json:"detail"`
	PlaceID int    `json:"gather_place_id"`
	Finish  string `json:"finish_datetime"`
}

type Place struct {
	ID   int     `json:"place_id" orm:"column(place_id)"`
	Name string  `json:"place_name" orm:"column(place_name)"`
	Lat  float64 `json:"place_lat" orm:"column(place_lat)"`
	Lng  float64 `json:"place_lng" orm:"column(place_lng)"`
}

// func (t *Task) TableName() string { return "Tasks" }

// AcMem 参与任务的组织、单位、个人
type AcMem struct {
	AcOrgIDs  []int `json:"accept_org_ids"`
	AcOffIDs  []int `json:"accept_office_ids"`
	AcSoldIDs []int `json:"accept_soldr_ids"`
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
