package model

// Task 发布任务的POST数据
type Task struct {
	ID      int    `json:"task_id"`
	Title   string `json:"title"`
	Count   int    `json:"mem_count"`
	AdminID int    `json:"launch_admin_id"`
	Launch  string `json:"launch_datetime"`
	Gather  string `json:"gather_datetime"`
	Detail  string `json:"detail"`
	PlaceID int    `json:"place_id"`
	Finish  string `json:"finish_datetime"`
}

type Place struct {
	ID   int     `json:"place_id" orm:"column(place_id)"`
	Name string  `json:"place_name" orm:"column(place_name)"`
	Lat  float64 `json:"place_lat" orm:"column(place_lat)"`
	Lng  float64 `json:"place_lng" orm:"column(place_lng)"`
}

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

type Office struct {
	ID        int       `json:"office_id" orm:"column(office_id)"`
	Name      string    `json:"name" orm:"column(name)"`
	Level     string    `json:"office_level" orm:"column(office_level)"`
	Members   []Soldier `json:"members"`
	LowerOffs []Office  `json:"lower_offices"`
}

// OrgInfo 获取下属组织及人员
type OrgInfo struct {
	TotalMems int       `json:"total_mems"`
	Orgdetail OrgDetail `json:"detail"`
}

type OrgDetail struct {
	// Orgs所属单位的名称
	OfficeName   string      `json:"office_name"`
	Orgs         []Org       `json:"orgs"`
	LowerOffices []OrgDetail `json:"lower_offices"`
}

type Org struct {
	ID          int       `json:"org_id,omitempty" orm:"column(org_id)"`
	Name        string    `json:"name" orm:"column(name)"`
	Level       string    `json:"org_level,omitempty" orm:"column(org_level)"`
	Members     []Soldier `json:"members"`
	LowerOrgIDs []int     `json:"lower_org_ids,omitempty"`
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

// List 页数, 任务列表
type List struct {
	PageCount int        `json:"total_pages"`
	TaskCount int        `json:"total_tasks"`
	Tasks     []TaskInfo `json:"data"`
}

// TaskInfo 用于任务列表、任务详情
type TaskInfo struct {
	ID           int     `json:"task_id" orm:"column(task_id)"`
	Title        string  `json:"title" orm:"column(title)"`
	Launcher     string  `json:"launch_admin"`
	AdminID      int     `json:"-" orm:"column(launch_admin_id)"`
	LaunchTime   string  `json:"launch_datetime" orm:"column(launch_datetime)"`
	GatherTime   string  `json:"gather_datetime,omitempty" orm:"column(gather_datetime)"`
	FinishTime   string  `json:"finish_datetime,omitempty" orm:"column(finish_datetime)"`
	PlaceID      int     `json:"-" orm:"column(gather_place_id)"`
	Place        string  `json:"gather_place" orm:"column(place_name)"`
	PlaceLat     float64 `json:"place_lat,omitempty" orm:"column(place_lat)"`
	PlaceLng     float64 `json:"place_lng,omitempty" orm:"column(place_lng)"`
	MemCount     int     `json:"mem_count" orm:"column(mem_count)"`
	Status       string  `json:"status,omitempty"`
	StatusDetail float32 `json:"detail,omitempty"`
	RespCount    int     `json:"response_count,omitempty"`
	AcCount      int     `json:"accept_count,omitempty"`
	CheckCount   int     `json:"check_count,omitempty"`
	IsLauncher   bool    `json:"is_launcher,omitempty"`
}
