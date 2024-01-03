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

const (
	redisDefaultTimeout = 5 * time.Second
)

type RedisDriver struct {
	c           redis.UniversalClient
	serviceName string
	nodeID      string
	timeout     time.Duration
	started     bool

	// this context is used to define
	// the life time of this driver.
	runtimeCtx    context.Context
	runtimeCancel context.CancelFunc

	sync.Mutex
}

func newRedisDriver(redisClient redis.UniversalClient) *RedisDriver {
	rd := &RedisDriver{
		c:       redisClient,
		timeout: redisDefaultTimeout,
	}
	rd.started = false
	return rd
}

func (rd *RedisDriver) Init(serviceName string, opts ...Option) {
	rd.serviceName = serviceName
	rd.nodeID = GetNodeId(rd.serviceName)

	for _, opt := range opts {
		rd.withOption(opt)
	}
}

func (rd *RedisDriver) NodeID() string {
	return rd.nodeID
}

func (rd *RedisDriver) Start(ctx context.Context) (err error) {
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

func (rd *RedisDriver) Stop(ctx context.Context) (err error) {
	rd.Lock()
	defer rd.Unlock()
	rd.runtimeCancel()
	rd.started = false
	return
}

func (rd *RedisDriver) GetNodes(ctx context.Context) (nodes []string, err error) {
	mathStr := fmt.Sprintf("%s*", GetKeyPre(rd.serviceName))
	return rd.scan(ctx, mathStr)
}

// private function

func (rd *RedisDriver) heartBeat() {
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
				if err := rd.c.Del(context.Background(), rd.nodeID, rd.nodeID).Err(); err != nil {
					log.ZError(context.Background(), "unregister service node error", err)
				}
				return
			}
		}
	}
}

func (rd *RedisDriver) registerServiceNode() error {
	return rd.c.SetEx(context.Background(), rd.nodeID, rd.nodeID, rd.timeout).Err()
}

func (rd *RedisDriver) scan(ctx context.Context, matchStr string) ([]string, error) {
	ret := make([]string, 0)
	iter := rd.c.Scan(ctx, 0, matchStr, -1).Iterator()
	for iter.Next(ctx) {
		err := iter.Err()
		if err != nil {
			return nil, err
		}
		ret = append(ret, iter.Val())
	}
	return ret, nil
}

func (rd *RedisDriver) withOption(opt Option) (err error) {
	switch opt.Type() {
	case OptionTypeTimeout:
		rd.timeout = opt.(TimeoutOption).timeout
	}
	return
}
