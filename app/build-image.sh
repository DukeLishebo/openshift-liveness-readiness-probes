#!/bin/sh
set -ue

# Get access to the OpenShift client https://docs.okd.io/3.11/minishift/openshift/openshift-client-binary.html
eval $(minishift oc-env)
oc login --username developer --password doesnotmatter
oc project probes
# Connect Docker client to Minishift Docker daemon https://docs.okd.io/3.11/minishift/using/docker-daemon.html
eval $(minishift docker-env)
# Build and push Docker image of app to Minishift Docker registry
docker login --username developer --password $(oc whoami -t) $(minishift openshift registry)
docker build --tag probes .
docker tag probes $(minishift openshift registry)/probes/probes
docker push $(minishift openshift registry)/probes/probes
oc describe is/probes
