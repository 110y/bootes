CONTROLLER_GEN := bin/controller-gen
TYPE_SCAFFOLD  := bin/type-scaffold
KIND           := bin/kind
KUBECTL        := bin/kubectl

KIND_NODE_VERSION := v1.17.2

$(CONTROLLER_GEN): go.sum
	@go build -o $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen

$(TYPE_SCAFFOLD): go.sum
	@go build -o $(TYPE_SCAFFOLD) sigs.k8s.io/controller-tools/cmd/type-scaffold

$(KIND): go.sum
	@go build -o $(KIND) sigs.k8s.io/kind

$(KUBECTL): .kubectl-version
	@curl -Lso $(KUBECTL) https://storage.googleapis.com/kubernetes-release/release/$(shell cat .kubectl-version)/bin/$(shell go env GOOS)/$(shell go env GOARCH)/kubectl
	@chmod +x $(KUBECTL)

.PHONY: manifests
manifests: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) crd paths=./internal/k8s/api/... output:crd:dir=./kubernetes/crd/bases output:stdout

.PHONY: deepcopy
deepcopy: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) object paths=./internal/k8s/api/...

.PHONY: kind-cluster
kind-cluster: $(KIND) $(KUBECTL)
	@$(KIND) delete cluster --name bootes
	@$(KIND) create cluster --name bootes --image kindest/node:${KIND_NODE_VERSION}
	@$(KUBECTL) apply -f ./kubernetes/crd/bases/labolith.com_clusters.yaml
