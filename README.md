# OpenShift Liveness and Readiness Probes

This repository is meant to demonstrate the [liveness and readiness probes in OpenShift](https://www.openshift.com/blog/liveness-and-readiness-probes)
and how they interact with the [service](https://docs.openshift.com/container-platform/3.11/architecture/core_concepts/pods_and_services.html) abstraction.

It relies on [Minishift](https://docs.okd.io/3.11/minishift/index.html) to run
OpenShift locally. The code in the repository was tested using Minishift
version 1.34 (OKD/OpenShift version 3.11).

The demonstration in this repo most likely applies to apps deployed on
Kubernetes as OpenShift [is a Kubernetes distribution](https://www.redhat.com/en/blog/openshift-and-kubernetes-whats-difference).

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/)
- [Setting up the virtualization environment](https://docs.okd.io/3.11/minishift/getting-started/setting-up-virtualization-environment.html)
- [Minishift (OKD 3.11)](https://docs.okd.io/3.11/minishift/getting-started/installing.html)

## Getting Started

Make sure you have all the [prerequisites](#prerequisites) installed and
configured. We are now going to deploy the sample [app](./app) in this
repository on Minishift.

First start Minishift.

```sh
minishift start
```

To access the [OpenShift CLI](https://docs.okd.io/3.11/minishift/openshift/openshift-client-binary.html)
execute

```sh
eval $(minishift oc-env)
```

Create and switch to a new OpenShift project where we will deploy the app.

```sh
oc login --username developer --password doesnotmatter
oc new-project probes \
    --description="project for playing around with probes" \
    --display-name="probes"
```

Create an OpenShift [service with pods](https://docs.openshift.com/container-platform/3.11/architecture/core_concepts/pods_and_services.html)
and an [image stream](https://docs.openshift.com/container-platform/3.11/dev_guide/managing_images.html) for our apps Docker image.

```sh
oc new-app --file app/app.yml
```

Build and push the Docker image for the app.

```sh
cd app
./build-image.sh
```

Check that a `Deployment` was created due to the new Docker image
(`ImageChange` trigger).

```sh
oc describe dc/probes
```

You should see something like

```
Deployment #1 (latest):
	Name:		probes-1
	Created:	3 seconds ago
	Status:		Running
	Replicas:	2 current / 2 desired
	Selector:	deployment=probes-1,deploymentconfig=probes
	Labels:		app=probes,openshift.io/deployment-config.name=probes
	Pods Status:	0 Running / 2 Waiting / 0 Succeeded / 0 Failed
```

You can now send requests to the apps [service](https://docs.openshift.com/container-platform/3.11/architecture/core_concepts/pods_and_services.html)
via

```sh
curl --silent --include http://$(minishift ip):30080/pod
```

Note that the port is specified in the [./app/app.yml](./app/app.yml) service
`spec.ports[].nodePort`. You might also need to wait a little until at least
one pod is `ready`. You can check that with
`oc get pods --selector app=probes`.

You'll see that if you repeat the request a few times it will be handled
by a different pod. This is due to the service acting as an internal load
balancer.

## Probes

The following describes the type of probes, their purpose and how you can try
them out in our app.

Make sure you completed the [Getting Started](#getting-started) section before
moving on.

### App Endpoints

Once you have running pods you can access the following endpoints our app
provides via HTTP.

| METHOD | PATH          | DESCRIPTION                                       |
| :----- | :------------ | :------------------------------------------------ |
| GET    | /pod          | Responds with the pod name processing the request |
| GET    | /live         | Liveness probe                                    |
| POST   | /live/toggle  | Toggles the state of the liveness probe           |
| GET    | /ready        | Readiness probe                                   |
| POST   | /ready/toggle | Toggle the state of the readiness probe           |

To target any of our deployed pods execute

```sh
oc rsh dc/probes curl --silent --include http://localhost:8080/pod
```

To target a specific one run

```sh
oc exec <fill_in_podname> -- curl --silent --include http://localhost:8080/pod
```

Fill in the name of one of the running pods `oc get pods --selector app=probes`.

To talk to the service which routes traffic to the pods execute

```sh
curl --silent --include http://$(minishift ip):30080/pod
```

### Readiness

> Sometimes, applications are temporarily unable to serve traffic. For example,
> an application might need to load large data or configuration files during
> startup, or depend on external services after startup.
> -- <cite>[kubernetes.io](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes)</cite>

This is why we can declare a [readiness probe](https://docs.openshift.com/container-platform/3.11/dev_guide/application_health.html#dev-guide-application-health).
The readiness probe tells OpenShift when the container is `ready` to receive
traffic. Once a container is `unready` it will not get any traffic from the
service we defined in our [./app/app.yml](./app/app.yml).

OpenShift will call the probe defined in your containers spec at regular
intervals. There are a few things you can configure, which you can find
[here](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes).

The probe will be called during the entire lifetime of the pod! So a pod can
go from `unready` to `ready` many times during its existence.

You can try it out for yourself by executing the following scripts in separate
terminal windows

- [watch-probes.sh](./watch-probes.sh)
- [send-to-service.sh](./send-to-service.sh)

Arrange the windows side-by-side so you can follow the changes in a pods state
and which pods gets traffic closely.

Pick a pods name

```sh
oc get pods --selector app=probes
```

And toggle it to `unready`

```sh
oc exec <fill_in_podname> -- curl --request POST --silent --include http://localhost:8080/ready/toggle
```

After toggling the readiness state one of the pods should show up as `unready`
(READY 0/1) in the terminal running `watch-probes.sh`.

```
Every 1.0s: oc -n probes get pods

NAME             READY     STATUS    RESTARTS   AGE
probes-1-bd2wt   0/1       Running   0          1h
probes-1-w5rvp   1/1       Running   0          1h
```

From now on `send-to-service.sh` should show you that only the pod that is
`ready` is processing requests from the service.

You can try toggling the readiness state back to `ready` and observe the pod
getting traffic again.

Try setting both pods to `unready` and see what happens 🙃️

### Liveness

> Many applications running for long periods of time eventually transition to
> broken states, and cannot recover except by being restarted. Kubernetes
> provides liveness probes to detect and remedy such situations. --
> <cite>[kubernetes.io](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command)</cite>

This is why we can declare a [liveness probe](https://docs.openshift.com/container-platform/3.11/dev_guide/application_health.html#dev-guide-application-health)
for our containers. The liveness probe tells OpenShift whether it should
restart the container.

To see the effects of a failing liveness probe follow the steps at
[readiness probe](#readiness). Exchange `ready` with `live` in the URLs.

You should see that a container gets restarted when it is not `live`. The
`RESTARTS` counter in `oc get pods` should also reflect that. While the
container is starting and is not yet `ready` it will also not get any traffic.

Try setting both pods to not `live` and see what happens 🙃️
You can also first toggle a pods `ready` state to `unready` and then toggle its
`live` state. It should be restarted and receive traffic again once its
`ready`.

### Useful Commands

Check the logs to see when probes are invoked.

```sh
oc logs --follow svc/probes
Found 2 pods, using pod/probes-4-hpdtw
probes-4-hpdtw 2021/05/30 11:47:32 Listening on port 8080...
probes-4-hpdtw 2021/05/30 11:47:51 Readyness probe invoked. Pod is ready.
probes-4-hpdtw 2021/05/30 11:47:55 Liveness probe invoked. Pod is live.
probes-4-hpdtw 2021/05/30 11:48:01 Readyness probe invoked. Pod is ready.
probes-4-hpdtw 2021/05/30 11:48:05 Liveness probe invoked. Pod is live.
```

Note that this will pick one pod out of your replicas.

Stream OpenShift events

```sh
oc --namespace probes get events
```

which shows you when probes failed, containers are restarted, killed, ...

## Further Reading & Exploration

For more in-depth information read

[OpenShift Application Health](https://docs.openshift.com/container-platform/3.11/dev_guide/application_health.html#dev-guide-application-health)

[Kubernetes Configure Liveness, Readiness and Startup Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)

### Startup Probe

Kubernetes added a [startup probe](https://github.com/kubernetes/enhancements/issues/950) which helps
with [slow starting pods](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-startup-probes).

OpenShift 3.11 does not provide a startup probe so you would need to adapt the
example here to Kubernetes to play with that. You can adjust the code or the
probe configuration to see how a slow starting pod is affected by readiness and
liveness probes.

## Start Over

If you want to delete all resources you created in the walk-through and start
over execute

```sh
oc delete project probes
```
