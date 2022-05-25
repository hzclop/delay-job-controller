/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1 "k8s.io/delay-job-controller/controller/delayjob/v1"
	scheme "k8s.io/delay-job-controller/pkg/generated/clientset/versioned/scheme"
)

// DelayJobsGetter has a method to return a DelayJobInterface.
// A group's client should implement this interface.
type DelayJobsGetter interface {
	DelayJobs(namespace string) DelayJobInterface
}

// DelayJobInterface has methods to work with DelayJob resources.
type DelayJobInterface interface {
	Create(ctx context.Context, delayJob *v1.DelayJob, opts metav1.CreateOptions) (*v1.DelayJob, error)
	Update(ctx context.Context, delayJob *v1.DelayJob, opts metav1.UpdateOptions) (*v1.DelayJob, error)
	UpdateStatus(ctx context.Context, delayJob *v1.DelayJob, opts metav1.UpdateOptions) (*v1.DelayJob, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.DelayJob, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.DelayJobList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.DelayJob, err error)
	DelayJobExpansion
}

// delayJobs implements DelayJobInterface
type delayJobs struct {
	client rest.Interface
	ns     string
}

// newDelayJobs returns a DelayJobs
func newDelayJobs(c *DelayjobV1Client, namespace string) *delayJobs {
	return &delayJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the delayJob, and returns the corresponding delayJob object, and an error if there is any.
func (c *delayJobs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.DelayJob, err error) {
	result = &v1.DelayJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("delayjobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DelayJobs that match those selectors.
func (c *delayJobs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.DelayJobList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.DelayJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("delayjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested delayJobs.
func (c *delayJobs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("delayjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a delayJob and creates it.  Returns the server's representation of the delayJob, and an error, if there is any.
func (c *delayJobs) Create(ctx context.Context, delayJob *v1.DelayJob, opts metav1.CreateOptions) (result *v1.DelayJob, err error) {
	result = &v1.DelayJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("delayjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(delayJob).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a delayJob and updates it. Returns the server's representation of the delayJob, and an error, if there is any.
func (c *delayJobs) Update(ctx context.Context, delayJob *v1.DelayJob, opts metav1.UpdateOptions) (result *v1.DelayJob, err error) {
	result = &v1.DelayJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("delayjobs").
		Name(delayJob.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(delayJob).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *delayJobs) UpdateStatus(ctx context.Context, delayJob *v1.DelayJob, opts metav1.UpdateOptions) (result *v1.DelayJob, err error) {
	result = &v1.DelayJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("delayjobs").
		Name(delayJob.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(delayJob).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the delayJob and deletes it. Returns an error if one occurs.
func (c *delayJobs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("delayjobs").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *delayJobs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("delayjobs").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched delayJob.
func (c *delayJobs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.DelayJob, err error) {
	result = &v1.DelayJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("delayjobs").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
