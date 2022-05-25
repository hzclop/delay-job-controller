package controller

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	batchv1informers "k8s.io/client-go/informers/batch/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ref "k8s.io/client-go/tools/reference"
	"k8s.io/client-go/util/workqueue"
	delayjobv1 "k8s.io/delay-job-controller/controller/delayjob/v1"
	"k8s.io/delay-job-controller/log"
	delayclientset "k8s.io/delay-job-controller/pkg/generated/clientset/versioned"
	delayscheme "k8s.io/delay-job-controller/pkg/generated/clientset/versioned/scheme"
	informers "k8s.io/delay-job-controller/pkg/generated/informers/externalversions/delayjob/v1"
	delaylisters "k8s.io/delay-job-controller/pkg/generated/listers/delayjob/v1"
	"reflect"
	"time"
)

const (
	controllerAgentName = "delayjobv1-controller"
)

var (
	controllerKind = delayjobv1.SchemeGroupVersion.WithKind("DelayJob")
)

type DelayJobController struct {
	recorder record.EventRecorder
	queue    workqueue.RateLimitingInterface

	jobControl      jobControlInterface
	delayJobControl djControlInterface

	jobLister            batchv1listers.JobLister
	jobInformerSynced   cache.InformerSynced
	
	delayJobLister      delaylisters.DelayJobLister
	delayInformerSynced cache.InformerSynced
}

func NewController(kubeClient clientset.Interface, delayClient delayclientset.Interface, jobInformer batchv1informers.JobInformer, delayJobInformer informers.DelayJobInformer) (*DelayJobController, error) {
	utilruntime.Must(delayscheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	dm := &DelayJobController{
		jobControl:        &realJobControl{KubeClient: kubeClient},
		delayJobControl:   &realDelayJobControl{KubeClient: delayClient},
		jobLister:         jobInformer.Lister(),
		jobInformerSynced: jobInformer.Informer().HasSynced,

		delayJobLister:      delayJobInformer.Lister(),
		delayInformerSynced: delayJobInformer.Informer().HasSynced,
		queue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "delayjobs"),
		recorder:            recorder,
	}

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dm.addJob,
		UpdateFunc: dm.updateJob,
		DeleteFunc: dm.deleteJob,
	})

	delayJobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			dm.enqueueController(obj)
		},
		UpdateFunc: dm.updateDelayJob,
		DeleteFunc: func(obj interface{}) {
			dm.enqueueController(obj)
		},
	})

	return dm, nil
}

// Run starts the main goroutine responsible for watching and syncing jobs.
func (jm *DelayJobController) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer jm.queue.ShutDown()

	log.Logger().Info("Starting delayJob controller")
	log.Logger().Info("Waiting for informer caches to sync")
	if !cache.WaitForCacheSync(ctx.Done(), jm.delayInformerSynced, jm.jobInformerSynced, ) {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	log.Logger().Info("Starting workers")
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, jm.worker, time.Second)
	}
	log.Logger().Info("Started workers")
	<-ctx.Done()
	return nil
}

func (jm *DelayJobController) worker(ctx context.Context) {
	for jm.processNextWorkItem(ctx) {
	}
}

func (jm *DelayJobController) processNextWorkItem(ctx context.Context) bool {
	key, quit := jm.queue.Get()
	if quit {
		return false
	}
	defer jm.queue.Done(key)

	requeueAfter, err := jm.sync(ctx, key.(string))
	switch {
	case err != nil:
		utilruntime.HandleError(fmt.Errorf("error syncing delayJobController %v, requeuing: %v", key.(string), err))
		jm.queue.AddRateLimited(key)
	case requeueAfter != nil:
		jm.queue.Forget(key)
		jm.queue.AddAfter(key, *requeueAfter)
	}

	return true
}

func (jm *DelayJobController) sync(ctx context.Context, delayJobKey string) (*time.Duration, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(delayJobKey)
	if err != nil {
		return nil, err
	}

	delayJob, err := jm.delayJobControl.GetDelayJob(ctx, ns, name)
	switch {
	case errors.IsNotFound(err):
		// may be delayJobKey is deleted, don't need to requeue this key
		log.Logger().Info("delayJobKey not found, may be it is deleted", "delayJobKey", fmt.Sprint(ns, name), "err", err)
		return nil, nil
	case err != nil:
		// for other transient apiserver error requeue with exponential backoff
		return nil, err
	}

	jobsToBeReconciled, err := jm.getJobsToBeReconciled(delayJob)
	if err != nil {
		return nil, err
	}
	delayJobCopy, requeueAfter, err := jm.syncDelayJob(ctx, delayJob, jobsToBeReconciled)
	if err != nil {
		log.Logger().Info("Error reconciling cronjob ns=%s name=%s err=%s", delayJob.GetNamespace(), delayJob.GetName(), err.Error())
		return nil, err
	}
	err = jm.cleanupFinishedJobs(ctx, delayJobCopy, jobsToBeReconciled)
	if err != nil {
		log.Logger().Info("Error cleaning up jobs cronjob ns=%s name=%s resourceVersion=%s err=%s", delayJob.GetNamespace(), delayJob.GetName(), delayJob.GetResourceVersion(), err.Error())
		return nil, err
	}
	if requeueAfter != nil {
		log.Logger().Info("Re-queuing cronjob ns=%s name=%s requeueAfter=%s", "cronjob", delayJob.GetNamespace(), delayJob.GetName(), requeueAfter)
		return requeueAfter, nil
	}
	return nil, nil
}

func (jm *DelayJobController) syncDelayJob(
	ctx context.Context,
	dj *delayjobv1.DelayJob,
	js []*batchv1.Job) (*delayjobv1.DelayJob, *time.Duration, error) {
	dj = dj.DeepCopy()

	jobReq, err := getJobFromTemplate(dj)
	if err != nil {
		log.Logger().Error("Unable to make Job from template namespace=%s name=%s cause=%s", dj.GetNamespace(), dj.GetName(), err.Error())
		return dj, nil, err
	}
	jobResp, err := jm.jobControl.CreateJob(dj.Namespace, jobReq)
	switch {
	case errors.HasStatusCause(err, corev1.NamespaceTerminatingCause):
	case errors.IsAlreadyExists(err):
		// If the job is created by other actor, assume  it has updated the cronjob status accordingly
		log.Logger().Info("Job already exists delayJob ns=%s name=%s job ns=%s name=%s", dj.GetNamespace(), dj.GetName(), jobReq.GetNamespace(), jobReq.GetName())
		return dj, nil, err
	case err != nil:
		// default error handling
		jm.recorder.Eventf(dj, corev1.EventTypeWarning, "FailedCreate", "Error creating job: %v", err)
		return dj, nil, err
	}
	_, err = getRef(jobResp)
	if err != nil {
		log.Logger().Info("Unable to make object reference delayjob ns=%s name=%s err=%s", dj.GetNamespace(), dj.GetName(), "err", err)
		return dj, nil, fmt.Errorf("unable to make object reference for job for ns=%s name=%s", dj.GetNamespace(), dj.GetName())
	}
	//dj.Status.Active = append(dj.Status.Active, *jobRef)
	if _, err := jm.delayJobControl.UpdateStatus(ctx, dj); err != nil {
		log.Logger().Info("Unable to update status delayjob ns=%s name=%s resourceVersion=%s err=%s", dj.GetNamespace(), dj.GetName(), dj.ResourceVersion, err)
		return dj, nil, fmt.Errorf("unable to update status for %s/%s (rv = %s): %v", dj.GetNamespace(), dj.GetName(), dj.ResourceVersion, err)
	}
	return dj, nil, nil
}

func getRef(object runtime.Object) (*corev1.ObjectReference, error) {
	return ref.GetReference(scheme.Scheme, object)
}


// todo
func (jm *DelayJobController) cleanupFinishedJobs(ctx context.Context, cj *delayjobv1.DelayJob, js []*batchv1.Job) error {
	return nil
}

func (jm *DelayJobController) getJobsToBeReconciled(cronJob *delayjobv1.DelayJob) ([]*batchv1.Job, error) {
	var jobSelector labels.Selector
	if len(cronJob.Spec.JobTemplate.Labels) == 0 {
		jobSelector = labels.Everything()
	} else {
		jobSelector = labels.Set(cronJob.Spec.JobTemplate.Labels).AsSelector()
	}
	jobList, err := jm.jobLister.Jobs(cronJob.Namespace).List(jobSelector)
	if err != nil {
		return nil, err
	}

	jobsToBeReconciled := []*batchv1.Job{}

	for _, job := range jobList {
		// If it has a ControllerRef, that's all that matters.
		if controllerRef := metav1.GetControllerOf(job); controllerRef != nil && controllerRef.Name == cronJob.Name {
			// this job is needs to be reconciled
			jobsToBeReconciled = append(jobsToBeReconciled, job)
		}
	}
	return jobsToBeReconciled, nil
}

func getFinishedStatus(j *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range j.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}
// IsJobFinished returns whether or not a job has completed successfully or failed.
func IsJobFinished(j *batchv1.Job) bool {
	isFinished, _ := getFinishedStatus(j)
	return isFinished
}

// When a job is created, enqueue the controller that manages it and update it's expectations.
func (jm *DelayJobController) addJob(obj interface{}) {
	job := obj.(*batchv1.Job)
	if job.DeletionTimestamp != nil {
		// on a restart of the controller, it's possible a new job shows up in a state that
		// is already pending deletion. Prevent the job from being a creation observation.
		jm.deleteJob(job)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(job); controllerRef != nil {
		delayJob := jm.resolveControllerRef(job.Namespace, controllerRef)
		if delayJob == nil {
			return
		}
		jm.enqueueController(delayJob)
		return
	}
}

func (jm *DelayJobController) deleteJob(obj interface{}) {
	job, ok := obj.(*batchv1.Job)

	// When a delete is dropped, the relist will notice a job in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		job, ok = tombstone.Obj.(*batchv1.Job)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a ReplicaSet %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(job)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	delayJob := jm.resolveControllerRef(job.Namespace, controllerRef)
	if delayJob == nil {
		return
	}
	jm.enqueueController(delayJob)
}

// updateJob figures out what DelayJob(s) manage a Job when the Job
// is updated and wake them up. If the anything of the Job have changed, we need to
// awaken both the old and new DelayJob. old and cur must be *batchv1.Job
// types.
func (jm *DelayJobController) updateJob(old, cur interface{}) {
	curJob := cur.(*batchv1.Job)
	oldJob := old.(*batchv1.Job)
	if curJob.ResourceVersion == oldJob.ResourceVersion {
		// Periodic resync will send update events for all known jobs.
		// Two different versions of the same jobs will always have different RVs.
		return
	}

	curControllerRef := metav1.GetControllerOf(curJob)
	oldControllerRef := metav1.GetControllerOf(oldJob)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if delayJob := jm.resolveControllerRef(oldJob.Namespace, oldControllerRef); delayJob != nil {
			jm.enqueueController(delayJob)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		delayJob := jm.resolveControllerRef(curJob.Namespace, curControllerRef)
		if delayJob == nil {
			return
		}
		jm.enqueueController(delayJob)
		return
	}
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (jm *DelayJobController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *delayjobv1.DelayJob {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	delayJob, err := jm.delayJobLister.DelayJobs(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if delayJob.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return delayJob
}

func (jm *DelayJobController) enqueueController(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	jm.queue.Add(key)
}

// updateDelayJob re-queues the DelayJob for next scheduled time if there is a
// change in spec.schedule otherwise it re-queues it now
func (jm *DelayJobController) updateDelayJob(old interface{}, curr interface{}) {
	oldCJ, okOld := old.(*delayjobv1.DelayJob)
	newCJ, okNew := curr.(*delayjobv1.DelayJob)
	if !okOld || !okNew {
		// typecasting of one failed, handle this better, may be log entry
		return
	}
	if oldCJ.Spec.ExecutionTime != newCJ.Spec.ExecutionTime {
		// TODO update
		//jm.enqueueControllerAfter(curr, *)
	}
	jm.enqueueController(curr)
}

func (jm *DelayJobController) enqueueControllerAfter(obj interface{}, t time.Duration) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	jm.queue.AddAfter(key, t)
}

func getJobFromTemplate(dj *delayjobv1.DelayJob) (*batchv1.Job, error) {
	labels := copyLabels(&dj.Spec.JobTemplate)
	annotations := copyAnnotations(&dj.Spec.JobTemplate)
	// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
	name := getJobName(dj)
	createTime := time.Now()
	if dj.Spec.ExecutionTime > createTime.Unix() {
		createTime = time.Unix(dj.Spec.ExecutionTime, 0)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:            labels,
			Annotations:       annotations,
			Name:              name,
			CreationTimestamp: metav1.Time{Time: createTime},
			OwnerReferences:   []metav1.OwnerReference{*metav1.NewControllerRef(dj, controllerKind)},
		},
	}
	dj.Spec.JobTemplate.Spec.DeepCopyInto(&job.Spec)
	return job, nil
}

func copyLabels(template *batchv1.JobTemplateSpec) labels.Set {
	l := make(labels.Set)
	for k, v := range template.Labels {
		l[k] = v
	}
	return l
}

func copyAnnotations(template *batchv1.JobTemplateSpec) labels.Set {
	a := make(labels.Set)
	for k, v := range template.Annotations {
		a[k] = v
	}
	return a
}

func getJobName(cj *delayjobv1.DelayJob) string {
	return fmt.Sprintf("%s-%d", cj.Name, cj.Spec.ExecutionTime)
}