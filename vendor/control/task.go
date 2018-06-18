package control

import (
	"encoding/json"
	"fmt"
	"model"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Controller 全局控制层管理器
type Controller struct{}

var db model.DBManager

// EndTask 根据TaskID，AdminID，结束相应任务
func (c Controller) EndTask(taskID, adminID int) error {
	err := db.EndTask(taskID, adminID)
	return err
}

// GetCommonPlaces 根据AdminID、Admin类型获取其常用地点
func (c Controller) GetCommonPlaces(adminID int) ([]model.Place, bool, error) {
	isOffice, err := db.GetAdminType(adminID)
	if err != nil {
		return nil, false, err
	}
	places := db.GetCommonPlaces(adminID, isOffice)
	return places, isOffice, nil
}

// CreateTask 创建任务
func (c Controller) CreateTask(task *model.Task, place *model.Place, acmem *model.AcMem) (map[int]bool, error) {
	uniqueSoldierIDs, err := db.CreateTask(task, place, acmem)
	return uniqueSoldierIDs, err
}

// SendMessgs 向民兵发送“查看新任务”的短信、语音
func (c Controller) SendMessgs(task *model.Task, soldierIDs map[int]bool) {
	isOffice, _ := db.GetAdminType(task.AdminID)
	officeOrgName := "【" + db.GetOfficeOrgNameFromAdmin(task.AdminID, isOffice) + "】"
	placeName := "【" + db.GetPlaceName(task.PlaceID) + "】"
	detail := "【" + task.Detail + "】"
	telNums := db.GetTelNums(soldierIDs)

	messg := messgTemplate{
		TelNums:     phoneNums(telNums).String(),
		TemplateNum: "3",
		Var1:        officeOrgName,
		Var2:        placeName,
		Var3:        detail,
	}
	messgTemplateBytes, _ := json.Marshal(&messg)
	messgTemplatePayload := strings.NewReader(string(messgTemplateBytes))
	// 调用发送短信接口
	{
		resp, err := http.Post("http://localhost:9400/sendInterfaceTemplateSms?vars=3", "application/json", messgTemplatePayload)
		if err != nil {
			fmt.Println("调用发送短信接口 错误：", err)
		} else {
			defer resp.Body.Close()
			fmt.Println("调用发送短信接口：", resp.StatusCode, resp.Status)
		}
	}

	voice := voiceTemplate{
		Action:      "Webcall",
		ServiceNo:   "02033275113",
		Exten:       phoneNums(telNums).String(),
		WebCallType: "asynchronous",
		CallBackUrl: "http://172.17.0.1:8080/webCall/callback",
		Variable:    "role:2",
	}
	voiceTemplateBytes, _ := json.Marshal(&voice)
	voiceTemplatePayload := strings.NewReader(string(voiceTemplateBytes))
	// 调用发送语音接口
	{
		resp, err := http.Post("http://localhost:9400/webCall", "application/json", voiceTemplatePayload)
		if err != nil {
			fmt.Println("调用发送语音接口 错误：", err)
		} else {
			defer resp.Body.Close()
			fmt.Println("调用发送语音接口：", resp.StatusCode, resp.Status)
		}
	}
}

// 语音模板
type voiceTemplate struct {
	Action      string
	ServiceNo   string
	Exten       string
	WebCallType string
	CallBackUrl string
	Variable    string
}

type phoneNums []int

// 格式：1,2,3
func (p phoneNums) String() string {
	str := ""
	for _, phoneNum := range p {
		str += strconv.Itoa(phoneNum) + ","
	}
	if len(str) == 0 {
		return str
	}
	return str[:len(str)-1]
}

// 发送短信模板
type messgTemplate struct {
	TelNums     string `json:"num"`
	TemplateNum string `json:"templateNum"`
	Var1        string `json:"var1"` // 发送单位
	Var2        string `json:"var2"` // 集合地点
	Var3        string `json:"var3"` // 集合事由
}

// GetOrgInfoAndMems 获取下属组织及成员
func (c Controller) GetOrgInfoAndMems(adminID int) (*model.OrgInfo, error) {
	orgDetail := model.OrgDetail{Orgs: make([]model.Org, 0), LowerOffices: make([]model.OrgDetail, 0)}
	orgInfo := model.OrgInfo{}

	isOffice, _ := db.GetAdminType(adminID)
	if isOffice { // 单位
		// 获取OfficeID
		officeID := db.GetOfficeIDFromAdminID(adminID)
		// 获取OrgDetail
		orgDetail, uniqueSoldiers := c.getOrgDetail(officeID)
		// 获取Total Members
		orgInfo.TotalMems = len(uniqueSoldiers)
		orgInfo.Orgdetail = orgDetail
	} else { // 组织
		// 通过AdminID获取其所在Office的名称
		orgDetail.OfficeName = db.GetOfficeOrgNameFromAdmin(adminID, false)
		// 通过AdminID获取其所在Org名称
		orgID := db.GetOrgIDFromAdminID(adminID)
		// 通过OrgID获取组织、下属组织的信息，如OrgName，成员，成员数量
		orgs, memCount := c.getOrgAndAllLowerOrgs(orgID)
		orgDetail.Orgs = orgs

		orgInfo.Orgdetail = orgDetail
		orgInfo.TotalMems = memCount
	}
	return &orgInfo, nil
}

type safeMap struct {
	lock           sync.Mutex
	uniqueSoldrIDs map[int]bool
}

// 根据OfficeID获取OrgInfo.OrgDetail
func (c Controller) getOrgDetail(officeID int) (model.OrgDetail, map[int]bool) {
	uniqueSoldiers := safeMap{uniqueSoldrIDs: make(map[int]bool)}
	var waitGroup sync.WaitGroup
	orgDetail := model.OrgDetail{Orgs: make([]model.Org, 0), LowerOffices: make([]model.OrgDetail, 0)}

	// 获取officeName
	orgDetail.OfficeName = db.GetOfficeName(officeID)
	// 获取Orgs，Orgs的不重复SoldierIDs
	orgIDs := db.GetOrgIDsFromOffices([]int{officeID})
	if len(orgIDs) > 0 {
		orgs, subUniqueSoldrs := c.getOrgs(orgIDs)
		orgDetail.Orgs = orgs
		// 插入Unique SoldierIDs
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			uniqueSoldiers.lock.Lock()
			defer uniqueSoldiers.lock.Unlock()

			for soldierID := range subUniqueSoldrs {
				uniqueSoldiers.uniqueSoldrIDs[soldierID] = true
			}
		}()
	}
	// 获取LowerOffices
	lowerOfficeIDs := db.GetLowerOfficeIDsFromOffices([]int{officeID})
	if len(lowerOfficeIDs) > 0 {
		var lowerOfficesLock sync.Mutex
		for _, lowerOfficeID := range lowerOfficeIDs {
			waitGroup.Add(1)
			go func(officeID int) {
				defer waitGroup.Done()

				// 根据LowerOfficeIDs获取对应OrgDetails（OrgDetail.LowerOffices）
				lowerOrgDetail, subUniqueSoldrs := c.getOrgDetail(officeID)
				// 插入lowerOffices
				waitGroup.Add(1)
				go func() {
					defer waitGroup.Done()
					lowerOfficesLock.Lock()
					defer lowerOfficesLock.Unlock()

					orgDetail.LowerOffices = append(orgDetail.LowerOffices, lowerOrgDetail)
				}()
				// 插入UniqueSoldiers
				waitGroup.Add(1)
				go func() {
					defer waitGroup.Done()
					uniqueSoldiers.lock.Lock()
					defer uniqueSoldiers.lock.Unlock()

					for soldierID := range subUniqueSoldrs {
						uniqueSoldiers.uniqueSoldrIDs[soldierID] = true
					}
				}()
			}(lowerOfficeID)
		}
	}
	waitGroup.Wait()
	return orgDetail, uniqueSoldiers.uniqueSoldrIDs
}

// 通过OrgID获取组织及其所有下属的名称、成员
// 返回该组织及所有下属组织（的信息），成员数量（成员不重复）
func (c Controller) getOrgAndAllLowerOrgs(orgID int) ([]model.Org, int) {
	orgs := make([]model.Org, 0)
	// 记录该组织及其下属组织的成员数量，人员不重复
	uniqueSoldiers := safeMap{uniqueSoldrIDs: make(map[int]bool)}
	var waitGroup sync.WaitGroup

	// 使用queue找到orgID的所有下属Org，及其name, members, lowerOrgIDs
	queue := make([]int, 1)
	queue[0] = orgID
	for len(queue) != 0 {
		orgID := queue[0]
		org := c.getOrg(orgID)
		orgs = append(orgs, org)

		queue = append(queue, org.LowerOrgIDs...)
		queue = queue[1:] // queue.pop()

		// 记录不重复的民兵
		waitGroup.Add(1)
		go func(members []model.Soldier) {
			defer waitGroup.Done()
			uniqueSoldiers.lock.Lock()
			defer uniqueSoldiers.lock.Unlock()
			for _, soldier := range members {
				uniqueSoldiers.uniqueSoldrIDs[soldier.ID] = true
			}
		}(org.Members)
	}
	waitGroup.Wait()
	memCount := len(uniqueSoldiers.uniqueSoldrIDs)
	return orgs, memCount
}

// 通过OrgID获取Org（name, 成员, 下属OrgIDs）
func (c Controller) getOrg(orgID int) model.Org {
	org := model.Org{ID: orgID}
	org.Name = db.GetOrgName(orgID)
	org.Members = c.getOrgMemsAndAdmins(orgID, false)
	org.LowerOrgIDs = db.GetLowerOrgIDsFromOrgIDs([]int{orgID})
	return org
}

// 通过OrgID获取其成员
func (c Controller) getOrgMemsAndAdmins(orgID int, needIMUser bool) []model.Soldier {
	// 获取是管理员的民兵ID
	isSoldierAdmin := make(map[int]bool) // map[soldierID]isAdmin
	soldierIDs := db.GetSoldrIDsWhoAreAdmins(orgID)
	for _, soldierID := range soldierIDs {
		isSoldierAdmin[soldierID] = true
	}

	soldiers := db.GetOrgMems(orgID, needIMUser)
	// 判断每个Soldier是否为Admin
	for i := range soldiers {
		if isSoldierAdmin[soldiers[i].ID] {
			soldiers[i].IsAdmin = true
		}
	}
	return soldiers
}

// 通过OrgIDs获取Orgs（name, members, lowerOrgIDs）
func (c Controller) getOrgs(orgIDs []int) ([]model.Org, map[int]bool) {
	orgs := make([]model.Org, len(orgIDs))
	uniqueSoldiers := safeMap{uniqueSoldrIDs: make(map[int]bool)}
	var waitGroup sync.WaitGroup

	for i := range orgs {
		waitGroup.Add(1)
		go func(i int) {
			defer waitGroup.Done()
			// 根据OrgID获取组织信息
			orgs[i] = c.getOrg(orgIDs[i])
			// 获取不重复的SoldierID
			uniqueSoldiers.lock.Lock()
			defer uniqueSoldiers.lock.Unlock()
			for _, soldier := range orgs[i].Members {
				uniqueSoldiers.uniqueSoldrIDs[soldier.ID] = true
			}
		}(i)
	}
	waitGroup.Wait()
	return orgs, uniqueSoldiers.uniqueSoldrIDs
}

// GetOfficeInfoAndMems 根据AdminID获取单位、下属单位及成员
func (c Controller) GetOfficeInfoAndMems(adminID int) (*model.OfficeInfo, error) {
	officeID := db.GetOfficeIDFromAdminID(adminID)

	var officeInfo model.OfficeInfo
	officeDetail, memCounts, err := c.getOfficeDetail(officeID)
	if err != nil {
		return nil, err
	}

	officeInfo.OfficeDetail = officeDetail
	officeInfo.TotalMems = memCounts
	return &officeInfo, err
}

// 递归, 获取下属单位、人员及人数
func (c Controller) getOfficeDetail(officeID int) (model.Office, int, error) {
	office := model.Office{ID: officeID, LowerOffs: make([]model.Office, 0)}

	// 根据OfficeID获取单位名称
	office.Name = db.GetOfficeName(officeID)
	// 根据OfficeID获取所含民兵及人数
	soldiers, memCounts := db.GetOfficeMems(officeID, false)
	office.Members = soldiers

	// 获取该单位的下属单位
	lowerOffIDs := db.GetLowerOfficeIDsFromOffices([]int{officeID})
	for _, lowerOffID := range lowerOffIDs {
		lowerOffice, counts, err := c.getOfficeDetail(lowerOffID)
		if err != nil {
			return office, 0, err
		}

		office.LowerOffs = append(office.LowerOffs, lowerOffice)
		memCounts += int64(counts)
	}

	return office, int(memCounts), nil
}
