// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package jobs

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
)

// Default polling interval for jobs termination.
// (Defining as `var` rather than `const` allows tests to lower the interval.)
var DEFAULT_WATCHER_POLLING_INTERVAL = 15000 // 15000 毫秒 = 15 秒

type Watcher struct {
	srv     *JobServer
	workers *Workers

	stop            chan struct{}
	stopped         chan struct{}
	pollingInterval int
}

func (srv *JobServer) MakeWatcher(workers *Workers, pollingInterval int) *Watcher {
	return &Watcher{
		stop:            make(chan struct{}),
		stopped:         make(chan struct{}),
		pollingInterval: pollingInterval,
		workers:         workers,
		srv:             srv,
	}
}

func (watcher *Watcher) Start() {
	mlog.Debug("Watcher Started")

	// Delay for some random number of milliseconds before starting to ensure that multiple
	// instances of the jobserver  don't poll at a time too close to each other.
	rand.Seed(time.Now().UTC().UnixNano())

	// 随机等待 [0s, 15s] 一段时间
	<-time.After(time.Duration(rand.Intn(watcher.pollingInterval)) * time.Millisecond)

	defer func() {
		mlog.Debug("Watcher Finished")
		close(watcher.stopped)
	}()

	for {
		select {
		// 监听退出信号
		case <-watcher.stop:
			mlog.Debug("Watcher: Received stop signal")
			return

		// 每 15 秒执行一次
		case <-time.After(time.Duration(watcher.pollingInterval) * time.Millisecond):
			watcher.PollAndNotify()
		}
	}
}

func (watcher *Watcher) Stop() {
	mlog.Debug("Watcher Stopping")
	close(watcher.stop)
	<-watcher.stopped
}



// 1. 获取所有 PENDING 状态的 Jobs
// 2. 遍历这些 Jobs，根据 Job 类型发送给对应的管道，交由对应的协程处理。
func (watcher *Watcher) PollAndNotify() {

	// 获取所有 PENDING 状态的 Jobs
	jobs, err := watcher.srv.Store.Job().GetAllByStatus(model.JOB_STATUS_PENDING)
	if err != nil {
		mlog.Error(fmt.Sprintf("Error occurred getting all pending statuses: %v", err.Error()))
		return
	}

	// 遍历这些 Jobs，根据 Job 类型发送给对应的管道，交由对应的协程处理。
	for _, job := range jobs {

		if job.Type == model.JOB_TYPE_DATA_RETENTION {

			if watcher.workers.DataRetention != nil {
				select {
				case watcher.workers.DataRetention.JobChannel() <- *job:
				default:
				}
			}

		} else if job.Type == model.JOB_TYPE_MESSAGE_EXPORT {
			if watcher.workers.MessageExport != nil {
				select {
				case watcher.workers.MessageExport.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_ELASTICSEARCH_POST_INDEXING {
			if watcher.workers.ElasticsearchIndexing != nil {
				select {
				case watcher.workers.ElasticsearchIndexing.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_ELASTICSEARCH_POST_AGGREGATION {
			if watcher.workers.ElasticsearchAggregation != nil {
				select {
				case watcher.workers.ElasticsearchAggregation.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_LDAP_SYNC {
			if watcher.workers.LdapSync != nil {
				select {
				case watcher.workers.LdapSync.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_MIGRATIONS {
			if watcher.workers.Migrations != nil {
				select {
				case watcher.workers.Migrations.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_PLUGINS {
			if watcher.workers.Plugins != nil {
				select {
				case watcher.workers.Plugins.JobChannel() <- *job:
				default:
				}
			}
		}
	}
}
