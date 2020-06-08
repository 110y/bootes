# Installation with [KPT](https://googlecontainertools.github.io/kpt/)

## Usage

```console
> kpt pkg get https://github.com/110y/bootes/kubernetes/kpt ./bootes
> kubectl apply -f ./bootes/kubernetes/kpt/namespace.yaml
> kubectl apply -f ./bootes/kubernetes/kpt/crd/
> kubectl apply -f ./bootes/kubernetes/kpt/role/
> kubectl apply -k ./bootes/kubernetes/kpt/deployment/
```
