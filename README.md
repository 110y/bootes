# Bootes

xDS Control-Plane Kubernetes Controller.

## Overview

Bootes is a minimalistic xDS Control-Plane which is implemented as a Kubernetes Controller.
You can distribute any configurations via Kubernetes Custom Resources like below:

```yaml
---
apiVersion: bootes.io/v1
kind: Cluster
metadata:
  name: cluster-1
  namespace: test
spec:
  config:
    name: cluster-1
    connect_timeout: 1s
    type: LOGICAL_DNS
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    load_assignment:
      cluster_name: cluster-1
      endpoints:
        - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: cluster-1.test.svc.cluster.local
                    port_value: 10000
```

By applying this example resource, Bootes sends one cluster configuration named `cluster-1` to connected data-planes.

## Installation

See (this guide)[https://github.com/110y/bootes/blob/master/kubernetes/kpt/README.md].

## Supported Resource Types

- [x] Listener
- [x] Route
- [x] Cluster
- [x] Endpoint
- [ ] VirtualHost
- [ ] Secret
- [ ] Runtime
- [ ] ScopedRoute
