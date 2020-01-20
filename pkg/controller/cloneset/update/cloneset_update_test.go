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

package update

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/openkruise/kruise/pkg/apis"
	appsv1alpha1 "github.com/openkruise/kruise/pkg/apis/apps/v1alpha1"
	clonesetutils "github.com/openkruise/kruise/pkg/controller/cloneset/utils"
	"github.com/openkruise/kruise/pkg/util"
	"github.com/openkruise/kruise/pkg/util/expectations"
	"github.com/openkruise/kruise/pkg/util/inplaceupdate"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type manageCase struct {
	name           string
	cs             *appsv1alpha1.CloneSet
	updateRevision *apps.ControllerRevision
	revisions      []*apps.ControllerRevision
	pods           []*v1.Pod
	pvcs           []*v1.PersistentVolumeClaim
	expectedPods   []*v1.Pod
	expectedPVCs   []*v1.PersistentVolumeClaim
}

func (mc *manageCase) initial() []runtime.Object {
	var initialObjs []runtime.Object
	mc.cs.Name = "clone-test"
	initialObjs = append(initialObjs, mc.cs)

	for i := range mc.pods {
		initialObjs = append(initialObjs, mc.pods[i])
	}

	for i := range mc.pvcs {
		initialObjs = append(initialObjs, mc.pvcs[i])
	}

	return initialObjs
}

func getInt32Pointer(i int32) *int32 {
	return &i
}

func TestMange(t *testing.T) {
	apis.AddToScheme(scheme.Scheme)
	now := metav1.NewTime(time.Unix(time.Now().Add(-time.Hour).Unix(), 0))
	cases := []manageCase{
		{
			name:           "do nothing",
			cs:             &appsv1alpha1.CloneSet{Spec: appsv1alpha1.CloneSetSpec{Replicas: getInt32Pointer(1)}},
			updateRevision: &apps.ControllerRevision{ObjectMeta: metav1.ObjectMeta{Name: "rev-new"}},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0", Labels: map[string]string{apps.ControllerRevisionHashLabelKey: "rev-new"}},
					Spec:       v1.PodSpec{ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}}},
					Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{
						{Type: v1.PodReady, Status: v1.ConditionTrue},
						{Type: appsv1alpha1.InPlaceUpdateReady, Status: v1.ConditionTrue},
					}},
				},
			},
			expectedPods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0", Labels: map[string]string{apps.ControllerRevisionHashLabelKey: "rev-new"}},
					Spec:       v1.PodSpec{ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}}},
					Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{
						{Type: v1.PodReady, Status: v1.ConditionTrue},
						{Type: appsv1alpha1.InPlaceUpdateReady, Status: v1.ConditionTrue},
					}},
				},
			},
		},
		{
			name:           "normal update condition",
			cs:             &appsv1alpha1.CloneSet{Spec: appsv1alpha1.CloneSetSpec{Replicas: getInt32Pointer(1)}},
			updateRevision: &apps.ControllerRevision{ObjectMeta: metav1.ObjectMeta{Name: "rev-new"}},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0", Labels: map[string]string{apps.ControllerRevisionHashLabelKey: "rev-new"}},
					Spec:       v1.PodSpec{ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}}},
					Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{
						{Type: v1.PodReady, Status: v1.ConditionTrue},
					}},
				},
			},
			expectedPods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0", Labels: map[string]string{apps.ControllerRevisionHashLabelKey: "rev-new"}},
					Spec:       v1.PodSpec{ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}}},
					Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{
						{Type: v1.PodReady, Status: v1.ConditionTrue},
						{Type: appsv1alpha1.InPlaceUpdateReady, LastTransitionTime: now, Status: v1.ConditionTrue},
					}},
				},
			},
		},
		{
			name: "recreate update 1",
			cs: &appsv1alpha1.CloneSet{Spec: appsv1alpha1.CloneSetSpec{
				Replicas:       getInt32Pointer(1),
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{Type: appsv1alpha1.RecreateCloneSetUpdateStrategyType},
			}},
			updateRevision: &apps.ControllerRevision{ObjectMeta: metav1.ObjectMeta{Name: "rev-new"}},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0", Labels: map[string]string{
						apps.ControllerRevisionHashLabelKey: "rev-old",
						appsv1alpha1.CloneSetInstanceID:     "id-0",
					}},
					Spec: v1.PodSpec{ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}}},
					Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{
						{Type: v1.PodReady, Status: v1.ConditionTrue},
						{Type: appsv1alpha1.InPlaceUpdateReady, Status: v1.ConditionTrue},
					}},
				},
			},
			pvcs: []*v1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-0", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-1"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-2", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
			},
			expectedPods: []*v1.Pod{},
			expectedPVCs: []*v1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-1"}}},
			},
		},
		{
			name: "recreate update 2",
			cs: &appsv1alpha1.CloneSet{Spec: appsv1alpha1.CloneSetSpec{
				Replicas:       getInt32Pointer(1),
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{Type: appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType},
			}},
			updateRevision: &apps.ControllerRevision{
				ObjectMeta: metav1.ObjectMeta{Name: "rev-new"},
				Data:       runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"$patch":"replace","spec":{"containers":[{"name":"c1","image":"foo2","env":["name":"k", "value":"v"]}]}}}}`)},
			},
			revisions: []*apps.ControllerRevision{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "rev-old"},
					Data:       runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"$patch":"replace","spec":{"containers":[{"name":"c1","image":"foo1"}]}}}}`)},
				},
			},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0", Labels: map[string]string{
						apps.ControllerRevisionHashLabelKey: "rev-old",
						appsv1alpha1.CloneSetInstanceID:     "id-0",
					}},
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}},
						Containers:     []v1.Container{{Name: "c1", Image: "foo1"}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{
							{Type: v1.PodReady, Status: v1.ConditionTrue},
							{Type: appsv1alpha1.InPlaceUpdateReady, Status: v1.ConditionTrue},
						},
						ContainerStatuses: []v1.ContainerStatus{{Name: "c1", ImageID: "image-id-xyz"}},
					},
				},
			},
			pvcs: []*v1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-0", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-1"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-2", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
			},
			expectedPods: []*v1.Pod{},
			expectedPVCs: []*v1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-1"}}},
			},
		},
		{
			name: "inplace update",
			cs: &appsv1alpha1.CloneSet{Spec: appsv1alpha1.CloneSetSpec{
				Replicas:       getInt32Pointer(1),
				UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{Type: appsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType},
			}},
			updateRevision: &apps.ControllerRevision{
				ObjectMeta: metav1.ObjectMeta{Name: "rev-new"},
				Data:       runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"$patch":"replace","spec":{"containers":[{"name":"c1","image":"foo2"}]}}}}`)},
			},
			revisions: []*apps.ControllerRevision{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "rev-old"},
					Data:       runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"$patch":"replace","spec":{"containers":[{"name":"c1","image":"foo1"}]}}}}`)},
				},
			},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0", Labels: map[string]string{
						apps.ControllerRevisionHashLabelKey: "rev-old",
						appsv1alpha1.CloneSetInstanceID:     "id-0",
					}},
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}},
						Containers:     []v1.Container{{Name: "c1", Image: "foo1"}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{
							{Type: v1.PodReady, Status: v1.ConditionTrue},
							{Type: appsv1alpha1.InPlaceUpdateReady, Status: v1.ConditionTrue},
						},
						ContainerStatuses: []v1.ContainerStatus{{Name: "c1", ImageID: "image-id-xyz"}},
					},
				},
			},
			pvcs: []*v1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-0", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-1"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-2", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
			},
			expectedPods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod-0",
						Labels: map[string]string{apps.ControllerRevisionHashLabelKey: "rev-new", appsv1alpha1.CloneSetInstanceID: "id-0"},
						Annotations: map[string]string{appsv1alpha1.InPlaceUpdateStateKey: util.DumpJSON(appsv1alpha1.InPlaceUpdateState{
							Revision:              "rev-new",
							UpdateTimestamp:       metav1.Now(),
							LastContainerStatuses: map[string]appsv1alpha1.InPlaceUpdateContainerStatus{"c1": {ImageID: "image-id-xyz"}},
						})},
					},
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{{ConditionType: appsv1alpha1.InPlaceUpdateReady}},
						Containers:     []v1.Container{{Name: "c1", Image: "foo2"}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{
							{Type: v1.PodReady, Status: v1.ConditionTrue},
							{Type: appsv1alpha1.InPlaceUpdateReady, Status: v1.ConditionFalse, Reason: "StartInPlaceUpdate", LastTransitionTime: now},
						},
						ContainerStatuses: []v1.ContainerStatus{{Name: "c1", ImageID: "image-id-xyz"}},
					},
				},
			},
			expectedPVCs: []*v1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-0", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-1"}}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pvc-2", Labels: map[string]string{appsv1alpha1.CloneSetInstanceID: "id-0"}}},
			},
		},
	}

	for _, mc := range cases {
		initialObjs := mc.initial()
		fakeClient := fake.NewFakeClient(initialObjs...)
		ctrl := &realControl{
			fakeClient,
			inplaceupdate.NewForTest(fakeClient, apps.ControllerRevisionHashLabelKey, func() metav1.Time { return now }),
			record.NewFakeRecorder(10),
			expectations.NewScaleExpectations(),
			expectations.NewUpdateExpectations(clonesetutils.GetPodRevision),
		}
		if err := ctrl.Manage(mc.cs, mc.updateRevision, mc.revisions, mc.pods, mc.pvcs); err != nil {
			t.Fatalf("Failed to test %s, manage error: %v", mc.name, err)
		}
		podList := v1.PodList{}
		if err := ctrl.Client.List(context.TODO(), &client.ListOptions{}, &podList); err != nil {
			t.Fatalf("Failed to test %s, get pods error: %v", mc.name, err)
		}
		if len(podList.Items) != len(mc.expectedPods) {
			t.Fatalf("Failed to test %s, unexpected pods length, expected %v, got %v", mc.name, util.DumpJSON(mc.expectedPods), util.DumpJSON(podList.Items))
		}
		for _, p := range mc.expectedPods {
			gotPod := &v1.Pod{}
			if err := ctrl.Client.Get(context.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Name}, gotPod); err != nil {
				t.Fatalf("Failed to test %s, get pod %s error: %v", mc.name, p.Name, err)
			}
			if !reflect.DeepEqual(gotPod, p) {
				t.Fatalf("Failed to test %s, unexpected pod %s, expected \n%v\n got \n%v", mc.name, p.Name, util.DumpJSON(p), util.DumpJSON(gotPod))
			}
		}
	}
}

func TestSortUpdateIndexes(t *testing.T) {
	cases := []struct {
		strategy          appsv1alpha1.CloneSetUpdateStrategy
		pods              []*v1.Pod
		waitUpdateIndexes []int
		expectedIndexes   []int
	}{
		{
			strategy: appsv1alpha1.CloneSetUpdateStrategy{},
			pods: []*v1.Pod{
				{Status: v1.PodStatus{Phase: v1.PodPending, Conditions: []v1.PodCondition{{Type: v1.PodReady, Status: v1.ConditionTrue}}}},
				{Status: v1.PodStatus{Phase: v1.PodPending}},
				{Status: v1.PodStatus{Phase: v1.PodPending}},
				{Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{{Type: v1.PodReady, Status: v1.ConditionTrue}}}},
				{Status: v1.PodStatus{Phase: v1.PodRunning}},
			},
			waitUpdateIndexes: []int{0, 1, 3, 4},
			expectedIndexes:   []int{1, 0, 4, 3},
		},
	}

	for i, tc := range cases {
		got := sortUpdateIndexes(tc.strategy, tc.pods, tc.waitUpdateIndexes)
		if !reflect.DeepEqual(got, tc.expectedIndexes) {
			t.Fatalf("case #%d failed, expected %v, got %v", i, tc.expectedIndexes, got)
		}
	}
}

func TestCalculateUpdateCount(t *testing.T) {
	readyPod := func() *v1.Pod {
		return &v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{{Type: v1.PodReady, Status: v1.ConditionTrue}}}}
	}
	cases := []struct {
		strategy          appsv1alpha1.CloneSetUpdateStrategy
		totalReplicas     int
		waitUpdateIndexes []int
		pods              []*v1.Pod
		expectedResult    int
	}{
		{
			strategy:          appsv1alpha1.CloneSetUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			strategy:          appsv1alpha1.CloneSetUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), {}, readyPod()},
			expectedResult:    0,
		},
		{
			strategy:          appsv1alpha1.CloneSetUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{{}, readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			strategy:          appsv1alpha1.CloneSetUpdateStrategy{},
			totalReplicas:     10,
			waitUpdateIndexes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			pods:              []*v1.Pod{{}, readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), {}, readyPod()},
			expectedResult:    1,
		},
		{
			strategy:          appsv1alpha1.CloneSetUpdateStrategy{Partition: getInt32Pointer(2), MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromInt(3))},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1},
			pods:              []*v1.Pod{{}, readyPod(), readyPod()},
			expectedResult:    0,
		},
		{
			strategy:          appsv1alpha1.CloneSetUpdateStrategy{Partition: getInt32Pointer(2), MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromString("50%"))},
			totalReplicas:     8,
			waitUpdateIndexes: []int{0, 1, 2, 3, 4, 5, 6},
			pods:              []*v1.Pod{{}, readyPod(), {}, readyPod(), readyPod(), readyPod(), readyPod(), {}},
			expectedResult:    3,
		},
	}

	for i, tc := range cases {
		res := calculateUpdateCount(tc.strategy, tc.totalReplicas, tc.waitUpdateIndexes, tc.pods)
		if res != tc.expectedResult {
			t.Fatalf("case #%d failed, expected %d, got %d", i, tc.expectedResult, res)
		}
	}
}
