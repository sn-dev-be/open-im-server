package driver

import (
	"time"

	"github.com/google/uuid"
)

// GlobalKeyPrefix is global redis key preifx
const GlobalKeyPrefix = "distributed-cron:"

func GetKeyPre(serviceName string) string {
	return GlobalKeyPrefix + serviceName + ":"
}

func GetNodeId(serviceName string) string {
	return GetKeyPre(serviceName) + uuid.New().String()
}

func GetStableJobStore() string {
	return GlobalKeyPrefix + "stable-jobs"
}

func GetStableJobStoreTxKey() string {
	return GlobalKeyPrefix + "TX:stable-jobs"
}

func TimePre(t time.Time, preDuration time.Duration) int64 {
	return t.Add(-preDuration).Unix()
}
