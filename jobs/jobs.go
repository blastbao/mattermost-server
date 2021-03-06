// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package jobs

import (
	"context"
	"fmt"
	"time"

	"net/http"

	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
)

const (
	CANCEL_WATCHER_POLLING_INTERVAL = 5000
)

func (srv *JobServer) CreateJob(jobType string, jobData map[string]string) (*model.Job, *model.AppError) {

	// 创建 job 对象
	job := model.Job{
		Id:       model.NewId(),
		Type:     jobType,
		CreateAt: model.GetMillis(),
		Status:   model.JOB_STATUS_PENDING,
		Data:     jobData,
	}

	// 参数检查
	if err := job.IsValid(); err != nil {
		return nil, err
	}

	// 保存 job 到 store 中
	if _, err := srv.Store.Job().Save(&job); err != nil {
		return nil, err
	}

	// 返回 job
	return &job, nil
}

func (srv *JobServer) GetJob(id string) (*model.Job, *model.AppError) {
	return srv.Store.Job().Get(id)
}

//
func (srv *JobServer) ClaimJob(job *model.Job) (bool, *model.AppError) {
	return srv.Store.Job().UpdateStatusOptimistically(job.Id, model.JOB_STATUS_PENDING, model.JOB_STATUS_IN_PROGRESS)
}

func (srv *JobServer) SetJobProgress(job *model.Job, progress int64) *model.AppError {
	job.Status = model.JOB_STATUS_IN_PROGRESS
	job.Progress = progress

	// 更新 job 信息和状态
	if _, err := srv.Store.Job().UpdateOptimistically(job, model.JOB_STATUS_IN_PROGRESS); err != nil {
		return err
	}
	return nil
}

func (srv *JobServer) SetJobSuccess(job *model.Job) *model.AppError {
	// 更新 job 状态
	if _, err := srv.Store.Job().UpdateStatus(job.Id, model.JOB_STATUS_SUCCESS); err != nil {
		return err
	}
	return nil
}

func (srv *JobServer) SetJobError(job *model.Job, jobError *model.AppError) *model.AppError {

	if jobError == nil {
		_, err := srv.Store.Job().UpdateStatus(job.Id, model.JOB_STATUS_ERROR)
		return err
	}

	job.Status = model.JOB_STATUS_ERROR
	job.Progress = -1
	if job.Data == nil {
		job.Data = make(map[string]string)
	}
	job.Data["error"] = jobError.Message + " — " + jobError.DetailedError

	// 更新 job 信息和状态
	updated, err := srv.Store.Job().UpdateOptimistically(job, model.JOB_STATUS_IN_PROGRESS)
	if err != nil {
		return err
	}

	// 更新 job 信息和状态
	if !updated {
		updated, err = srv.Store.Job().UpdateOptimistically(job, model.JOB_STATUS_CANCEL_REQUESTED)
		if err != nil {
			return err
		}
		if !updated {
			return model.NewAppError("Jobs.SetJobError", "jobs.set_job_error.update.error", nil, "id="+job.Id, http.StatusInternalServerError)
		}
	}

	return nil
}

func (srv *JobServer) SetJobCanceled(job *model.Job) *model.AppError {
	if _, err := srv.Store.Job().UpdateStatus(job.Id, model.JOB_STATUS_CANCELED); err != nil {
		return err
	}
	return nil
}

func (srv *JobServer) UpdateInProgressJobData(job *model.Job) *model.AppError {
	job.Status = model.JOB_STATUS_IN_PROGRESS
	job.LastActivityAt = model.GetMillis()
	if _, err := srv.Store.Job().UpdateOptimistically(job, model.JOB_STATUS_IN_PROGRESS); err != nil {
		return err
	}
	return nil
}

func (srv *JobServer) RequestCancellation(jobId string) *model.AppError {
	updated, err := srv.Store.Job().UpdateStatusOptimistically(jobId, model.JOB_STATUS_PENDING, model.JOB_STATUS_CANCELED)
	if err != nil {
		return err
	}
	if updated {
		return nil
	}

	updated, err = srv.Store.Job().UpdateStatusOptimistically(jobId, model.JOB_STATUS_IN_PROGRESS, model.JOB_STATUS_CANCEL_REQUESTED)
	if err != nil {
		return err
	}

	if updated {
		return nil
	}

	return model.NewAppError("Jobs.RequestCancellation", "jobs.request_cancellation.status.error", nil, "id="+jobId, http.StatusInternalServerError)
}

func (srv *JobServer) CancellationWatcher(ctx context.Context, jobId string, cancelChan chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			mlog.Debug(fmt.Sprintf("CancellationWatcher for Job: %v Aborting as job has finished.", jobId))
			return
		case <-time.After(CANCEL_WATCHER_POLLING_INTERVAL * time.Millisecond):
			mlog.Debug(fmt.Sprintf("CancellationWatcher for Job: %v polling.", jobId))
			if jobStatus, err := srv.Store.Job().Get(jobId); err == nil {
				if jobStatus.Status == model.JOB_STATUS_CANCEL_REQUESTED {
					close(cancelChan)
					return
				}
			}
		}
	}
}

func GenerateNextStartDateTime(now time.Time, nextStartTime time.Time) *time.Time {
	nextTime := time.Date(now.Year(), now.Month(), now.Day(), nextStartTime.Hour(), nextStartTime.Minute(), 0, 0, time.Local)

	if !now.Before(nextTime) {
		nextTime = nextTime.AddDate(0, 0, 1)
	}

	return &nextTime
}

// 检查是否存在 `类型为 jobType 且状态为 pending 的 job(s)`
func (srv *JobServer) CheckForPendingJobsByType(jobType string) (bool, *model.AppError) {

	// 获取类型为 jobType 的所有 pending 状态的 jobs 总数
	count, err := srv.Store.Job().GetCountByStatusAndType(model.JOB_STATUS_PENDING, jobType)
	if err != nil {
		return false, err
	}
	// 判断是否非零
	return count > 0, nil
}

//
func (srv *JobServer) GetLastSuccessfulJobByType(jobType string) (*model.Job, *model.AppError) {
	return srv.Store.Job().GetNewestJobByStatusAndType(model.JOB_STATUS_SUCCESS, jobType)
}
