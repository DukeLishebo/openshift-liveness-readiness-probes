#!/bin/sh

eval $(minishift oc-env)
# Ensure port is equal to app/app.yml Service .spec.ports[0].nodePort
watch --differences --interval 0.5 \
    curl --silent "http://$(minishift ip):30080/pod"
