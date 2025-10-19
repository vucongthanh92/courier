package cronjob

type CronJobImpl struct {
}

func NewCronJobService() CronJobService {
	return &CronJobImpl{}
}
