# Installation

What we have to do are just fetching and applying the Kubernetes manifests for Bootes like below:

```
> git clone https://github.com/110y/bootes.git ./bootes

> kustomize build ./bootes/kubernetes/kpt | kubectl apply --filename -
```

But, to make the installation, upgrade and configuration processes to be more easier, it's recommend to use [KPT](https://googlecontainertools.github.io/kpt/) as described in the next section.

## Using [KPT](https://googlecontainertools.github.io/kpt/)

Bootes uses KPT to package its Kubernetes manifests. Instead of just using `git`, we can also use KPT to manage Kubernetes manifests for Bootes. This section explains how to use KPT to install Bootes in step by step.

### Installation

First, fetch the Kubernetes manifests by using the following command:

```
> kpt pkg get https://github.com/110y/bootes/kubernetes/kpt ./bootes
```

This command will fetch its Kubernetes manifests into the `./bootes` directory. It's highly recommended to mange this directory with `git`. So, we should execute the following commands:

```
> cd ./bootes
> git init && git add . && git commit -m "initialize"
```

After that, install Bootes by executing the following command.

```
> make install

namespace/bootes created
customresourcedefinition.apiextensions.k8s.io/clusters.bootes.io created
customresourcedefinition.apiextensions.k8s.io/endpoints.bootes.io created
customresourcedefinition.apiextensions.k8s.io/listeners.bootes.io created
customresourcedefinition.apiextensions.k8s.io/routes.bootes.io created
clusterrole.rbac.authorization.k8s.io/bootes-manager created
clusterrole.rbac.authorization.k8s.io/bootes-pod-reader created
clusterrolebinding.rbac.authorization.k8s.io/bootes-manager created
clusterrolebinding.rbac.authorization.k8s.io/bootes-pod-reader created
service/bootes created
deployment.apps/bootes created
```

To ensure that the installation has been succeeded, execute the following command and we can see the Pods of Bootes are ready:

```
> kubectl get pods --namespace bootes
```

### Configuration

Bootes provides some configuration points. To see what we can configure, execute the following command:

```
> kpt cfg list-setters ./bootes/kubernetes/kpt

bootes/kubernetes/kpt/
             NAME              VALUE    SET BY            DESCRIPTION             COUNT   REQUIRED   IS SET
  cpu-limit                    2                 CPU limit for each Pod           1       No         No
  cpu-request                  1                 CPU request for each Pod         1       No         No
  enable-gcp-cloud-trace       false             Use Cloud Trace or not           1       No         No
  enable-jaeger-trace          false             Use Jeager or not                1       No         No
  enable-stdout-trace          false             Use STDOUT trace or not          1       No         No
  enable-xds-grpc-channelz     false             Enable gRPC channelz or not      1       No         No
  enable-xds-grpc-reflection   false             Enable gRPC reflection or not    1       No         No
  gcp-cloud-trace-project-id                     GCP Project ID for Cloud Trace   1       No         No
  jaeger-trace-endpoint                          Jaeger endpoint                  1       No         No
  memory-limit                 4Gi               Memory limit for each Pod        1       No         No
  memory-request               2Gi               Memory request for each Pod      1       No         No
  namespace                    bootes            Namespace                        6       No         No
  replicas                     2                 Number of Pod replicas           1       No         No
  version                      1.0.0             Version                          1       No         No
```

With KPT, we can easily modify some configurations.
For example, if we want to change the namespace used by Bootes (default is `bootes`), execute the following command:

```
> kpt cfg set ./bootes/kubernetes/kpt namespace your-bootes-namespace
```
