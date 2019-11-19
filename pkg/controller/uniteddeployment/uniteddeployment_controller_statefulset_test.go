/*
Copyright 2019 The Kruise Authors.

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

package uniteddeployment

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/openkruise/kruise/pkg/apis/apps/v1alpha1"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
var deploy = client.ObjectKey{Namespace: "default", Name: "foo"}

const timeout = time.Second * 2

var (
	one int32 = 1
	two int32 = 2
	ten int32 = 10
)

func TestStsReconcile(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"node-a"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)
	expectedStsCount(g, 1)
}

func TestStsSubsetProvision(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"node-a"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 1)
	sts := &stsList.Items[0]
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).ShouldNot(gomega.BeNil())
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)).Should(gomega.BeEquivalentTo(1))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key).Should(gomega.BeEquivalentTo("node-name"))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator).Should(gomega.BeEquivalentTo(corev1.NodeSelectorOpIn))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0]).Should(gomega.BeEquivalentTo("node-a"))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.Topology.Subsets = append(instance.Spec.Topology.Subsets, appsv1alpha1.Subset{
		Name: "subset-b",
		NodeSelector: corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "node-name",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"node-b"},
						},
					},
				},
			},
		},
	})
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	sts = getSubsetByName(stsList, "subset-a")
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).ShouldNot(gomega.BeNil())
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)).Should(gomega.BeEquivalentTo(1))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key).Should(gomega.BeEquivalentTo("node-name"))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator).Should(gomega.BeEquivalentTo(corev1.NodeSelectorOpIn))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0]).Should(gomega.BeEquivalentTo("node-a"))

	sts = getSubsetByName(stsList, "subset-b")
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).ShouldNot(gomega.BeNil())
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)).Should(gomega.BeEquivalentTo(1))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key).Should(gomega.BeEquivalentTo("node-name"))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator).Should(gomega.BeEquivalentTo(corev1.NodeSelectorOpIn))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0]).Should(gomega.BeEquivalentTo("node-b"))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.Topology.Subsets = instance.Spec.Topology.Subsets[1:]
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 1)
	sts = &stsList.Items[0]
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).ShouldNot(gomega.BeNil())
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)).Should(gomega.BeEquivalentTo(1))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key).Should(gomega.BeEquivalentTo("node-name"))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator).Should(gomega.BeEquivalentTo(corev1.NodeSelectorOpIn))
	g.Expect(len(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values)).Should(gomega.BeEquivalentTo(1))
	g.Expect(sts.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0]).Should(gomega.BeEquivalentTo("node-b"))
}

func TestStsDupSubset(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"node-a"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 1)

	subsetA := stsList.Items[0]
	dupSts := subsetA.DeepCopy()
	dupSts.Name = "dup-subset-a"
	dupSts.ResourceVersion = ""
	g.Expect(c.Create(context.TODO(), dupSts)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 3)
	expectedStsCount(g, 1)
}

func TestStsScale(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeA"},
										},
									},
								},
							},
						},
					},
					{
						Name: "subset-b",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeB"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas + *stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(1))

	var two int32 = 2
	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.Replicas = &two
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas).Should(gomega.BeEquivalentTo(1))
	g.Expect(*stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(1))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Status.SubsetReplicas).Should(gomega.BeEquivalentTo(map[string]int32{
		"subset-a": 1,
		"subset-b": 1,
	}))

	var five int32 = 6
	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.Replicas = &five
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas).Should(gomega.BeEquivalentTo(3))
	g.Expect(*stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(3))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Status.SubsetReplicas).Should(gomega.BeEquivalentTo(map[string]int32{
		"subset-a": 3,
		"subset-b": 3,
	}))

	var four int32 = 4
	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.Replicas = &four
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas).Should(gomega.BeEquivalentTo(2))
	g.Expect(*stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(2))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Status.SubsetReplicas).Should(gomega.BeEquivalentTo(map[string]int32{
		"subset-a": 2,
		"subset-b": 2,
	}))
}

func TestStsUpdate(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &two,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "containerA",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeA"},
										},
									},
								},
							},
						},
					},
					{
						Name: "subset-b",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeB"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas + *stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(2))
	revisionList := &appsv1.ControllerRevisionList{}
	g.Expect(c.List(context.TODO(), &client.ListOptions{}, revisionList))
	g.Expect(len(revisionList.Items)).Should(gomega.BeEquivalentTo(1))
	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	v1 := revisionList.Items[0].Name
	g.Expect(instance.Status.CurrentRevision).Should(gomega.BeEquivalentTo(v1))

	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:2.0"
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(stsList.Items[0].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	g.Expect(stsList.Items[1].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	revisionList = &appsv1.ControllerRevisionList{}
	g.Expect(c.List(context.TODO(), &client.ListOptions{}, revisionList))
	g.Expect(len(revisionList.Items)).Should(gomega.BeEquivalentTo(2))
	v2 := revisionList.Items[0].Name
	if v2 == v1 {
		v2 = revisionList.Items[1].Name
	}
	g.Expect(instance.Status.UpdateStatus.UpdatedRevision).Should(gomega.BeEquivalentTo(v2))
}

func TestStsRollingUpdatePartition(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &ten,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			UpdateStrategy: appsv1alpha1.UnitedDeploymentUpdateStrategy{
				Type: appsv1alpha1.ManualUpdateStrategyType,
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeA"},
										},
									},
								},
							},
						},
					},
					{
						Name: "subset-b",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeB"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas).Should(gomega.BeEquivalentTo(5))
	g.Expect(*stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(5))

	// update with partition
	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.UpdateStrategy.ManualUpdate = &appsv1alpha1.ManualUpdate{
		Partitions: map[string]int32{
			"subset-a": 4,
			"subset-b": 3,
		},
	}
	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:2.0"
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(stsList.Items[0].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	g.Expect(stsList.Items[1].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))

	stsA := getSubsetByName(stsList, "subset-a")
	g.Expect(stsA).ShouldNot(gomega.BeNil())
	g.Expect(*stsA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(4))

	stsB := getSubsetByName(stsList, "subset-b")
	g.Expect(stsB).ShouldNot(gomega.BeNil())
	g.Expect(*stsB.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(3))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Status.UpdateStatus.CurrentPartitions).Should(gomega.BeEquivalentTo(map[string]int32{
		"subset-a": 4,
		"subset-b": 3,
	}))

	// move on
	instance.Spec.UpdateStrategy.ManualUpdate = &appsv1alpha1.ManualUpdate{
		Partitions: map[string]int32{
			"subset-a": 0,
			"subset-b": 3,
		},
	}
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 4)

	stsList = expectedStsCount(g, 2)
	g.Expect(stsList.Items[0].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	g.Expect(stsList.Items[1].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))

	stsA = getSubsetByName(stsList, "subset-a")
	g.Expect(stsA).ShouldNot(gomega.BeNil())
	g.Expect(*stsA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(0))

	stsB = getSubsetByName(stsList, "subset-b")
	g.Expect(stsB).ShouldNot(gomega.BeNil())
	g.Expect(*stsB.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(3))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Status.UpdateStatus.CurrentPartitions).Should(gomega.BeEquivalentTo(map[string]int32{
		"subset-a": 0,
		"subset-b": 3,
	}))

	// move on
	instance.Spec.UpdateStrategy.ManualUpdate = &appsv1alpha1.ManualUpdate{
		Partitions: map[string]int32{},
	}
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(stsList.Items[0].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	g.Expect(stsList.Items[1].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))

	stsA = getSubsetByName(stsList, "subset-a")
	g.Expect(stsA).ShouldNot(gomega.BeNil())
	g.Expect(*stsA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(0))

	stsB = getSubsetByName(stsList, "subset-b")
	g.Expect(stsB).ShouldNot(gomega.BeNil())
	g.Expect(*stsB.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(0))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Status.UpdateStatus.CurrentPartitions).Should(gomega.BeEquivalentTo(map[string]int32{
		"subset-a": 0,
		"subset-b": 0,
	}))
}

func TestStsRollingUpdateDeleteStuckPod(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &ten,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			UpdateStrategy: appsv1alpha1.UnitedDeploymentUpdateStrategy{
				Type: appsv1alpha1.ManualUpdateStrategyType,
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeA"},
										},
									},
								},
							},
						},
					},
					{
						Name: "subset-b",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeB"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas).Should(gomega.BeEquivalentTo(5))
	g.Expect(*stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(5))

	g.Expect(provisionStatefulSetMockPod(c, &stsList.Items[0])).Should(gomega.BeNil())
	g.Expect(provisionStatefulSetMockPod(c, &stsList.Items[1])).Should(gomega.BeNil())

	g.Expect(collectPodOrdinal(c, &stsList.Items[0], "0,1,2,3,4")).Should(gomega.BeNil())
	g.Expect(collectPodOrdinal(c, &stsList.Items[1], "0,1,2,3,4")).Should(gomega.BeNil())

	// update with partition
	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.UpdateStrategy.ManualUpdate = &appsv1alpha1.ManualUpdate{
		Partitions: map[string]int32{
			"subset-a": 4,
			"subset-b": 3,
		},
	}
	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:2.0"
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(stsList.Items[0].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	g.Expect(stsList.Items[1].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))

	stsA := getSubsetByName(stsList, "subset-a")
	g.Expect(stsA).ShouldNot(gomega.BeNil())
	g.Expect(*stsA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(4))
	g.Expect(collectPodOrdinal(c, stsA, "0,1,2,3")).Should(gomega.BeNil())

	stsB := getSubsetByName(stsList, "subset-b")
	g.Expect(stsB).ShouldNot(gomega.BeNil())
	g.Expect(*stsB.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(3))
	g.Expect(collectPodOrdinal(c, stsB, "0,1,2")).Should(gomega.BeNil())
}

func TestStsOnDelete(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &ten,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container-a",
										Image: "nginx:1.0",
									},
								},
							},
						},
						UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
							Type: appsv1.OnDeleteStatefulSetStrategyType,
						},
					},
				},
			},
			UpdateStrategy: appsv1alpha1.UnitedDeploymentUpdateStrategy{
				Type: appsv1alpha1.ManualUpdateStrategyType,
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeA"},
										},
									},
								},
							},
						},
					},
					{
						Name: "subset-b",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeB"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas).Should(gomega.BeEquivalentTo(5))
	g.Expect(*stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(5))

	g.Expect(stsList.Items[0].Spec.UpdateStrategy.Type).Should(gomega.BeEquivalentTo(appsv1.OnDeleteStatefulSetStrategyType))
	g.Expect(stsList.Items[1].Spec.UpdateStrategy.Type).Should(gomega.BeEquivalentTo(appsv1.OnDeleteStatefulSetStrategyType))

	// update with partition
	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	instance.Spec.UpdateStrategy.ManualUpdate = &appsv1alpha1.ManualUpdate{
		Partitions: map[string]int32{
			"subset-a": 4,
			"subset-b": 3,
		},
	}
	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:2.0"
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(stsList.Items[0].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	g.Expect(stsList.Items[1].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))

	stsA := getSubsetByName(stsList, "subset-a")
	g.Expect(stsA).ShouldNot(gomega.BeNil())
	g.Expect(stsA.Spec.UpdateStrategy.Type).Should(gomega.BeEquivalentTo(appsv1.OnDeleteStatefulSetStrategyType))
	g.Expect(stsA.Spec.UpdateStrategy.RollingUpdate).Should(gomega.BeNil())

	stsB := getSubsetByName(stsList, "subset-b")
	g.Expect(stsB).ShouldNot(gomega.BeNil())
	g.Expect(stsB.Spec.UpdateStrategy.Type).Should(gomega.BeEquivalentTo(appsv1.OnDeleteStatefulSetStrategyType))
	g.Expect(stsB.Spec.UpdateStrategy.RollingUpdate).Should(gomega.BeNil())

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Status.UpdateStatus.CurrentPartitions).Should(gomega.BeEquivalentTo(map[string]int32{
		"subset-a": 4,
		"subset-b": 3,
	}))

	// move on
	instance.Spec.UpdateStrategy.ManualUpdate = &appsv1alpha1.ManualUpdate{
		Partitions: map[string]int32{
			"subset-a": 0,
			"subset-b": 3,
		},
	}
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	g.Expect(stsList.Items[0].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	g.Expect(stsList.Items[1].Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
}

func TestStsSubsetCount(t *testing.T) {
	g, requests, stopMgr, mgrStopped := setUp(t)
	defer func() {
		clean(g, c)
		close(stopMgr)
		mgrStopped.Wait()
	}()

	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Replicas: &ten,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "foo",
				},
			},
			Template: appsv1alpha1.SubsetTemplate{
				StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "containerA",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Subsets: []appsv1alpha1.Subset{
					{
						Name: "subset-a",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeA"},
										},
									},
								},
							},
						},
					},
					{
						Name: "subset-b",
						NodeSelector: corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeB"},
										},
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &ten,
		},
	}

	// Create the UnitedDeployment object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	waitReconcilerProcessFinished(g, requests, 3)

	stsList := expectedStsCount(g, 2)
	g.Expect(*stsList.Items[0].Spec.Replicas).Should(gomega.BeEquivalentTo(5))
	g.Expect(*stsList.Items[1].Spec.Replicas).Should(gomega.BeEquivalentTo(5))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	nine := intstr.FromInt(9)
	instance.Spec.Topology.Subsets[0].Replicas = &nine
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	setsubA := getSubsetByName(stsList, "subset-a")
	g.Expect(*setsubA.Spec.Replicas).Should(gomega.BeEquivalentTo(9))
	setsubB := getSubsetByName(stsList, "subset-b")
	g.Expect(*setsubB.Spec.Replicas).Should(gomega.BeEquivalentTo(1))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	percentage := intstr.FromString("40%")
	instance.Spec.Topology.Subsets[0].Replicas = &percentage
	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:2.0"
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	setsubA = getSubsetByName(stsList, "subset-a")
	g.Expect(*setsubA.Spec.Replicas).Should(gomega.BeEquivalentTo(4))
	g.Expect(setsubA.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))
	setsubB = getSubsetByName(stsList, "subset-b")
	g.Expect(*setsubB.Spec.Replicas).Should(gomega.BeEquivalentTo(6))
	g.Expect(setsubB.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:2.0"))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	percentage = intstr.FromString("30%")
	instance.Spec.Topology.Subsets[0].Replicas = &percentage
	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:3.0"
	instance.Spec.UpdateStrategy.Type = appsv1alpha1.ManualUpdateStrategyType
	instance.Spec.UpdateStrategy.ManualUpdate = &appsv1alpha1.ManualUpdate{
		Partitions: map[string]int32{
			"subset-a": 1,
		},
	}
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 2)
	setsubA = getSubsetByName(stsList, "subset-a")
	g.Expect(*setsubA.Spec.Replicas).Should(gomega.BeEquivalentTo(3))
	g.Expect(setsubA.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:3.0"))
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate).ShouldNot(gomega.BeNil())
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).ShouldNot(gomega.BeNil())
	g.Expect(*setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(1))
	setsubB = getSubsetByName(stsList, "subset-b")
	g.Expect(*setsubB.Spec.Replicas).Should(gomega.BeEquivalentTo(7))
	g.Expect(setsubB.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:3.0"))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	percentage = intstr.FromString("20%")
	instance.Spec.Topology.Subsets[0].Replicas = &percentage
	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:4.0"
	instance.Spec.UpdateStrategy.ManualUpdate.Partitions = map[string]int32{
		"subset-a": 2,
	}
	instance.Spec.Topology.Subsets = append(instance.Spec.Topology.Subsets, appsv1alpha1.Subset{
		Name: "subset-c",
		NodeSelector: corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "node-name",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"nodeC"},
						},
					},
				},
			},
		},
	})
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 2)

	stsList = expectedStsCount(g, 3)
	setsubA = getSubsetByName(stsList, "subset-a")
	g.Expect(*setsubA.Spec.Replicas).Should(gomega.BeEquivalentTo(2))
	g.Expect(setsubA.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:4.0"))
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate).ShouldNot(gomega.BeNil())
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).ShouldNot(gomega.BeNil())
	g.Expect(*setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(2))
	setsubB = getSubsetByName(stsList, "subset-b")
	g.Expect(*setsubB.Spec.Replicas).Should(gomega.BeEquivalentTo(4))
	g.Expect(setsubB.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:4.0"))
	setsubB = getSubsetByName(stsList, "subset-c")
	g.Expect(*setsubB.Spec.Replicas).Should(gomega.BeEquivalentTo(4))
	g.Expect(setsubB.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:4.0"))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	percentage = intstr.FromString("10%")
	instance.Spec.Topology.Subsets[0].Replicas = &percentage
	instance.Spec.Template.StatefulSetTemplate.Spec.Template.Spec.Containers[0].Image = "nginx:5.0"
	instance.Spec.UpdateStrategy.ManualUpdate.Partitions = map[string]int32{
		"subset-a": 2,
	}
	instance.Spec.Topology.Subsets = instance.Spec.Topology.Subsets[:2]
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 3)

	stsList = expectedStsCount(g, 2)
	setsubA = getSubsetByName(stsList, "subset-a")
	g.Expect(*setsubA.Spec.Replicas).Should(gomega.BeEquivalentTo(1))
	g.Expect(setsubA.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:5.0"))
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate).ShouldNot(gomega.BeNil())
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).ShouldNot(gomega.BeNil())
	g.Expect(*setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(1))
	setsubB = getSubsetByName(stsList, "subset-b")
	g.Expect(*setsubB.Spec.Replicas).Should(gomega.BeEquivalentTo(9))
	g.Expect(setsubB.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:5.0"))

	g.Expect(c.Get(context.TODO(), client.ObjectKey{Namespace: instance.Namespace, Name: instance.Name}, instance)).Should(gomega.BeNil())
	g.Expect(instance.Spec.UpdateStrategy.ManualUpdate.Partitions["subset-a"]).Should(gomega.BeEquivalentTo(2))
	percentage = intstr.FromString("40%")
	instance.Spec.Topology.Subsets[0].Replicas = &percentage
	g.Expect(c.Update(context.TODO(), instance)).Should(gomega.BeNil())
	waitReconcilerProcessFinished(g, requests, 3)

	stsList = expectedStsCount(g, 2)
	setsubA = getSubsetByName(stsList, "subset-a")
	g.Expect(*setsubA.Spec.Replicas).Should(gomega.BeEquivalentTo(4))
	g.Expect(setsubA.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:5.0"))
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate).ShouldNot(gomega.BeNil())
	g.Expect(setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).ShouldNot(gomega.BeNil())
	g.Expect(*setsubA.Spec.UpdateStrategy.RollingUpdate.Partition).Should(gomega.BeEquivalentTo(2))
	setsubB = getSubsetByName(stsList, "subset-b")
	g.Expect(*setsubB.Spec.Replicas).Should(gomega.BeEquivalentTo(6))
	g.Expect(setsubB.Spec.Template.Spec.Containers[0].Image).Should(gomega.BeEquivalentTo("nginx:5.0"))
}

func collectPodOrdinal(c client.Client, sts *appsv1.StatefulSet, expected string) error {
	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return err
	}

	podList := &corev1.PodList{}
	if err := c.List(context.TODO(), &client.ListOptions{LabelSelector: selector}, podList); err != nil {
		return err
	}

	marks := make([]bool, len(podList.Items))
	for _, pod := range podList.Items {
		ordinal := int(getOrdinal(&pod))
		if ordinal >= len(marks) || ordinal < 0 {
			continue
		}

		marks[ordinal] = true
	}

	got := ""
	for idx, mark := range marks {
		if mark {
			got = fmt.Sprintf("%s,%d", got, idx)
		}
	}

	if len(got) > 0 {
		got = got[1:]
	}

	if got != expected {
		return fmt.Errorf("expected %s, got %s", expected, got)
	}

	return nil
}

func provisionStatefulSetMockPod(c client.Client, sts *appsv1.StatefulSet) error {
	if sts.Spec.Replicas == nil {
		return nil
	}

	replicas := *sts.Spec.Replicas
	for {
		replicas--
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: sts.Namespace,
				Name:      fmt.Sprintf("%s-%d", sts.Name, replicas),
				Labels:    sts.Spec.Template.Labels,
			},
			Spec: sts.Spec.Template.Spec,
		}

		if err := c.Create(context.TODO(), pod); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return err
			}
		}

		if replicas == 0 {
			return nil
		}
	}
}

func waitReconcilerProcessFinished(g *gomega.GomegaWithT, requests chan reconcile.Request, minCount int) {
	timeoutChan := time.After(timeout)
	maxTimeoutChan := time.After(timeout * 2)
	for {
		minCount--
		select {
		case <-requests:
			continue
		case <-timeoutChan:
			if minCount <= 0 {
				return
			}
		case <-maxTimeoutChan:
			return
		}
	}
}

func getSubsetByName(stsList *appsv1.StatefulSetList, name string) *appsv1.StatefulSet {
	for _, sts := range stsList.Items {
		if sts.Labels[appsv1alpha1.SubSetNameLabelKey] == name {
			return &sts
		}
	}

	return nil
}

func expectedStsCount(g *gomega.GomegaWithT, count int) *appsv1.StatefulSetList {
	stsList := &appsv1.StatefulSetList{}
	g.Eventually(func() error {
		if err := c.List(context.TODO(), &client.ListOptions{}, stsList); err != nil {
			return err
		}

		if len(stsList.Items) != count {
			return fmt.Errorf("expected %d sts, got %d", count, len(stsList.Items))
		}

		return nil
	}, timeout).Should(gomega.Succeed())

	return stsList
}

func setUp(t *testing.T) (*gomega.GomegaWithT, chan reconcile.Request, chan struct{}, *sync.WaitGroup) {
	g := gomega.NewGomegaWithT(t)
	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()
	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	stopMgr, mgrStopped := StartTestManager(mgr, g)

	subsetReplicasFn = func(subset *Subset) int32 {
		return subset.Spec.Replicas
	}

	return g, requests, stopMgr, mgrStopped
}

func clean(g *gomega.GomegaWithT, c client.Client) {
	udList := &appsv1alpha1.UnitedDeploymentList{}
	if err := c.List(context.TODO(), &client.ListOptions{}, udList); err == nil {
		for _, ud := range udList.Items {
			c.Delete(context.TODO(), &ud)
		}
	}
	g.Eventually(func() error {
		if err := c.List(context.TODO(), &client.ListOptions{}, udList); err != nil {
			return err
		}

		if len(udList.Items) != 0 {
			return fmt.Errorf("expected %d sts, got %d", 0, len(udList.Items))
		}

		return nil
	}, timeout, time.Second).Should(gomega.Succeed())

	rList := &appsv1.ControllerRevisionList{}
	if err := c.List(context.TODO(), &client.ListOptions{}, rList); err == nil {
		for _, ud := range rList.Items {
			c.Delete(context.TODO(), &ud)
		}
	}
	g.Eventually(func() error {
		if err := c.List(context.TODO(), &client.ListOptions{}, rList); err != nil {
			return err
		}

		if len(rList.Items) != 0 {
			return fmt.Errorf("expected %d sts, got %d", 0, len(rList.Items))
		}

		return nil
	}, timeout, time.Second).Should(gomega.Succeed())

	stsList := &appsv1.StatefulSetList{}
	if err := c.List(context.TODO(), &client.ListOptions{}, stsList); err == nil {
		for _, sts := range stsList.Items {
			c.Delete(context.TODO(), &sts)
		}
	}
	g.Eventually(func() error {
		if err := c.List(context.TODO(), &client.ListOptions{}, stsList); err != nil {
			return err
		}

		if len(stsList.Items) != 0 {
			return fmt.Errorf("expected %d sts, got %d", 0, len(stsList.Items))
		}

		return nil
	}, timeout, time.Second).Should(gomega.Succeed())

	podList := &corev1.PodList{}
	if err := c.List(context.TODO(), &client.ListOptions{}, podList); err == nil {
		for _, pod := range podList.Items {
			c.Delete(context.TODO(), &pod)
		}
	}
	g.Eventually(func() error {
		if err := c.List(context.TODO(), &client.ListOptions{}, podList); err != nil {
			return err
		}

		if len(podList.Items) != 0 {
			return fmt.Errorf("expected %d sts, got %d", 0, len(podList.Items))
		}

		return nil
	}, timeout, time.Second).Should(gomega.Succeed())
}
