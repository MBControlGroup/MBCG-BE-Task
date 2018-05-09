# 任务 [/task]

该API的简化版在
https://mubu.com/doc/Hn0KEDKdm

## 发布任务 [/task] [POST]
```
request:
{
	"title"			:"任务主题",		// 任务主题
	"mem_count"		:60,			// 集合人数
	"launch_datetime"	:"2018-05-06 17:04:01",	// 发起时间，可缺省
	"gather_datetime"	:"2018-05-07 20:00:00",	// 集合时间
	"detail"		:"任务详情",		// 任务详情
	"gather_place_id"	:123456,		// 集合地点id，为-1视为新建地点
	"gather_place_name"	:"集合地点名称",	// 可缺省
	"gather_place_lat"	:39.071510,		// 集合地点纬度，可缺省
	"gather_place_lng"	:117.190091,		// 集合地点经度，可缺省
	"finish_datetime"	:"2018-05-07 21:00:00",	// 任务结束时间
	"accept_org_ids"	:[11,22,33,44],		// 接受该任务的组织id，可缺省
	"accept_office_ids"	:[11,22,33,44],		// 接受该任务的单位id，可缺省
	"accept_soldr_ids"	:[11,22,33,44],		// 接受该任务的民兵id，可缺省.当该项不为空时,说明指挥官选择了个人下发任务
}
response:
	201 Created

	202 Failed
	{
		"cnmsg":"发布任务失败，XXX有误"
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法进行操作"
	}

	500 Internal Server Error
	{
		"cnmsg":"很抱歉,服务器出错了"
	}
```

## 获取基本信息 [/task] [GET]
点击“发布任务”时，获取常用地点、是否为单位/组织等信息
```
request:
	null
response:
	200 OK
	{
		"is_office"	:true,	// true为单位，false为组织；组织只能获取组织及人员信息，不能获取单位信息
		"places"	:[	// 单位/组织常用地点
			{
				"place_id"	:123456,	// 常用集合地点id
				"place_name":"新天地",	// 常用集合地点名称
				"place_lat"	:39.071510,	// 常用集合地点纬度，不知前端需不需要
				"place_lng"	:117.190091	// 常用集合地点经度
			},
			...
		]
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法进行操作"
	}

	500 Internal Server Error
	{
		"cnmsg":"很抱歉,服务器出错了"
	}
```

### 获取所有下属组织及人员 [/task/orgs] [GET]
点击"发布任务"时，服务器会返回可选择的组织、人员
```
request:
	null
response:
	200 OK
	{
		"total_mems":200,				// 所有下属单位总人数.因为一个人可在多个组织内,故返回单位人数
		"orgs_detail":					// 组织详情
		{
			"office_name":"海珠人武部",		// 单位名称
			"orgs":[				// 单位所含有的组织
				{
					"org_id":123456,	// 组织id
					"name"	:"海珠区一排",	// 组织名称
					"members":[		// 组织成员
						{
							"soldier_id"	:123456,	// 民兵id
							"name"		:"张三",	// 民兵姓名
							"is_admin"	:true		// 是否为管理员
						},
						...
					]
					"lower_orgs":[		// 下属组织,嵌套"orgs"
						{},
						...
					]
				},
				...
			],
			"lower_offices":[			// 下属单位,嵌套"orgs_detail"
				{},
				...
			]
		}
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法进行操作"
	}

	500 Internal Server Error
	{
		"cnmsg":"很抱歉,服务器出错了"
	}
```

### 获取所有下属单位及成员 [/task/offices] [GET]
点击"发布任务"时，服务器会返回可选择的单位、人员
```
request:
	null
response:
	200 OK
	{
		"total_mems":200	// 所有下属单位的总人数
		"office_detail":	// 单位详情
		{
			"office_id"	:123456,			// 单位id
			"name"		:"海珠人武部",			// 单位名称
			"memebers"	:[
				{
					"soldier_id"	:123456,	// 民兵id
					"name"		:"李四",	// 民兵姓名
				},
				...
			]
			"lower_offices 下级单位":[			// 嵌套 office_detail
				{},
				...
			]
		},

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法进行操作"
	}

	500 Internal Server Error
	{
		"cnmsg":"很抱歉,服务器出错了"
	}
```


## 执行中任务 [/task/working]

### 结束执行中任务 [/task/working] [PUT]
```
request:
	{
		"task_id":123456
	}
response:
	204 No Content	// 成功结束任务

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}
	
	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法进行操作"
	}

	500 Internal Server Error
	{
		"cnmsg":"很抱歉,服务器出错了"
	}
```
	
### 查看执行中任务列表 [/task/working/{item_counts_per_page}/{cur_page}] [GET]
item_counts_per_page: 每页显示执行中任务的数目
cur_page: 当前页数
由于任务数量可能会很多,故分页显示
```
request:
	null
response:
	200 OK
	{
		"total_pages":1,	// 总页数
		"total_tasks":5,	// 总任务数(执行中)
		"data":[
			{
				"title"			:"消防演习",		// 任务主题
				"launch_admin"		:"海珠区一排",		// 发起人
				"launch_datetime"	:"2018-05-06 12:00:00",	// 发起时间
				"gather_datetime"	:"2018-05-06 13:00:00",	// 集合时间
				"gather_place"		:"新天地",		// 集合地点
				"status"		:"zj/jh/zx",		// 人员征集中,集合中,任务执行中
				"detail"		:0.4,			// 若为zj,0.4, 就是人员征集了40%
				"mem_count"		:30,			// 目标征集人数
			},
			...
		]
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"参数错误"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```


下面"执行中任务实时情况","响应情况","集合情况"参考了管理端的原型设计
### 执行中任务实时情况 [/task/working/deatil]
"实时情况"包括任务的所有详细信息,地图,人员地理位置.(人员提供了即时通讯id,供即时通讯使用)

#### 查看执行中任务实时情况 [/task/working/detail/{task_id}] [GET]
```
request:
	null
response:
	200 OK
	{				// 若需获取参与任务的成员信息,请访问 /task/working/detail/mem/{task_id}
		"task_id"		:123456,		// 任务id
		"title"			:"消防演习",		// 任务主题
		"launch_admin"		:"海珠区一排",		// 发起人
		"launch_datetime"	:"2018-05-06 12:00:00",	// 发起时间
		"gather_datetime"	:"2018-05-06 13:00:00",	// 集合时间
		"gather_place"		:"新天地",		// 集合地点
		"place_lat"		:39.071510,		// 集合地点纬度，不知前端需不需要
		"place_lng"		:117.190091,		// 集合地点经度
		"status"		:"zj/jh/zx",		// 人员征集中,集合中,任务执行中
		"detail"		:0.4,			// 若为zj,0.4, 就是人员征集了40%
		"mem_count"		:30			// 目标征集人数
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```

#### 查看执行中任务的人员 [/task/working/detail/mem/{task_id}] [GET]
人员的显示按照指挥官选取的单位、组织、个人来显示.
若指挥官选取的是"海珠区一排"(属于组织),则显示"orgs",其他为空.若选取的是个人,则显示"indiv",其他为空
暂不显示单位、组织的上下级关系
```
request:
	null
response:
	200 OK
	{
		"offices":[					// 单位.若发布任务时没选择单位,则"offices"为空
			{
				"name"		:"海珠人武部",	// 单位名称
				"office_level"	:"S"/"D"/"C",	// S代表街道，D代表区，C代表市
				"members"	:[		// 单位中参与任务的人员
					{
						"soldier_name"	:"王五",	// 民兵姓名
						"im_user_id"	:123456		// 即时通讯id
					},
					...
				]
			},
			...
		],
		"orgs":[					// 组织.若发布任务时没选择组织,则"orgs"为空
			{
				"name"		:"海珠区一排",	// 组织名称
				"orgs_level":"S"/"D"/"C",	// 所属单位的级别.S代表街道，D代表区，C代表市
				"members"	:[
					{
						"soldier_name"	:"王五",	// 民兵姓名
						"im_user_id"	:123456,	// 即时通讯id
						"is_admin"	:true		// 是否为组织的管理员
					},
					...
				]
			},
			...
		],
		"indiv":[	// 上级可能单独选了某些人发布任务.若上级没有单独选择某些人,则"indiv"为空
			{
				"name"		:"张三",	// 民兵姓名
				"serve_office"	:"海珠人武部",	// 所属单位名称
				"serve_orgs"	:"海珠区一排",	// 所属组织名称,可能返回第一个找到的组织
				"im_user_id"	:123456		// 即时通讯id
			},
			...
		]
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```


### 执行中任务响应情况 [/task/working/response]

#### 查看执行中任务响应情况 [/task/working/response/{task_id}] [GET]
若要查看个人响应情况的列表,请访问 /task/working/response/mem/{task_id}/{item_counts_per_page}/{cur_page}
```
request:
	null
response:
	200 OK
	{
		"mem_count"	:30		// 目标征集人数
		"notify_count"	:100		// 通知人数
		"response_count":70		// 响应人数
		"accept_count"	:30		// 接受人数
		"avg_acpt_time"	:"00:30:10"	// 平均接受任务耗时
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```

#### 查看执行中任务响应人员 [/task/working/response/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
item_counts_per_page: 每页显示人员的数量
cur_page: 当前页数
跟任务列表类似,列出人员姓名、响应状态、所属组织/单位、直属指挥官、耗时等信息

### 执行中任务集合情况 [/task/working/gather]

#### 查看执行中任务集合情况 [/task/working/gather/{task_id}] [GET]
若要查看个人响应情况的列表,请访问 /task/working/gather/mem/{task_id}/{item_counts_per_page}/{cur_page}
```
request:
	null
response:
	200 OK
	{
		"check_counts"	:30,		// 签到人数
		"avg_chk_time"	:"00:50:31"	// 平均签到耗时
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```

#### 查看执行中任务集合人员 [/task/working/gather/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
item_counts_per_page: 每页显示人员的数量
cur_page: 当前页数


## 已完成任务 [/task/done]

### 查看已完成任务列表 [/task/done/{item_counts_per_page}/{cur_page}] [GET]
item_counts_per_page: 每页显示已完成任务的数目
cur_page: 当前页数
```
request:
	null
response:
	200 OK
	{
		"total_pages":1,	// 总页数
		"total_tasks":5,	// 总任务数(已完成)
		"data":[
			{
				"title"			:"消防演习",		// 任务主题
				"launch_admin"		:"海珠区一排",		// 发起人
				"launch_datetime"	:"2018-05-06 12:00:00",	// 发起时间
				"gather_place"		:"新天地",		// 集合地点
				"mem_count"		:30,			// 目标征集人数
				"response_count"	:100,			// 响应人数
				"accept_count"		:30,			// 接收人数
				"check_count"		:30			// 签到人数
			},
			...
		]
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"参数错误"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```


### 已完成任务详情 [/task/done/detail]

#### 查看已完成任务详情 [/task/done/detail/{task_id}] [GET]
```
request:
	null
response:
	200 OK
	{				// 由于已结束任务不需要地图信息,就不提供"查看已完成任务的人员"接口了
		"task_id"		:123456,		// 任务id
		"title"			:"消防演习",		// 任务主题
		"launch_admin"		:"海珠区一排",		// 发起人
		"launch_datetime"	:"2018-05-06 12:00:00",	// 发起时间
		"gather_datetime"	:"2018-05-06 13:00:00",	// 集合时间
		"gather_place"		:"新天地",		// 集合地点
		"place_lat"		:39.071510,		// 集合地点纬度，不知前端需不需要
		"place_lng"		:117.190091,		// 集合地点经度
		"mem_count"		:30			// 目标征集人数
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```


### 已完成任务响应情况 [/task/done/response]

#### 查看已完成任务响应情况 [/task/done/response/{task_id}] [GET]
```
request:
	null
response:
	200 OK
	{
		"mem_count"	:30		// 目标征集人数
		"notify_count"	:100		// 通知人数
		"response_count":70		// 响应人数
		"accept_count"	:30		// 接受人数
		"avg_acpt_time"	:"00:30:10"	// 平均接受任务耗时
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}
```

#### 查看已完成任务响应人员 [/task/done/response/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
// item_counts_per_page: 每页显示人员的数量
// cur_page: 当前页数
// 与"查看执行中任务相应人员"类似

### 已完成任务集合情况 [/task/done/gather/{task_id}]

#### 查看已完成任务集合情况 [/task/done/gather/{task_id}] [GET]
request:
	null
response:
	200 OK
	{
		"check_counts"	:30			// 签到人数
		"avg_chk_time"	:"00:50:31"	// 平均签到耗时
	}

	307 Temporary Redirect
	{
		"cnmsg":"登录超时，请重新登录",
		"url":"/user"
	}

	400 Bad Request
	{
		"cnmsg":"很抱歉,服务器没有找到该任务"
	}

	403 Forbidden
	{
		"cnmsg":"您的权限不足，无法获取该资源"
	}

#### 查看已完成任务集合人员 [/task/done/gather/mem/{task_id}/{item_counts_per_page}/{cur_page}] [GET]
// 与"查看执行中任务相应人员"类似