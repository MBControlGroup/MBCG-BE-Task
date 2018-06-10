package service

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// NewServer 返回注册路由的negroni
func NewServer() *negroni.Negroni {
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	mx := mux.NewRouter()

	formatter := render.New(render.Options{IndentJSON: true})

	initRoutes(mx, formatter)
	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/task", createTask(formatter)).Methods("POST")      // 创建任务
	mx.HandleFunc("/task/info", basicInfo(formatter)).Methods("POST")  // 获取基本信息
	mx.HandleFunc("/task/finish", endTask(formatter)).Methods("POST")  // 结束任务
	mx.HandleFunc("/task/orgs", orgs(formatter)).Methods("POST")       // 获取所有下属组织及人员
	mx.HandleFunc("/task/offices", offices(formatter)).Methods("POST") // 获取所有下属单位及人员

	mx.HandleFunc("/task/working/{countsPerPage}/{curPage}", workingList(formatter)).Methods("POST") // 查看执行中任务列表
	mx.HandleFunc("/task/done/{countsPerPage}/{curPage}", doneList(formatter)).Methods("POST")       // 查看已完成任务列表

	mx.HandleFunc("/task/detail/{taskID}", detail(formatter)).Methods("POST")         // 查看任务详情
	mx.HandleFunc("/task/detail/mem/{taskID}", detail_mem(formatter)).Methods("POST") // 查看参与任务的人员

	mx.HandleFunc("/task/response/{taskID}", response(formatter)).Methods("POST")                                   // 查看任务响应情况
	mx.HandleFunc("/task/response/mem/{taskID}/{countsPerPage}/{curPage}", response_mem(formatter)).Methods("POST") // 查看任务的响应人员列表

	mx.HandleFunc("/task/gather/{taskID}", gather(formatter)).Methods("POST")                                   // 查看任务集合情况
	mx.HandleFunc("/task/gather/mem/{taskID}/{countsPerPage}/{curPage}", gather_mem(formatter)).Methods("POST") // 查看任务的集合人员列表
}
