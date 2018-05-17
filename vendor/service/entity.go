package service

type Task struct {
	ID      int    `json:"task_id"`
	Title   string `json:"title"`
	Count   int    `json:"mem_count"`
	AdminID uint   `json:"launch_admin_id"`
	Launch  string `json:"launch_datetime"`
	Gather  string `json:"gather_datetime"`
	Detail  string `json:"detail"`
	PlaceID int    `json:"gather_place_id"`
	Finish  string `json:"finish_datetime"`
}

func (t *Task) TableName() string { return "Tasks" }

type Place struct {
	ID   int     `json:"gather_place_id" orm:"column(place_id)"`
	Name string  `json:"gather_place_name" orm:"column(place_name)"`
	Lat  float64 `json:"gather_place_lat" orm:"column(place_lat)"`
	Lng  float64 `json:"gather_place_lng" orm:"column(place_lng)"`
}

func (p *Place) TableName() string { return "Places" }

type AcMem struct {
	AcOrgIDs  []int `json:"accept_org_ids"`
	AcOffIDs  []int `json:"accept_office_ids"`
	AcSoldIDs []int `json:"accept_soldr_ids"`
}

type TaskAcceptOffices struct {
	ID int
}

func (t *TaskAcceptOffices) TableName() string { return "TaskAcceptOffices" }
