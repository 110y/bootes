# Bootes

A xDS Control-Plane Kubernetes Controller.

## Overview

Bootes is a minimalistic xDS Control-Plane which is implemented as a Kubernetes Controller.
You can write any xDS configurations as Kubernetes Custom Resources like below:

```yaml
---
apiVersion: bootes.io/v1
kind: Cluster
metadata:
  name: my-cluster
  namespace: my-namespace
spec:
  config:
    name: my-cluster
    connect_timeout: 1s
    type: LOGICAL_DNS
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    load_assignment:
      cluster_name: my-cluster
      endpoints:
        - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: my-cluster.my-namespace.svc.cluster.local
                    port_value: 10000
```

By applying this `Cluster` resource, Bootes sends the cluster configuration named `my-cluster` to connected data-planes.

## Installation

See [this guide](https://github.com/110y/bootes/blob/master/doc/installation.md).

## How to use

See [this guide](https://github.com/110y/bootes/blob/master/doc/how-to-connect.md).

## Supported Resource Types

- [x] Listener
- [x] Route
- [x] Cluster
- [x] Endpoint
- [ ] VirtualHost
- [ ] Secret
- [ ] Runtime
- [ ] ScopedRoute
