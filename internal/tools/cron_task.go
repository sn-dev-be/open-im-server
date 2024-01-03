// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tools

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/OpenIMSDK/protocol/constant"
	pbcron "github.com/OpenIMSDK/protocol/cron"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"github.com/OpenIMSDK/tools/log"

	"github.com/OpenIMSDK/tools/utils"

	dcron "github.com/openimsdk/open-im-server/v3/internal/tools/cron"
	"github.com/openimsdk/open-im-server/v3/internal/tools/cron/driver"
	"github.com/openimsdk/open-im-server/v3/internal/tools/cron/persist"
	"github.com/openimsdk/open-im-server/v3/internal/tools/job"
	"github.com/openimsdk/open-im-server/v3/internal/tools/msg"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/startrpc"
)

type cronServer struct {
	dcron   *dcron.Dcron
	msgTool *msg.MsgTool
}

func StartTask(rpcPort, prometheusPort int) error {
	fmt.Println("cron task start, config", config.Config.ChatRecordsClearTime)
	msgTool, err := msg.InitMsgTool()
	if err != nil {
		return err
	}
	msgTool.ConvertTools()

	rdb, err := cache.NewRedis()
	if err != nil {
		return err
	}

	cronSever := &cronServer{
		msgTool: msgTool,
	}

	persistJob := persist.NewRedisPersist(rdb)
	dcron := dcron.NewDcronWithOption(
		config.Config.RpcRegisterName.OpenImCronName,
		driver.NewRedisDriver(rdb),
		// dcron.CronOptionSeconds(),
		dcron.WithPersist(persistJob),
		dcron.WithClusterStable(time.Second*30),
		dcron.WithRecoverFunc(func(d *dcron.Dcron) {
			jobs, err := persistJob.RecoverAllJob()
			if err != nil {
				log.ZError(context.Background(), "recover all job failed", err)
				panic(err)
			}
			cronSever.recoverAllStableJob(jobs)
		}),
	)
	cronSever.dcron = dcron

	go func() {
		err := startrpc.Start(
			rpcPort,
			config.Config.RpcRegisterName.OpenImCronName,
			prometheusPort,
			cronSever.registerRpc,
		)
		if err != nil {
			panic(utils.Wrap1(err))
		}
	}()

	// register cron tasks
	// log.ZInfo(context.Background(), "start chatRecordsClearTime cron task", "cron config", config.Config.ChatRecordsClearTime)
	// err = dcron.AddFunc("cron_clear_msg_and_fix_seq", config.Config.ChatRecordsClearTime, msgTool.AllConversationClearMsgAndFixSeq)
	// if err != nil {
	// 	log.ZError(context.Background(), "start allConversationClearMsgAndFixSeq cron failed", err)
	// 	panic(err)
	// }
	//
	// log.ZInfo(context.Background(), "start msgDestruct cron task", "cron config", config.Config.MsgDestructTime)
	// err = dcron.AddFunc("cron_conversations_destruct_msgs", config.Config.MsgDestructTime, msgTool.ConversationsDestructMsgs)
	// if err != nil {
	// 	log.ZError(context.Background(), "start conversationsDestructMsgs cron failed", err)
	// 	panic(err)
	// }

	// start crontab
	dcron.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigs

	// stop crontab, Wait for the running task to exit.
	ctx := dcron.Stop()

	select {
	case <-ctx.Done():
	// graceful exit

	case <-time.After(15 * time.Second):
		// forced exit on timeout
	}

	return nil
}

func (c *cronServer) registerRpc(disCov discoveryregistry.SvcDiscoveryRegistry, server *grpc.Server) error {
	pbcron.RegisterCronServer(server, c)
	return nil
}

func (c *cronServer) recoverAllStableJob(jobs map[string]string) error {
	log.ZInfo(context.Background(), "sizeof stablejobs", "containers", len(jobs))
	for jobName, v := range jobs {
		log.ZInfo(context.Background(), "recover", "jobName", jobName, "jobBoyd", v)
		clearMsgJob := job.ClearMsgJob{}
		clearMsgJob.MsgTool = c.msgTool
		err := clearMsgJob.UnSerialize([]byte(v))
		if err != nil {
			log.ZError(context.Background(), "unserialize job error", err)
			continue
		}

		err = c.dcron.AddJob(jobName, clearMsgJob.GetCron(), &clearMsgJob)
		if err != nil {
			log.ZError(context.Background(), "add job error", err)
			continue
		}
	}
	return nil
}

func (c *cronServer) AddClearMsgJob(ctx context.Context, req *pbcron.AddClearMsgJobReq) (*pbcron.AddClearMsgJobResp, error) {
	resp := &pbcron.AddClearMsgJobResp{}
	job := job.NewClearMsgJob(req.ConversationID, getCronExpr(req.CronCycle), c.msgTool)
	// job := job.NewClearMsgJob(req.ConversationID, "1 * * * * *", c.msgTool)
	err := c.dcron.AddJob(job.Name, job.CronExpr, job)
	log.ZInfo(ctx, "add job", "jobName", job.Name)
	if err != nil {
		log.ZError(ctx, "add Job failed", err, "jobName", job.Name)
	}
	return resp, err
}

func (c *cronServer) RemoveClearMsgJob(ctx context.Context, req *pbcron.AddClearMsgJobReq) (*pbcron.AddClearMsgJobResp, error) {
	resp := &pbcron.AddClearMsgJobResp{}
	job := job.NewClearMsgJob(req.ConversationID, "", nil)
	c.dcron.Remove(job.Name)
	log.ZInfo(ctx, "remove job", "jobName", job.Name)
	return resp, nil
}

// netlock redis lock.
func netlock(rdb redis.UniversalClient, key string, ttl time.Duration) bool {
	value := "used"
	ok, err := rdb.SetNX(context.Background(), key, value, ttl).Result() // nolint
	if err != nil {
		// when err is about redis server, return true.
		return false
	}

	return ok
}

func cronWrapFunc(rdb redis.UniversalClient, key string, fn func()) func() {
	enableCronLocker := config.Config.EnableCronLocker
	return func() {
		// if don't enable cron-locker, call fn directly.
		if !enableCronLocker {
			fn()
			return
		}

		// when acquire redis lock, call fn().
		if netlock(rdb, key, 5*time.Second) {
			fn()
		}
	}
}

func getCronExpr(cycle int32) (expr string) {
	now := time.Unix(time.Now().Unix(), 0)
	switch cycle {
	case constant.CrontabDay:
		expr = fmt.Sprintf("%d %d * * *", now.Minute(), now.Hour())
	case constant.CrontabWeek:
		expr = fmt.Sprintf("%d %d */7 * *", now.Minute(), now.Hour())
	case constant.CrontabHalfMonth:
		expr = fmt.Sprintf("%d %d %d,*/15 * *", now.Minute(), now.Hour(), now.Day())
	case constant.CrontabMonth:
		expr = fmt.Sprintf("%d %d %d * *", now.Minute(), now.Hour(), now.Day())
	}
	return expr
}
