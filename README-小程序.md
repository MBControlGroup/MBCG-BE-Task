# 任务管理-小程序
这是小程序端的接口说明

## 新任务
### 获取新任务 [/task/miniprogram/new-task] [GET]
### 接受任务 [/task/miniprogram/accept] [PUT]
### 拒绝任务 [/task/miniprogram/refuse] [PUT]

## 执行中的任务
### 获取执行中的任务 [/task/miniprogram/working] [GET]
### 获取接受该任务的队友 [/task/miniprogram/soldiers/{task_id}] [GET]
返回同组织/单位的接受该任务的队友和负责人
### 签到 [/task/miniprogram/check] [PUT]
request: task_id, place_lat, place_lng

## 已完成任务
### 提交任务总结 [/task/miniprogram/summary] [POST]
request: task_id, summary
### 查看已完成任务列表 [/task/miniprogram/done] [GET]
### 查看已完成任务详情 [/task/miniprogram/done/{task_id} [GET]