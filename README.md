# OpenKruise/Kruise

[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/openkruise/kruise)](https://goreportcard.com/report/github.com/openkruise/kruise)
[![codecov](https://codecov.io/gh/openkruise/kruise/branch/master/graph/badge.svg)](https://codecov.io/gh/openkruise/kruise)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2908/badge)](https://bestpractices.coreinfrastructure.org/en/projects/2908)

Kruise is at the core of the OpenKruise project. It is a set of controllers which extends and complements 
[Kubernetes core controllers](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/)
on application workload management.

Today, Kruise offers three application workload controllers:

* [Advanced StatefulSet](./docs/astatefulset/README.md): An enhanced version of default [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) with extra functionalities such as `inplace-update`, sharding by namespace.

* [BroadcastJob](./docs/broadcastJob/README.md): A job that runs pods to completion across all the nodes in the cluster.

* [SidecarSet](./docs/sidecarSet/README.md): A controller that injects sidecar container into the pod spec based on selectors.

Please see [documents](./docs/README.md) for more technical information.

Several [tutorials](./docs/tutorial/README.md) are provided to demonstrate how to use the workload controllers.

## Getting started

### Install with YAML files

##### Install CRDs

```
kubectl apply -f https://raw.githubusercontent.com/kruiseio/kruise/master/config/crds/apps_v1alpha1_broadcastjob.yaml
kubectl apply -f https://raw.githubusercontent.com/kruiseio/kruise/master/config/crds/apps_v1alpha1_sidecarset.yaml
kubectl apply -f https://raw.githubusercontent.com/kruiseio/kruise/master/config/crds/apps_v1alpha1_statefulset.yaml
```
Note that ALL three CRDs need to be installed for kruise-controller to run properly.

##### Install kruise-controller-manager

`kubectl apply -f https://raw.githubusercontent.com/kruiseio/kruise/master/config/manager/all_in_one.yaml`

Or run from the repo root directory:

`kustomize build config/default | kubectl apply -f -`

To Install Kustomize, check kustomize [website](https://github.com/kubernetes-sigs/kustomize).

Note that use Kustomize 1.0.11. Version 2.0.3 has compatibility issues with kube-builder

The official kruise-controller-manager image is hosted under [docker hub](https://hub.docker.com/r/openkruise/kruise-manager).

## Usage examples

### Advanced StatefulSet
```yaml
apiVersion: apps.kruise.io/v1alpha1
kind: StatefulSet
metadata:
  name: sample
spec:
  replicas: 3
  serviceName: fake-service
  selector:
    matchLabels:
      app: sample
  template:
    metadata:
      labels:
        app: sample
    spec:
      readinessGates:
        # A new condition must be added to ensure the pod remain at NotReady state while the in-place update is happening
      - conditionType: InPlaceUpdateReady 
      containers:
      - name: main
        image: nginx:alpine
  podManagementPolicy: Parallel  # allow parallel updates, works together with maxUnavailable
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      # Do in-place update if possible, currently only image update is supported for in-place update
      podUpdatePolicy: InPlaceIfPossible
      # Allow parallel updates with max number of unavailable instances equals to 2
      maxUnavailable: 2
```
### Broadcast Job
Run a BroadcastJob that each Pod computes pi, with `ttlSecondsAfterFinished` set to 30. The job
will be deleted in 30 seconds after the job is finished.

```yaml
apiVersion: apps.kruise.io/v1alpha1
kind: BroadcastJob
metadata:
  name: broadcastjob-ttl
spec:
  template:
    spec:
      containers:
        - name: pi
          image: perl
          command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
  completionPolicy:
    type: Always
    ttlSecondsAfterFinished: 30
```
### SidecarSet

The yaml file below describes a SidecarSet that contains a sidecar container named `sidecar1`

```yaml
# sidecarset.yaml
apiVersion: apps.kruise.io/v1alpha1
kind: SidecarSet
metadata:
  name: test-sidecarset
spec:
  selector: # select the pods to be injected with sidecar containers
    matchLabels:
      app: nginx
  containers:
  - name: sidecar1
    image: centos:7
    command: ["sleep", "999d"] # do nothing at all 
```

## Developer Guide

There's a `Makefile` in the root folder which describes the options to build and install. Here are some common ones:

Build the controller manager binary

`make manager`

Run the tests

`make test`

Build the docker image, by default the image name is `openkruise/kruise-manager:v1alpha1`

`export IMG=<your_image_name> && make docker-build`

Push the image

`export IMG=<your_image_name> && make docker-push`
or just
`docker push <your_image_name>`

Generate manifests e.g. CRD, RBAC etc.

`make manifests` 

## Community

If you have any questions or want to contribute, you are welcome to join our
[slack channel](https://join.slack.com/t/kruise-workspace/shared_invite/enQtNjU5NzQ0ODcyNjYzLWMzZDI5NTM3ZjM1MGY2Mjg1NzU4ZjBjMDJmNjZmZTEwYTZkMzk4ZTAzNmY5NTczODhkZDU2NzVhM2I2MzNmODc)

Mailing List: todo

## Copyright

Certain implementation relies on existing code from Kubernetes and the credit goes to original Kubernetes authors.
