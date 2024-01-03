package job

type CommonJob struct {
	CronExpr string `json:"cronExpr"`
	Name     string `json:"name"`
}

func (commonjob *CommonJob) GetName() string {
	return commonjob.Name
}

func (commonjob *CommonJob) GetCron() string {
	return commonjob.CronExpr
}
