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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	delayjobv1 "k8s.io/delay-job-controller/controller/delayjob/v1"
	versioned "k8s.io/delay-job-controller/pkg/generated/clientset/versioned"
	internalinterfaces "k8s.io/delay-job-controller/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "k8s.io/delay-job-controller/pkg/generated/listers/delayjob/v1"
)

// DelayJobInformer provides access to a shared informer and lister for
// DelayJobs.
type DelayJobInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.DelayJobLister
}

type delayJobInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewDelayJobInformer constructs a new informer for DelayJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewDelayJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredDelayJobInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredDelayJobInformer constructs a new informer for DelayJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredDelayJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DelayjobV1().DelayJobs(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DelayjobV1().DelayJobs(namespace).Watch(context.TODO(), options)
			},
		},
		&delayjobv1.DelayJob{},
		resyncPeriod,
		indexers,
	)
}

func (f *delayJobInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredDelayJobInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *delayJobInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&delayjobv1.DelayJob{}, f.defaultInformer)
}

func (f *delayJobInformer) Lister() v1.DelayJobLister {
	return v1.NewDelayJobLister(f.Informer().GetIndexer())
}
