package cron

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenIMSDK/tools/log"
	"github.com/openimsdk/open-im-server/v3/internal/tools/cron/driver"
	ps "github.com/openimsdk/open-im-server/v3/internal/tools/cron/persist"
	"github.com/robfig/cron/v3"
)

const (
	defaultReplicas = 50
	defaultDuration = 3 * time.Second

	dcronRunning = 1
	dcronStopped = 0

	dcronStateSteady  = "dcronStateSteady"
	dcronStateUpgrade = "dcronStateUpgrade"
)

type RecoverFuncType func(d *Dcron)

type Dcron struct {
	jobs      map[string]*JobWarpper
	jobsRWMut sync.Mutex

	ServerName string
	nodePool   INodePool
	running    int32

	nodeUpdateDuration time.Duration
	hashReplicas       int

	cr        *cron.Cron
	crOptions []cron.Option

	RecoverFunc RecoverFuncType

	recentJobs IRecentJobPacker
	state      atomic.Value

	persistJob ps.PersistJob
}

func NewDcron(serverName string, driver driver.DriverV2, cronOpts ...cron.Option) *Dcron {
	dcron := newDcron(serverName)
	dcron.crOptions = cronOpts
	dcron.cr = cron.New(cronOpts...)
	dcron.running = dcronStopped
	dcron.nodePool = NewNodePool(serverName, driver, dcron.nodeUpdateDuration, dcron.hashReplicas)
	return dcron
}

func NewDcronWithOption(serverName string, driver driver.DriverV2, dcronOpts ...Option) *Dcron {
	dcron := newDcron(serverName)
	for _, opt := range dcronOpts {
		opt(dcron)
	}

	dcron.cr = cron.New(dcron.crOptions...)
	dcron.nodePool = NewNodePool(serverName, driver, dcron.nodeUpdateDuration, dcron.hashReplicas)
	return dcron
}

func newDcron(serverName string) *Dcron {
	return &Dcron{
		ServerName:         serverName,
		jobs:               make(map[string]*JobWarpper),
		crOptions:          make([]cron.Option, 0),
		nodeUpdateDuration: defaultDuration,
		hashReplicas:       defaultReplicas,
	}
}

// AddJob  add a job
func (d *Dcron) AddJob(jobName, cronStr string, job Job) (err error) {
	return d.addJob(jobName, cronStr, nil, job)
}

// AddFunc add a cron func
func (d *Dcron) AddFunc(jobName, cronStr string, cmd func()) (err error) {
	return d.addJob(jobName, cronStr, cmd, nil)
}

func (d *Dcron) addJob(jobName, cronStr string, cmd func(), job Job) (err error) {
	log.ZInfo(context.Background(), "addJob", "jobName", jobName, "cronStr", cronStr)

	d.jobsRWMut.Lock()
	defer d.jobsRWMut.Unlock()
	if _, ok := d.jobs[jobName]; ok {
		return errors.New("jobName already exist")
	}
	innerJob := JobWarpper{
		Name:    jobName,
		CronStr: cronStr,
		Func:    cmd,
		Job:     job,
		Dcron:   d,
	}
	entryID, err := d.cr.AddJob(cronStr, innerJob)
	if d.persistJob != nil {
		if cjob, ok := (job).(ps.StableJob); ok {
			d.persistJob.AddJob(jobName, cjob)
		}
	}
	if err != nil {
		return err
	}
	innerJob.ID = entryID
	d.jobs[jobName] = &innerJob
	return nil
}

// Remove Job
func (d *Dcron) Remove(jobName string) {
	d.jobsRWMut.Lock()
	defer d.jobsRWMut.Unlock()

	if job, ok := d.jobs[jobName]; ok {
		delete(d.jobs, jobName)
		d.cr.Remove(job.ID)
		if d.persistJob != nil {
			d.persistJob.RemoveJob(jobName)
		}
	}
}

func (d *Dcron) allowThisNodeRun(jobName string) (ok bool) {
	ok, err := d.nodePool.CheckJobAvailable(jobName)
	if err != nil {
		log.ZError(context.Background(), "allow this node run error, err", err)
		ok = false
		d.state.Store(dcronStateUpgrade)
	} else {
		d.state.Store(dcronStateSteady)
		if d.recentJobs != nil {
			go d.reRunRecentJobs(d.recentJobs.PopAllJobs())
		}
	}
	if d.recentJobs != nil {
		if d.state.Load().(string) == dcronStateUpgrade {
			d.recentJobs.AddJob(jobName, time.Now())
		}
	}
	return
}

// Start job
func (d *Dcron) Start() {
	// recover jobs before starting
	if d.RecoverFunc != nil {
		d.RecoverFunc(d)
	}
	if atomic.CompareAndSwapInt32(&d.running, dcronStopped, dcronRunning) {
		if err := d.startNodePool(); err != nil {
			atomic.StoreInt32(&d.running, dcronStopped)
			return
		}
		d.cr.Start()
		log.ZInfo(context.Background(), "dcron started", "nodeID", d.nodePool.GetNodeID())
	} else {
		log.ZInfo(context.Background(), "dcron have started")
	}
}

// Run Job
func (d *Dcron) Run() {
	// recover jobs before starting
	if d.RecoverFunc != nil {
		d.RecoverFunc(d)
	}
	if atomic.CompareAndSwapInt32(&d.running, dcronStopped, dcronRunning) {
		if err := d.startNodePool(); err != nil {
			atomic.StoreInt32(&d.running, dcronStopped)
			return
		}
		log.ZInfo(context.Background(), "dcron running", "nodeID", d.nodePool.GetNodeID())
		d.cr.Run()
	} else {
		log.ZInfo(context.Background(), "dcron already running")
	}
}

func (d *Dcron) startNodePool() error {
	if err := d.nodePool.Start(context.Background()); err != nil {
		log.ZError(context.Background(), "dcron start node pool error", err)
		return err
	}
	return nil
}

// Stop job
func (d *Dcron) Stop() (ctx context.Context) {
	tick := time.NewTicker(time.Millisecond)
	ctx = context.Background()
	d.nodePool.Stop(ctx)
	for range tick.C {
		if atomic.CompareAndSwapInt32(&d.running, dcronRunning, dcronStopped) {
			ctx = d.cr.Stop()
			log.ZInfo(context.Background(), "dcron stopped")
			return
		}
	}
	return
}

func (d *Dcron) reRunRecentJobs(jobNames []string) {
	log.ZInfo(context.Background(), "reRunRecentJobs", "length", len(jobNames))
	for _, jobName := range jobNames {
		if job, ok := d.jobs[jobName]; ok {
			if ok, _ := d.nodePool.CheckJobAvailable(jobName); ok {
				job.Execute()
			}
		}
	}
}

func (d *Dcron) NodeID() string {
	return d.nodePool.GetNodeID()
}
