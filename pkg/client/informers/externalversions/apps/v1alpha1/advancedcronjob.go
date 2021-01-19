/*
Copyright 2021 The Kruise Authors.

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

package v1alpha1

import (
	time "time"

	appsv1alpha1 "github.com/openkruise/kruise/apis/apps/v1alpha1"
	versioned "github.com/openkruise/kruise/pkg/client/clientset/versioned"
	internalinterfaces "github.com/openkruise/kruise/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/openkruise/kruise/pkg/client/listers/apps/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AdvancedCronJobInformer provides access to a shared informer and lister for
// AdvancedCronJobs.
type AdvancedCronJobInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.AdvancedCronJobLister
}

type advancedCronJobInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewAdvancedCronJobInformer constructs a new informer for AdvancedCronJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAdvancedCronJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAdvancedCronJobInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredAdvancedCronJobInformer constructs a new informer for AdvancedCronJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAdvancedCronJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1alpha1().AdvancedCronJobs(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1alpha1().AdvancedCronJobs(namespace).Watch(options)
			},
		},
		&appsv1alpha1.AdvancedCronJob{},
		resyncPeriod,
		indexers,
	)
}

func (f *advancedCronJobInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAdvancedCronJobInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *advancedCronJobInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&appsv1alpha1.AdvancedCronJob{}, f.defaultInformer)
}

func (f *advancedCronJobInformer) Lister() v1alpha1.AdvancedCronJobLister {
	return v1alpha1.NewAdvancedCronJobLister(f.Informer().GetIndexer())
}
