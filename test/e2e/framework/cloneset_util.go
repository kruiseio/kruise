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

type CloneSetTester struct {
	c  clientset.Interface
	kc kruiseclientset.Interface
	ns string
}

func NewCloneSetTester(c clientset.Interface, kc kruiseclientset.Interface, ns string) *CloneSetTester {
	return &CloneSetTester{
		c:  c,
		kc: kc,
		ns: ns,
	}
}

func (t *CloneSetTester) NewCloneSet(name string, replicas int32) *appsv1alpha1.CloneSet {
	return &appsv1alpha1.CloneSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.ns,
			Name:      name,
		},
		Spec: appsv1alpha1.CloneSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"owner": name}},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"owner": name},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.9.1",
							Env: []v1.EnvVar{
								{Name: "test", Value: "foo"},
							},
						},
					},
				},
			},
		},
	}
}

func (t *CloneSetTester) CreateCloneSet(cs *appsv1alpha1.CloneSet) (*appsv1alpha1.CloneSet, error) {
	return t.kc.AppsV1alpha1().CloneSets(t.ns).Create(cs)
}

func (t *CloneSetTester) GetCloneSet(name string) (*appsv1alpha1.CloneSet, error) {
	return t.kc.AppsV1alpha1().CloneSets(t.ns).Get(name, metav1.GetOptions{})
}
