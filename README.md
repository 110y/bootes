# Bootes

Minimalistic control plane for `Kubernetes` backed `Envoy` data planes.

## Overview

Bootes is implemented as a Kubernetes Controller and you can specify any Envoy configurations via Custom Resources like below:

```yaml
---
apiVersion: bootes.io/v1
kind: Cluster
metadata:
  name: cluster-1
  namespace: envoy
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

By applying this example resource, Bootes sends one cluster configuration named `cluster-1` to connected Envoys.
