// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package jobs

import (
	"fmt"
	"sync"
	"time"

	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
)




type Schedulers struct {
	stop                 chan bool
	stopped              chan bool
	configChanged        chan *model.Config
	clusterLeaderChanged chan bool
	listenerId           string
	startOnce            sync.Once
	jobs                 *JobServer


	// 保存不同类型的 Job 调度器
	schedulers   []model.Scheduler

	// 保存不同类型的 Job 调度器的下一次执行时间
	nextRunTimes []*time.Time
}

func (srv *JobServer) InitSchedulers() *Schedulers {
	mlog.Debug("Initialising schedulers.")

	schedulers := &Schedulers{
		stop:                 make(chan bool),
		stopped:              make(chan bool),
		configChanged:        make(chan *model.Config),
		clusterLeaderChanged: make(chan bool),
		jobs:                 srv,
	}

	// 添加不同类型的 Job 调度器，每个调度器都有一个管道负责接收 job，然后会在 Start() 协程中被定时执行。

	if srv.DataRetentionJob != nil {
		schedulers.schedulers = append(schedulers.schedulers, srv.DataRetentionJob.MakeScheduler())
	}

	if srv.MessageExportJob != nil {
		schedulers.schedulers = append(schedulers.schedulers, srv.MessageExportJob.MakeScheduler())
	}

	if elasticsearchAggregatorInterface := srv.ElasticsearchAggregator; elasticsearchAggregatorInterface != nil {
		schedulers.schedulers = append(schedulers.schedulers, elasticsearchAggregatorInterface.MakeScheduler())
	}

	if ldapSyncInterface := srv.LdapSync; ldapSyncInterface != nil {
		schedulers.schedulers = append(schedulers.schedulers, ldapSyncInterface.MakeScheduler())
	}

	if migrationsInterface := srv.Migrations; migrationsInterface != nil {
		schedulers.schedulers = append(schedulers.schedulers, migrationsInterface.MakeScheduler())
	}

	if pluginsInterface := srv.Plugins; pluginsInterface != nil {
		schedulers.schedulers = append(schedulers.schedulers, pluginsInterface.MakeScheduler())
	}

	schedulers.nextRunTimes = make([]*time.Time, len(schedulers.schedulers))
	return schedulers
}

func (schedulers *Schedulers) Start() *Schedulers {

	schedulers.listenerId = schedulers.jobs.ConfigService.AddConfigListener(schedulers.handleConfigChange)

	go func() {
		schedulers.startOnce.Do(func() {


			mlog.Info("Starting schedulers.")

			defer func() {
				mlog.Info("Schedulers stopped.")
				close(schedulers.stopped)
			}()

			now := time.Now()

			// 遍历调度器
			for idx, scheduler := range schedulers.schedulers {

				// 如果当前调度器被禁用，就置空 nextRunTime，下次不予执行。
				if !scheduler.Enabled(schedulers.jobs.Config()) {
					schedulers.nextRunTimes[idx] = nil

				// 否则，就设置下次执行时间
				} else {
					schedulers.setNextRunTime(schedulers.jobs.Config(), idx, now, false)
				}
			}


			for {
				select {

				// 监听退出信号
				case <-schedulers.stop:
					mlog.Debug("Schedulers received stop signal.")
					return

				// 每分钟遍历一次所有调度器，有需要执行的就执行它
				case now = <-time.After(1 * time.Minute):

					// 获取全局配置
					cfg := schedulers.jobs.Config()

					// 遍历
					for idx, nextTime := range schedulers.nextRunTimes {

						// 若当前调度器的 nextTime 为空，则忽略，不予执行。
						if nextTime == nil {
							continue
						}

						// 如果 nextTime 已经到达，就该执行当前调度器。
						if time.Now().After(*nextTime) {

							// 获取 idx 的调度器对象
							scheduler := schedulers.schedulers[idx]
							if scheduler != nil {
								// 检查是否被禁用
								if scheduler.Enabled(cfg) {
									// 执行调度，若成功则更新 nextTime 和 状态为 pending（等待执行）。
									if _, err := schedulers.scheduleJob(cfg, scheduler); err != nil {
										mlog.Warn(fmt.Sprintf("Failed to schedule job with scheduler: %v", scheduler.Name()))
										mlog.Error(fmt.Sprint(err))
									} else {
										schedulers.setNextRunTime(cfg, idx, now, true)
									}
								}
							}
						}
					}

				// 配置变更
				case newCfg := <-schedulers.configChanged:
					// 遍历所有调度器，修改关联的配置信息
					for idx, scheduler := range schedulers.schedulers {
						if !scheduler.Enabled(newCfg) {
							schedulers.nextRunTimes[idx] = nil
						} else {
							schedulers.setNextRunTime(newCfg, idx, now, false)
						}
					}
				// leader 变更？
				case isLeader := <-schedulers.clusterLeaderChanged:
					for idx := range schedulers.schedulers {
						if !isLeader {
							schedulers.nextRunTimes[idx] = nil
						} else {
							schedulers.setNextRunTime(schedulers.jobs.Config(), idx, now, false)
						}
					}
				}
			}
		})
	}()

	return schedulers
}


func (schedulers *Schedulers) Stop() *Schedulers {
	mlog.Info("Stopping schedulers.")
	// 触发关闭信号
	close(schedulers.stop)
	// 阻塞等待所有协程退出
	<-schedulers.stopped
	return schedulers
}


// 设置 scheduler 的下次执行时间。
func (schedulers *Schedulers) setNextRunTime(cfg *model.Config, idx int, now time.Time, pendingJobs bool) {

	// 根据 idx 取出对应调度器 scheduler
	scheduler := schedulers.schedulers[idx]

	// 如果未设置调度器的 pendingJobs（等待执行）标识，则检查是否存在 pending 状态的 jobs ，若存在，则 pendingJobs 被置为 true。
	if !pendingJobs {

		// 检查是否存在 `类型为 jobType 且状态为 pending` 的 job(s)，若存在则 pj 为 true，否则 false。
		if pj, err := schedulers.jobs.CheckForPendingJobsByType(scheduler.JobType()); err != nil {
			mlog.Error("Failed to set next job run time: " + err.Error())
			schedulers.nextRunTimes[idx] = nil
			return
		} else {
			pendingJobs = pj
		}
	}

	// ???
	lastSuccessfulJob, err := schedulers.jobs.GetLastSuccessfulJobByType(scheduler.JobType())
	if err != nil {
		mlog.Error("Failed to set next job run time: " + err.Error())
		schedulers.nextRunTimes[idx] = nil
		return
	}

	// 设置下次执行时间
	schedulers.nextRunTimes[idx] = scheduler.NextScheduleTime(cfg, now, pendingJobs, lastSuccessfulJob)

	mlog.Debug(fmt.Sprintf("Next run time for scheduler %v: %v", scheduler.Name(), schedulers.nextRunTimes[idx]))
}


// 执行调度器
func (schedulers *Schedulers) scheduleJob(cfg *model.Config, scheduler model.Scheduler) (*model.Job, *model.AppError) {

	// 检查是否存在 `类型为 jobType 且状态为 pending 的 job(s)`
	pendingJobs, err := schedulers.jobs.CheckForPendingJobsByType(scheduler.JobType())
	if err != nil {
		return nil, err
	}

	// ???
	lastSuccessfulJob, err2 := schedulers.jobs.GetLastSuccessfulJobByType(scheduler.JobType())
	if err2 != nil {
		return nil, err
	}

	// 执行调度
	return scheduler.ScheduleJob(cfg, pendingJobs, lastSuccessfulJob)
}

func (schedulers *Schedulers) handleConfigChange(oldConfig *model.Config, newConfig *model.Config) {
	mlog.Debug("Schedulers received config change.")
	schedulers.configChanged <- newConfig
}

func (schedulers *Schedulers) HandleClusterLeaderChange(isLeader bool) {
	select {
	case schedulers.clusterLeaderChanged <- isLeader:
	default:
		mlog.Debug("Did not send cluster leader change message to schedulers as no schedulers listening to notification channel.")
	}
}
