package v1

import (
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DelayJob
type DelayJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec DelayJobSpec `json:"spec,omitempty"`
	Status DelayJobStatus `json:"status,omitempty"`
}

type DelayJobSpec struct {
	// 执行时间
	ExecutionTime int64 `json:"execution_time,omitempty"`
	// Specifies the job that will be created when executing a CronJob.
	JobTemplate v1.JobTemplateSpec `json:"jobTemplate,omitempty"`
}

type DelayJobStatus struct {
	StartTime *metav1.Time `json:"startTime,omitempty"`
	Succeeded bool `json:"succeeded,omitempty"`
	Message   string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DelayJobList
type DelayJobList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty"`
	// items is the list of DelayJob.
	Items []DelayJob `json:"items"`
}
