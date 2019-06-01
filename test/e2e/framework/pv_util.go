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

package framework

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// create the PV resource. Fails test on error.
func createPV(c clientset.Interface, pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	pv, err := c.CoreV1().PersistentVolumes().Create(pv)
	if err != nil {
		return nil, fmt.Errorf("PV Create API error: %v", err)
	}
	return pv, nil
}

// create the PV resource. Fails test on error.
func CreatePV(c clientset.Interface, pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	return createPV(c, pv)
}
