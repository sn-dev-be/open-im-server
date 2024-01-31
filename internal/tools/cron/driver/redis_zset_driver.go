package driver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/OpenIMSDK/tools/log"
	"github.com/redis/go-redis/v9"
)

type RedisZSetDriver struct {
	c           redis.UniversalClient
	serviceName string
	nodeID      string
	timeout     time.Duration
	started     bool

	// this context is used to define
	// the lifetime of this driver.
	runtimeCtx    context.Context
	runtimeCancel context.CancelFunc

	sync.Mutex
}

func newRedisZSetDriver(redisClient redis.UniversalClient) *RedisZSetDriver {
	rd := &RedisZSetDriver{
		c:       redisClient,
		timeout: redisDefaultTimeout,
	}
	rd.started = false
	return rd
}

func (rd *RedisZSetDriver) Init(serviceName string, opts ...Option) {
	rd.serviceName = serviceName
	rd.nodeID = GetNodeId(serviceName)
	// if err := rd.c.Del(context.Background(), GetKeyPre(rd.serviceName)).Err(); err != nil {
	// 	log.ZError(context.Background(), "remove all nodes err", err)
	// }
	for _, opt := range opts {
		rd.withOption(opt)
	}
}

func (rd *RedisZSetDriver) NodeID() string {
	return rd.nodeID
}

func (rd *RedisZSetDriver) GetNodes(ctx context.Context) (nodes []string, err error) {
	rd.Lock()
	defer rd.Unlock()
	sliceCmd := rd.c.ZRangeByScore(ctx, GetKeyPre(rd.serviceName), &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", TimePre(time.Now(), rd.timeout)),
		Max: "+inf",
	})
	if err = sliceCmd.Err(); err != nil {
		return nil, err
	} else {
		nodes = make([]string, len(sliceCmd.Val()))
		copy(nodes, sliceCmd.Val())
	}
	log.ZInfo(context.Background(), "get nodes", "nodes", nodes)
	return
}
func (rd *RedisZSetDriver) Start(ctx context.Context) (err error) {
	rd.Lock()
	defer rd.Unlock()
	if rd.started {
		err = errors.New("this driver is started")
		return
	}
	rd.runtimeCtx, rd.runtimeCancel = context.WithCancel(context.TODO())
	rd.started = true
	// register
	err = rd.registerServiceNode()
	if err != nil {
		log.ZError(context.Background(), "register service error", err)
		return
	}
	// heartbeat timer
	go rd.heartBeat()
	return
}
func (rd *RedisZSetDriver) Stop(ctx context.Context) (err error) {
	rd.Lock()
	defer rd.Unlock()
	rd.runtimeCancel()
	rd.started = false
	return
}

func (rd *RedisZSetDriver) withOption(opt Option) (err error) {
	switch opt.Type() {
	case OptionTypeTimeout:
		{
			rd.timeout = opt.(TimeoutOption).timeout
		}
	}
	return
}

// private function

func (rd *RedisZSetDriver) heartBeat() {
	tick := time.NewTicker(rd.timeout / 2)
	for {
		select {
		case <-tick.C:
			{
				if err := rd.registerServiceNode(); err != nil {
					log.ZError(context.Background(), "register service node error", err)
				}
			}
		case <-rd.runtimeCtx.Done():
			{
				if err := rd.c.ZRem(context.Background(), GetKeyPre(rd.serviceName), rd.nodeID).Err(); err != nil {
					log.ZError(context.Background(), "unregister service node error", err)
				}
				return
			}
		}
	}
}

func (rd *RedisZSetDriver) registerServiceNode() error {
	return rd.c.ZAdd(context.Background(), GetKeyPre(rd.serviceName), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: rd.nodeID,
	}).Err()
}
