package cron

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenIMSDK/tools/log"
	"github.com/openimsdk/open-im-server/v3/internal/tools/cron/consistenthash"
	"github.com/openimsdk/open-im-server/v3/internal/tools/cron/driver"
)

const (
	NodePoolStateSteady  = "NodePoolStateSteady"
	NodePoolStateUpgrade = "NodePoolStateUpgrade"
)

// NodePool
// For cluster steable.
// NodePool has 2 states:
//  1. Steady
//     If this nodePoolLists is the same as the last update,
//     we will mark this node's state to Steady. In this state,
//     this node can run jobs.
//  2. Upgrade
//     If this nodePoolLists is different to the last update,
//     we will mark this node's state to Upgrade. In this state,
//     this node can not run jobs.
type NodePool struct {
	serviceName string
	nodeID      string

	rwMut sync.RWMutex
	nodes *consistenthash.Map

	driver         driver.DriverV2
	hashReplicas   int
	hashFn         consistenthash.Hash
	updateDuration time.Duration

	stopChan chan int
	preNodes []string // sorted

	lastUpdateNodesTime atomic.Value
	state               atomic.Value
}

func NewNodePool(
	serviceName string,
	drv driver.DriverV2,
	updateDuration time.Duration,
	hashReplicas int,
) INodePool {
	np := &NodePool{
		serviceName:    serviceName,
		driver:         drv,
		hashReplicas:   hashReplicas,
		updateDuration: updateDuration,
		stopChan:       make(chan int, 1),
	}
	np.driver.Init(serviceName, driver.NewTimeoutOption(updateDuration))
	return np
}

func (np *NodePool) Start(ctx context.Context) (err error) {
	err = np.driver.Start(ctx)
	if err != nil {
		log.ZError(context.Background(), "start pool error", err)
		return
	}
	np.nodeID = np.driver.NodeID()
	nowNodes, err := np.driver.GetNodes(ctx)
	if err != nil {
		log.ZError(context.Background(), "get nodes error", err)
		return
	}
	np.state.Store(NodePoolStateUpgrade)
	np.updateHashRing(nowNodes)
	go np.waitingForHashRing()

	// stuck util the cluster state came to steady.
	for np.getState() != NodePoolStateSteady {
		<-time.After(np.updateDuration)
	}
	log.ZInfo(context.Background(), "nodepool started for serve", "nodeID", np.nodeID)

	return
}

// Check if this job can be run in this node.
func (np *NodePool) CheckJobAvailable(jobName string) (bool, error) {
	np.rwMut.RLock()
	defer np.rwMut.RUnlock()
	if np.nodes == nil {
		log.ZWarn(context.Background(), "NodePool.nodes is nil", nil, "nodeID", np.nodeID)
	}
	if np.nodes.IsEmpty() {
		return false, nil
	}
	if np.state.Load().(string) != NodePoolStateSteady {
		return false, ErrNodePoolIsUpgrading
	}
	targetNode := np.nodes.Get(jobName)
	log.ZDebug(context.Background(), "checkJobAvailable", "jobName", jobName, "targetNode", targetNode, "nodeID", np.nodeID)
	if np.nodeID == targetNode {
		log.ZInfo(context.Background(), "check job available", "job", jobName, "running in node:", targetNode)
	}
	return np.nodeID == targetNode, nil
}

func (np *NodePool) Stop(ctx context.Context) error {
	np.stopChan <- 1
	np.driver.Stop(ctx)
	np.preNodes = make([]string, 0)
	return nil
}

func (np *NodePool) GetNodeID() string {
	return np.nodeID
}

func (np *NodePool) GetLastNodesUpdateTime() time.Time {
	return np.lastUpdateNodesTime.Load().(time.Time)
}

func (np *NodePool) getState() string {
	return np.state.Load().(string)
}

func (np *NodePool) waitingForHashRing() {
	tick := time.NewTicker(np.updateDuration)
	for {
		select {
		case <-tick.C:
			nowNodes, err := np.driver.GetNodes(context.Background())
			if err != nil {
				log.ZError(context.Background(), "get nodes error", err)
				continue
			}
			np.updateHashRing(nowNodes)
		case <-np.stopChan:
			return
		}
	}
}

func (np *NodePool) updateHashRing(nodes []string) {
	np.rwMut.Lock()
	defer np.rwMut.Unlock()
	if np.equalRing(nodes) {
		np.state.Store(NodePoolStateSteady)
		log.ZDebug(context.Background(), "update hashRing", "nowNodes", nodes, "preNodes", np.preNodes)
		return
	}
	np.lastUpdateNodesTime.Store(time.Now())
	np.state.Store(NodePoolStateUpgrade)
	log.ZDebug(context.Background(), "update hashRing", "nodes", nodes)
	np.preNodes = make([]string, len(nodes))
	copy(np.preNodes, nodes)
	np.nodes = consistenthash.New(np.hashReplicas, np.hashFn)
	for _, v := range nodes {
		np.nodes.Add(v)
	}
}

func (np *NodePool) equalRing(a []string) bool {
	if len(a) == len(np.preNodes) {
		la := len(a)
		sort.Strings(a)
		for i := 0; i < la; i++ {
			if a[i] != np.preNodes[i] {
				return false
			}
		}
		return true
	}
	return false
}
