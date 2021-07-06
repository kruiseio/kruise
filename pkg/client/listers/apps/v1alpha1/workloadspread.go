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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/openkruise/kruise/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// WorkloadSpreadLister helps list WorkloadSpreads.
type WorkloadSpreadLister interface {
	// List lists all WorkloadSpreads in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.WorkloadSpread, err error)
	// WorkloadSpreads returns an object that can list and get WorkloadSpreads.
	WorkloadSpreads(namespace string) WorkloadSpreadNamespaceLister
	WorkloadSpreadListerExpansion
}

// workloadSpreadLister implements the WorkloadSpreadLister interface.
type workloadSpreadLister struct {
	indexer cache.Indexer
}

// NewWorkloadSpreadLister returns a new WorkloadSpreadLister.
func NewWorkloadSpreadLister(indexer cache.Indexer) WorkloadSpreadLister {
	return &workloadSpreadLister{indexer: indexer}
}

// List lists all WorkloadSpreads in the indexer.
func (s *workloadSpreadLister) List(selector labels.Selector) (ret []*v1alpha1.WorkloadSpread, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WorkloadSpread))
	})
	return ret, err
}

// WorkloadSpreads returns an object that can list and get WorkloadSpreads.
func (s *workloadSpreadLister) WorkloadSpreads(namespace string) WorkloadSpreadNamespaceLister {
	return workloadSpreadNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// WorkloadSpreadNamespaceLister helps list and get WorkloadSpreads.
type WorkloadSpreadNamespaceLister interface {
	// List lists all WorkloadSpreads in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.WorkloadSpread, err error)
	// Get retrieves the WorkloadSpread from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.WorkloadSpread, error)
	WorkloadSpreadNamespaceListerExpansion
}

// workloadSpreadNamespaceLister implements the WorkloadSpreadNamespaceLister
// interface.
type workloadSpreadNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all WorkloadSpreads in the indexer for a given namespace.
func (s workloadSpreadNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.WorkloadSpread, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WorkloadSpread))
	})
	return ret, err
}

// Get retrieves the WorkloadSpread from the indexer for a given namespace and name.
func (s workloadSpreadNamespaceLister) Get(name string) (*v1alpha1.WorkloadSpread, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("workloadspread"), name)
	}
	return obj.(*v1alpha1.WorkloadSpread), nil
}
