GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

KUBEBUILDER_VERSION := 2.2.0
KUBEBUILDER_DIR     := $(shell pwd)/dev/kubebuilder
KUBEBUILDER_ASSETS  := $(KUBEBUILDER_DIR)/bin
KUBEBUILDER         := $(KUBEBUILDER_ASSETS)/kubebuilder

BIN_DIR := dev/bin

CONTROLLER_GEN := $(BIN_DIR)/controller-gen
TYPE_SCAFFOLD  := $(BIN_DIR)/type-scaffold
KIND           := $(BIN_DIR)/kind
KUBECTL        := $(BIN_DIR)/kubectl
SKAFFOLD       := $(BIN_DIR)/skaffold
DELVE          := $(BIN_DIR)/dlv

KIND_NODE_VERSION := 1.17.2
KIND_CLUSTER_NAME := bootes

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
$(KUBECTL): dev/.kubectl-version
	@curl -Lso $(KUBECTL) https://storage.googleapis.com/kubernetes-release/release/$(shell cat ./dev/.kubectl-version)/bin/$(GOOS)/$(GOARCH)/kubectl
	@chmod +x $(KUBECTL)

skaffold: $(SKAFFOLD)
$(SKAFFOLD): dev/.skaffold-version
	@curl -Lso $(SKAFFOLD) https://storage.googleapis.com/skaffold/releases/$(shell cat ./dev/.skaffold-version)/skaffold-$(GOOS)-$(GOARCH)
	@chmod +x $(SKAFFOLD)

delve: $(DELVE)
$(DELVE): go.sum
	@go build -o $(DELVE) github.com/go-delve/delve/cmd/dlv

# .PHONY: manifests
# manifests: $(CONTROLLER_GEN)
#         @$(CONTROLLER_GEN) crd paths=./internal/k8s/api/... output:crd:dir=./kubernetes/crd/bases output:stdout

.PHONY: deepcopy
deepcopy: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) object paths=./internal/k8s/api/...

.PHONY: kind-cluster
kind-cluster: $(KIND) $(KUBECTL)
	@$(KIND) delete cluster --name $(KIND_CLUSTER_NAME)
	@$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --image kindest/node:v${KIND_NODE_VERSION}
	make kind-image
	make kind-apply-manifests

.PHONY: kind-image
kind-image: $(KIND)
	@docker build -t 110y/bootes-envoy:latest ./dev/kind/envoy
	$(KIND) load docker-image 110y/bootes-envoy:latest --name $(KIND_CLUSTER_NAME)

.PHONY: run
run: $(SKAFFOLD)
	# NOTE: since skaffold is using kind and kubectl from PATH directly, override PATH to use project local kind executable.
	@$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME)
	@PATH=$${PWD}/dev/bin:$${PATH} $(SKAFFOLD) dev --filename=./dev/skaffold/skaffold.yaml

.PHONY: run-debug
run-debug: $(SKAFFOLD)
	# NOTE: since skaffold is using kind and kubectl from PATH directly, override PATH to use project local kind executable.
	@$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME)
	@PATH=$${PWD}/dev/bin:$${PATH} $(SKAFFOLD) debug --filename=./dev/skaffold/skaffold.yaml --port-forward=true

.PHONY: debug
debug: $(DELVE)
	@$(DELVE) connect --init=./dev/delve/init localhost:56268

.PHONY: test
test:
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) go test -count=1 -race --tags=test ./...

.PHONY: kind-apply-manifests
kind-apply-manifests: $(KUBECTL)
	@$(KUBECTL) apply -f ./kubernetes/crd/bases/
	@$(KUBECTL) apply -f ./dev/kind/namespace.yaml
	sleep 15 # wait for namespace booting
	@$(KUBECTL) apply -f ./dev/kind/manifest/
