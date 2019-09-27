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

package mutating

import (
	"context"
	"encoding/json"
	"net/http"

	appsv1alpha1 "github.com/openkruise/kruise/pkg/apis/apps/v1alpha1"
	"github.com/openkruise/kruise/pkg/util"
	patchutil "github.com/openkruise/kruise/pkg/util/patch"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

func init() {
	webhookName := "mutating-create-update-statefulset"
	if HandlerMap[webhookName] == nil {
		HandlerMap[webhookName] = []admission.Handler{}
	}
	HandlerMap[webhookName] = append(HandlerMap[webhookName], &StatefulSetCreateUpdateHandler{})
}

// StatefulSetCreateUpdateHandler handles StatefulSet
type StatefulSetCreateUpdateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - import sigs.k8s.io/controller-runtime/pkg/client
	// - uncomment the InjectClient method at the bottom of this file.
	// Client  client.Client

	// Decoder decodes objects
	Decoder types.Decoder
}

func (h *StatefulSetCreateUpdateHandler) mutatingStatefulSetFn(ctx context.Context, obj *appsv1alpha1.StatefulSet) error {
	// TODO(user): implement your admission logic
	return nil
}

var _ admission.Handler = &StatefulSetCreateUpdateHandler{}

// Handle handles admission requests.
func (h *StatefulSetCreateUpdateHandler) Handle(ctx context.Context, req types.Request) types.Response {
	obj := &appsv1alpha1.StatefulSet{}

	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	appsv1alpha1.SetDefaults_StatefulSet(obj)
	obj.Status = appsv1alpha1.StatefulSetStatus{}

	err = h.mutatingStatefulSetFn(ctx, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}

	marshaledPod, err := json.Marshal(obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	resp := patchutil.ResponseFromRaw(req.AdmissionRequest.Object.Raw, marshaledPod)
	if len(resp.Patches) > 0 {
		klog.V(5).Infof("Admit StatefulSet %s/%s patches: %v", obj.Namespace, obj.Name, util.DumpJSON(resp.Patches))
	}
	return resp
}

//var _ inject.Client = &StatefulSetCreateUpdateHandler{}
//
//// InjectClient injects the client into the StatefulSetCreateUpdateHandler
//func (h *StatefulSetCreateUpdateHandler) InjectClient(c client.Client) error {
//	h.Client = c
//	return nil
//}

var _ inject.Decoder = &StatefulSetCreateUpdateHandler{}

// InjectDecoder injects the decoder into the StatefulSetCreateUpdateHandler
func (h *StatefulSetCreateUpdateHandler) InjectDecoder(d types.Decoder) error {
	h.Decoder = d
	return nil
}
