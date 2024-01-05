package persist

import (
	"context"

	"github.com/openimsdk/open-im-server/v3/internal/tools/cron/driver"
	"github.com/redis/go-redis/v9"
)

type RedisPersist struct {
	redisClient redis.UniversalClient
}

func NewRedisPersist(client redis.UniversalClient) *RedisPersist {
	return &RedisPersist{
		redisClient: client,
	}
}

func (r *RedisPersist) AddJob(jobName string, job StableJob) error {
	bytes, err := job.Serialize()
	if err != nil {
		return err
	}
	_, err = r.redisClient.HSet(context.Background(), driver.GetStableJobStore(), job.GetName(), bytes).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisPersist) RemoveJob(jobName string) error {
	_, err := r.redisClient.HDel(context.Background(), driver.GetStableJobStore(), jobName).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisPersist) GetJob(jobName string) (string, error) {
	return r.redisClient.HGet(context.Background(), driver.GetStableJobStore(), jobName).Result()
}

func (r *RedisPersist) RecoverAllJob() (map[string]string, error) {
	return r.redisClient.HGetAll(context.Background(), driver.GetStableJobStore()).Result()
}
