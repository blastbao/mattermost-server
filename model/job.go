// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (

	// JOB 类型
	JOB_TYPE_DATA_RETENTION                 = "data_retention"
	JOB_TYPE_MESSAGE_EXPORT                 = "message_export"
	JOB_TYPE_ELASTICSEARCH_POST_INDEXING    = "elasticsearch_post_indexing"
	JOB_TYPE_ELASTICSEARCH_POST_AGGREGATION = "elasticsearch_post_aggregation"
	JOB_TYPE_LDAP_SYNC                      = "ldap_sync"
	JOB_TYPE_MIGRATIONS                     = "migrations"
	JOB_TYPE_PLUGINS                        = "plugins"

	// JOB 状态
	JOB_STATUS_PENDING          = "pending"				//等待执行
	JOB_STATUS_IN_PROGRESS      = "in_progress" 		//正在执行
	JOB_STATUS_SUCCESS          = "success" 			//执行成功
	JOB_STATUS_ERROR            = "error" 				//执行失败
	JOB_STATUS_CANCEL_REQUESTED = "cancel_requested" 	//正在取消
	JOB_STATUS_CANCELED         = "canceled" 			//已经取消
)


type Job struct {
	Id             string            `json:"id"`				//ID
	Type           string            `json:"type"`				//类型
	Priority       int64             `json:"priority"`			//优先级
	CreateAt       int64             `json:"create_at"`			//创建时间
	StartAt        int64             `json:"start_at"`			//开始时间
	LastActivityAt int64             `json:"last_activity_at"`	//最近活跃时间
	Status         string            `json:"status"`			//状态
	Progress       int64             `json:"progress"`			//执行进度
	Data           map[string]string `json:"data"` 				//附加数据
}

func (j *Job) IsValid() *AppError {
	// ID 为 26 位长度字符串
	if len(j.Id) != 26 {
		return NewAppError("Job.IsValid", "model.job.is_valid.id.app_error", nil, "id="+j.Id, http.StatusBadRequest)
	}

	// 创建时间
	if j.CreateAt == 0 {
		return NewAppError("Job.IsValid", "model.job.is_valid.create_at.app_error", nil, "id="+j.Id, http.StatusBadRequest)
	}

	// 类型检查
	switch j.Type {
	case JOB_TYPE_DATA_RETENTION:
	case JOB_TYPE_ELASTICSEARCH_POST_INDEXING:
	case JOB_TYPE_ELASTICSEARCH_POST_AGGREGATION:
	case JOB_TYPE_LDAP_SYNC:
	case JOB_TYPE_MESSAGE_EXPORT:
	case JOB_TYPE_MIGRATIONS:
	case JOB_TYPE_PLUGINS:
	default:
		return NewAppError("Job.IsValid", "model.job.is_valid.type.app_error", nil, "id="+j.Id, http.StatusBadRequest)
	}

	// 状态检查
	switch j.Status {
	case JOB_STATUS_PENDING:
	case JOB_STATUS_IN_PROGRESS:
	case JOB_STATUS_SUCCESS:
	case JOB_STATUS_ERROR:
	case JOB_STATUS_CANCEL_REQUESTED:
	case JOB_STATUS_CANCELED:
	default:
		return NewAppError("Job.IsValid", "model.job.is_valid.status.app_error", nil, "id="+j.Id, http.StatusBadRequest)
	}

	return nil
}

func (js *Job) ToJson() string {
	b, _ := json.Marshal(js)
	return string(b)
}

func JobFromJson(data io.Reader) *Job {
	var job Job
	if err := json.NewDecoder(data).Decode(&job); err == nil {
		return &job
	} else {
		return nil
	}
}

func JobsToJson(jobs []*Job) string {
	b, _ := json.Marshal(jobs)
	return string(b)
}

func JobsFromJson(data io.Reader) []*Job {
	var jobs []*Job
	if err := json.NewDecoder(data).Decode(&jobs); err == nil {
		return jobs
	} else {
		return nil
	}
}

func (js *Job) DataToJson() string {
	b, _ := json.Marshal(js.Data)
	return string(b)
}

type Worker interface {
	Run()
	Stop()
	JobChannel() chan<- Job
}

type Scheduler interface {
	Name() string
	JobType() string
	Enabled(cfg *Config) bool
	NextScheduleTime(cfg *Config, now time.Time, pendingJobs bool, lastSuccessfulJob *Job) *time.Time
	ScheduleJob(cfg *Config, pendingJobs bool, lastSuccessfulJob *Job) (*Job, *AppError)
}
