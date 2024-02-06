package job

type CommonJob struct {
	CronExpr string `json:"CronExpr"`
	Name     string `json:"Name"`
	Type     int    `json:"Type"`
}

func (commonjob *CommonJob) GetName() string {
	return commonjob.Name
}

func (commonjob *CommonJob) GetCron() string {
	return commonjob.CronExpr
}
