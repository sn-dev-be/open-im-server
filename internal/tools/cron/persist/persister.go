package persist

// This type of Job will be
// recovered in a node of service
// restarting.
type StableJob interface {
	Run()
	GetCron() string
	GetName() string
	Serialize() ([]byte, error)
	UnSerialize([]byte) error
}

type PersistJob interface {
	AddJob(jobName string, job StableJob) error
	RemoveJob(jobName string) error
	GetJob(jobName string) (string, error)
	RecoverAllJob() (map[string]string, error)
}
