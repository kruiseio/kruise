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

package framework

import (
	appsv1alpha1 "github.com/openkruise/kruise/apis/apps/v1alpha1"
	kruiseclientset "github.com/openkruise/kruise/pkg/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type BroadcastJobTester struct {
	c  clientset.Interface
	kc kruiseclientset.Interface
	ns string
}

func NewBroadcastJobTester(c clientset.Interface, kc kruiseclientset.Interface, ns string) *BroadcastJobTester {
	return &BroadcastJobTester{
		c:  c,
		kc: kc,
		ns: ns,
	}
}

func (t *BroadcastJobTester) CreateBroadcastJob(job *appsv1alpha1.BroadcastJob) (*appsv1alpha1.BroadcastJob, error) {
	return t.kc.AppsV1alpha1().BroadcastJobs(t.ns).Create(job)
}

func (t *BroadcastJobTester) GetBroadcastJob(name string) (*appsv1alpha1.BroadcastJob, error) {
	return t.kc.AppsV1alpha1().BroadcastJobs(t.ns).Get(name, metav1.GetOptions{})
}

func (t *BroadcastJobTester) GetPodsOfJob(job *appsv1alpha1.BroadcastJob) (pods []*v1.Pod, err error) {
	podList, err := t.c.CoreV1().Pods(t.ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for i := range podList.Items {
		pod := &podList.Items[i]
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef != nil && controllerRef.UID == job.UID {
			pods = append(pods, pod)
		}
	}
	return pods, nil
}
