package delayjob

import (
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/delay-job-controller/controller/delayjob/v1"
)

func inActiveList(cj v1.DelayJob, uid types.UID) bool {
	for _, j := range cj.Status.Active {
		if j.UID == uid {
			return true
		}
	}
	return false
}

func IsFinished(cj v1.DelayJob) bool {
	if cj.Status.Status == v1.Finish {
		return true
	}
	return false
}

// byJobStartTimeStar sorts a list of jobs by start timestamp, using their names as a tie breaker.
type byJobStartTimeStar []*batchv1.Job

func (o byJobStartTimeStar) Len() int      { return len(o) }
func (o byJobStartTimeStar) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o byJobStartTimeStar) Less(i, j int) bool {
	if o[i].Status.StartTime == nil && o[j].Status.StartTime != nil {
		return false
	}
	if o[i].Status.StartTime != nil && o[j].Status.StartTime == nil {
		return true
	}
	if o[i].Status.StartTime.Equal(o[j].Status.StartTime) {
		return o[i].Name < o[j].Name
	}
	return o[i].Status.StartTime.Before(o[j].Status.StartTime)
}
