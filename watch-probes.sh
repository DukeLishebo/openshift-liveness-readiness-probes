#!/bin/sh

eval $(minishift oc-env)
oc login --username developer --password doesnotmatter
watch --differences --interval 0.5 \
    oc --namespace probes get pods
