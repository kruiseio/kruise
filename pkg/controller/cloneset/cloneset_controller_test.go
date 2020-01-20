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

package cloneset

import (
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	appsv1alpha1 "github.com/openkruise/kruise/pkg/apis/apps/v1alpha1"
	clonesetutils "github.com/openkruise/kruise/pkg/controller/cloneset/utils"
	"github.com/openkruise/kruise/pkg/util/fieldindex"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}

var images = []string{"nginx:1.9.1", "nginx:1.9.2", "nginx:1.9.3"}

//const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &appsv1alpha1.CloneSet{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
		Spec: appsv1alpha1.CloneSetSpec{
			Replicas: getInt32(1),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "bar"}},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "nginx", Image: images[0]}},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "foo-vol"},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources:   v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("1Mi")}},
					},
				},
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	//recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, newReconciler(mgr))).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create an orphan pod
	orphanPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo-abcde",
			Labels:    map[string]string{"foo": "bar"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{Name: "nginx", Image: "nginx:apline"}},
		},
	}
	err = c.Create(context.TODO(), &orphanPod)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Create the CloneSet object and expect the Reconcile
	appsv1alpha1.SetDefaults_CloneSet(instance)
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)

	// Check 1 pod and 1 pvc have been created
	pods, pvcs := checkInstances(g, instance, 1, 1)
	g.Expect(pods[0].Labels[appsv1alpha1.CloneSetInstanceID]).Should(gomega.Equal(pvcs[0].Labels[appsv1alpha1.CloneSetInstanceID]))

	// Test for pods scale
	testScale(g, instance)

	// Get latest cloneset
	err = c.Get(context.TODO(), expectedRequest.NamespacedName, instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Test for pods update
	testUpdate(g, instance)
}

func testScale(g *gomega.GomegaWithT, instance *appsv1alpha1.CloneSet) {
	// Create orphan and owned pvcs
	pvc1 := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-vol-foo-hlfn7",
			Namespace: "default",
			Labels: map[string]string{
				"foo":                           "bar",
				appsv1alpha1.CloneSetInstanceID: "hlfn7",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources:   v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("1Mi")}},
		},
	}
	err := c.Create(context.TODO(), &pvc1)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	pvc2 := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-vol-foo-xub0a",
			Namespace: "default",
			Labels: map[string]string{
				"foo":                           "bar",
				appsv1alpha1.CloneSetInstanceID: "xub0a",
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         clonesetutils.ControllerKind.GroupVersion().String(),
				Kind:               clonesetutils.ControllerKind.Kind,
				Name:               instance.Name,
				UID:                instance.UID,
				Controller:         func() *bool { v := true; return &v }(),
				BlockOwnerDeletion: func() *bool { v := true; return &v }(),
			}},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources:   v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("1Mi")}},
		},
	}
	err = c.Create(context.TODO(), &pvc2)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Check 1 pod and 2 pvc have been created
	pods, pvcs := checkInstances(g, instance, 1, 2)
	g.Expect([]string{
		pods[0].Labels[appsv1alpha1.CloneSetInstanceID],
		pvc2.Labels[appsv1alpha1.CloneSetInstanceID],
	}).Should(gomega.ConsistOf(
		pvcs[0].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[1].Labels[appsv1alpha1.CloneSetInstanceID],
	))

	// Add replicas to 5, should reuse
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cs := appsv1alpha1.CloneSet{}
		if err := c.Get(context.TODO(), expectedRequest.NamespacedName, &cs); err != nil {
			return err
		}
		cs.Spec.Replicas = getInt32(5)
		return c.Update(context.TODO(), &cs)
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Check 5 pod and 5 pvc have been created
	pods, pvcs = checkInstances(g, instance, 5, 5)
	g.Expect([]string{
		pods[0].Labels[appsv1alpha1.CloneSetInstanceID],
		pods[1].Labels[appsv1alpha1.CloneSetInstanceID],
		pods[2].Labels[appsv1alpha1.CloneSetInstanceID],
		pods[3].Labels[appsv1alpha1.CloneSetInstanceID],
		pods[4].Labels[appsv1alpha1.CloneSetInstanceID],
	}).Should(gomega.ConsistOf(
		pvcs[0].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[1].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[2].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[3].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[4].Labels[appsv1alpha1.CloneSetInstanceID],
	))
	g.Expect(pvc1.Labels[appsv1alpha1.CloneSetInstanceID]).Should(gomega.And(
		gomega.Not(gomega.Equal(pods[0].Labels[appsv1alpha1.CloneSetInstanceID])),
		gomega.Not(gomega.Equal(pods[1].Labels[appsv1alpha1.CloneSetInstanceID])),
		gomega.Not(gomega.Equal(pods[2].Labels[appsv1alpha1.CloneSetInstanceID])),
		gomega.Not(gomega.Equal(pods[3].Labels[appsv1alpha1.CloneSetInstanceID])),
		gomega.Not(gomega.Equal(pods[4].Labels[appsv1alpha1.CloneSetInstanceID])),
	))

	// Delete instance 'xub0a'
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cs := appsv1alpha1.CloneSet{}
		if err := c.Get(context.TODO(), expectedRequest.NamespacedName, &cs); err != nil {
			return err
		}
		cs.Spec.Replicas = getInt32(4)
		cs.Spec.ScaleStrategy.PodsToDelete = append(cs.Spec.ScaleStrategy.PodsToDelete, "foo-xub0a")
		return c.Update(context.TODO(), &cs)
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	pods, pvcs = checkInstances(g, instance, 4, 4)
	g.Expect([]string{
		pods[0].Labels[appsv1alpha1.CloneSetInstanceID],
		pods[1].Labels[appsv1alpha1.CloneSetInstanceID],
		pods[2].Labels[appsv1alpha1.CloneSetInstanceID],
		pods[3].Labels[appsv1alpha1.CloneSetInstanceID],
	}).Should(gomega.ConsistOf(
		pvcs[0].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[1].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[2].Labels[appsv1alpha1.CloneSetInstanceID],
		pvcs[3].Labels[appsv1alpha1.CloneSetInstanceID],
	))
}

func testUpdate(g *gomega.GomegaWithT, instance *appsv1alpha1.CloneSet) {
	// No way to test maxUnavailable, for this is a k8s cluster with only etcd and kube-apiserver
	maxUnavailable := intstr.FromString("100%")
	pods0, pvcs0 := checkInstances(g, instance, 4, 4)

	// default to recreate update
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cs := appsv1alpha1.CloneSet{}
		if err := c.Get(context.TODO(), expectedRequest.NamespacedName, &cs); err != nil {
			return err
		}
		cs.Spec.Template.Spec.Containers[0].Image = images[1]
		cs.Spec.UpdateStrategy = appsv1alpha1.CloneSetUpdateStrategy{
			Type:           appsv1alpha1.RecreateCloneSetUpdateStrategyType,
			Partition:      getInt32(1),
			MaxUnavailable: &maxUnavailable,
		}
		return c.Update(context.TODO(), &cs)
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	checkStatus(g, 4, 3)
	pods1, pvcs1 := checkInstances(g, instance, 4, 4)
	samePodNames := getPodNames(pods0).Intersection(getPodNames(pods1))
	samePVCNames := getPVCNames(pvcs0).Intersection(getPVCNames(pvcs1))
	g.Expect(samePodNames.Len()).Should(gomega.Equal(1))
	g.Expect(samePVCNames.Len()).Should(gomega.Equal(1))
	g.Expect(strings.HasSuffix(samePVCNames.List()[0], samePodNames.List()[0])).Should(gomega.BeTrue())

	// inplace update
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cs := appsv1alpha1.CloneSet{}
		if err := c.Get(context.TODO(), expectedRequest.NamespacedName, &cs); err != nil {
			return err
		}
		cs.Spec.Template.Spec.Containers[0].Image = images[2]
		cs.Spec.UpdateStrategy = appsv1alpha1.CloneSetUpdateStrategy{
			Type:           appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
			Partition:      getInt32(2),
			MaxUnavailable: &maxUnavailable,
		}
		return c.Update(context.TODO(), &cs)
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	checkStatus(g, 4, 2)
	pods2, pvcs2 := checkInstances(g, instance, 4, 4)
	samePodNames = getPodNames(pods1).Intersection(getPodNames(pods2))
	samePVCNames = getPVCNames(pvcs1).Intersection(getPVCNames(pvcs2))
	g.Expect(samePodNames.Len()).Should(gomega.Equal(4))
	g.Expect(samePVCNames.Len()).Should(gomega.Equal(4))
}

func getInt32(i int32) *int32 {
	return &i
}

func checkInstances(g *gomega.GomegaWithT, cs *appsv1alpha1.CloneSet, podNum int, pvcNum int) ([]*v1.Pod, []*v1.PersistentVolumeClaim) {
	var pods []*v1.Pod
	g.Eventually(func() int {
		var err error
		pods, err = clonesetutils.GetActivePods(c, client.InNamespace("default").MatchingField(fieldindex.IndexNameForOwnerRefUID, string(cs.UID)))
		g.Expect(err).NotTo(gomega.HaveOccurred())
		return len(pods)
	}, time.Second*3, time.Millisecond*500).Should(gomega.Equal(podNum))

	var pvcs []*v1.PersistentVolumeClaim
	g.Eventually(func() int {
		pvcList := v1.PersistentVolumeClaimList{}
		err := c.List(context.TODO(), client.InNamespace(cs.Namespace).MatchingField(fieldindex.IndexNameForOwnerRefUID, string(cs.UID)), &pvcList)
		g.Expect(err).NotTo(gomega.HaveOccurred())
		pvcs = []*v1.PersistentVolumeClaim{}
		for i, pvc := range pvcList.Items {
			if pvc.DeletionTimestamp == nil {
				pvcs = append(pvcs, &pvcList.Items[i])
			}
		}
		return len(pvcs)
	}, time.Second*3, time.Millisecond*500).Should(gomega.Equal(pvcNum))

	return pods, pvcs
}

func checkStatus(g *gomega.GomegaWithT, total, updated int32) {
	g.Eventually(func() []int32 {
		cs := appsv1alpha1.CloneSet{}
		err := c.Get(context.TODO(), expectedRequest.NamespacedName, &cs)
		g.Expect(err).NotTo(gomega.HaveOccurred())
		return []int32{cs.Status.Replicas, cs.Status.UpdatedReplicas}
	}, time.Second*3, time.Millisecond*500).Should(gomega.Equal([]int32{total, updated}))
}

func getPodNames(pods []*v1.Pod) sets.String {
	s := sets.NewString()
	for _, p := range pods {
		s.Insert(p.Name)
	}
	return s
}

func getPVCNames(pvcs []*v1.PersistentVolumeClaim) sets.String {
	s := sets.NewString()
	for _, p := range pvcs {
		s.Insert(p.Name)
	}
	return s
}
