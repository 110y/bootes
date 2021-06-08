# How to make data-planes connect to Bootes

This document explains how to make data-planes connect to Bootes control-plane.\
If you haven't install Bootes into you Kubernetes cluster yet, see [this guide](https://github.com/110y/bootes/blob/master/doc/installation.md) first.

## Envoy

### Configuration

Bootes provides `Aggregated Discovery Service` to distribute all type of resources like Cluster, Listener etc.\
So first, add a static cluster for Bootes to Envoy configration (`envoy.yaml`) like below:\

**NOTE:** The address for Bootes is up to your configration. If you've installed Bootes with default configurations, it's `bootes.bootes.svc.cluster.local:5000`.

```yaml
static_resources:
  clusters:
    - name: bootes
      connect_timeout: 1s
      type: LOGICAL_DNS
      http2_protocol_options: {}
      load_assignment:
        cluster_name: bootes
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: bootes.bootes.svc.cluster.local
                      port_value: 5000
```

And then, add `ads_config`, `cds_config` and `lds_config` to your Envoy configuration like below:

**NOTE:** Bootes provides only v3 APIs. So we have to set `V3` to `resource_api_version`.

```yaml
dynamic_resources:
  ads_config:
    api_type: GRPC
    transport_api_version: V3
    grpc_services:
      - envoy_grpc:
          cluster_name: bootes
  cds_config:
    ads: {}
    resource_api_version: V3
  lds_config:
    ads: {}
    resource_api_version: V3
```

### Deployment

To make Bootes be albe to recognize each data-planes uniquely, we have to specify `--service-node` and `--service-cluster` as Envoy command arguments.\
In Kubernetes world, Pod name and Namespace name combination is guaranteed as unique, we can use it for `--service-node`. Regarding `--service-cluster`, we can use Deployment (or ReplicaSet) name (since Deployment name can't be populated by `valueFrom.fieldRef`, we have to specify it explicitly).

```yaml
# ...
spec:
  containers:
    - name: envoy
      image: envoyproxy/envoy:latest
      command:
        - 'envoy'
      args:
        - '--config-path'
        - '/etc/envoy/envoy.yaml'
        - '--service-node'
        - '$(POD_NAME).$(NAMESPACE)'
        - '--service-cluster'
        - '$(DEPLOYMENT_NAME).$(NAMESPACE)'
      env:
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: DEPLOYMENT_NAME
          value: 'your_deployment_name'
```

## gRPC

TODO
