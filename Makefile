GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

KUBEBUILDER_VERSION := 2.2.0
KUBEBUILDER_DIR     := $(shell pwd)/kubebuilder
KUBEBUILDER_ASSETS  := $(KUBEBUILDER_DIR)/bin
KUBEBUILDER         := $(KUBEBUILDER_ASSETS)/kubebuilder

CONTROLLER_GEN := bin/controller-gen
TYPE_SCAFFOLD  := bin/type-scaffold
KIND           := bin/kind
KUBECTL        := bin/kubectl
SKAFFOLD       := bin/skaffold

KIND_NODE_VERSION := 1.17.2

.PHONY: kubebuilder
kubebuilder: $(KUBEBUILDER)
$(KUBEBUILDER):
	@curl -sL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	@mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) $(KUBEBUILDER_DIR)

controlller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): go.sum
	@go build -o $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen

type-scaffold: $(TYPE_SCAFFOLD)
$(TYPE_SCAFFOLD): go.sum
	@go build -o $(TYPE_SCAFFOLD) sigs.k8s.io/controller-tools/cmd/type-scaffold

kind: $(KIND)
$(KIND): go.sum
	@go build -o $(KIND) sigs.k8s.io/kind

kubectl: $(KUBECTL)
$(KUBECTL): .kubectl-version
	@curl -Lso $(KUBECTL) https://storage.googleapis.com/kubernetes-release/release/$(shell cat .kubectl-version)/bin/$(GOOS)/$(GOARCH)/kubectl
	@chmod +x $(KUBECTL)

skaffold: $(SKAFFOLD)
$(SKAFFOLD): .skaffold-version
	@curl -Lso $(SKAFFOLD) https://storage.googleapis.com/skaffold/releases/$(shell cat .skaffold-version)/skaffold-$(GOOS)-$(GOARCH)
	@chmod +x $(SKAFFOLD)

.PHONY: manifests
manifests: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) crd paths=./internal/k8s/api/... output:crd:dir=./kubernetes/crd/bases output:stdout

.PHONY: deepcopy
deepcopy: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) object paths=./internal/k8s/api/...

.PHONY: kind-cluster
kind-cluster: $(KIND) $(KUBECTL)
	@$(KIND) delete cluster --name bootes
	@$(KIND) create cluster --name bootes --image kindest/node:v${KIND_NODE_VERSION}
	make kind-image
	make apply-manifests

.PHONY: kind-image
kind-image: $(KIND)
	@docker build -t 110y/bootes-envoy:latest ./kind/envoy
	$(KIND) load docker-image 110y/bootes-envoy:latest --name bootes

.PHONY: dev
dev: $(SKAFFOLD)
	# NOTE: since skaffold is using kind from PATH directly, override PATH to use project local kind executable.
	@$(KUBECTL) config use-context kind-bootes
	@PATH=$${PWD}/bin:$${PATH} $(SKAFFOLD) dev --filename=./skaffold/skaffold.yaml

.PHONY: debug
debug: $(SKAFFOLD)
	# NOTE: since skaffold is using kind from PATH directly, override PATH to use project local kind executable.
	@$(KUBECTL) config use-context kind-bootes
	@PATH=$${PWD}/bin:$${PATH} $(SKAFFOLD) debug --filename=./skaffold/skaffold.yaml

.PHONY: test
test:
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) go test -race --tags=test ./...

.PHONY: apply-manifests
apply-manifests: $(KUBECTL)
	@$(KUBECTL) apply -f ./kubernetes/crd/bases/
	@$(KUBECTL) apply -f ./kind/namespace.yaml
	sleep 15 # wait for namespace booting
	@$(KUBECTL) apply -f ./kind/manifest/
