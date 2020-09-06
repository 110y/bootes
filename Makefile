GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

DEV_DIR   := $(shell pwd)/dev
BIN_DIR   := $(DEV_DIR)/bin
TOOLS_DIR := $(DEV_DIR)/tools
TOOLS_SUM := $(TOOLS_DIR)/go.sum

KUBEBUILDER_VERSION := 2.3.1
KUBEBUILDER_DIR     := $(DEV_DIR)/kubebuilder
KUBEBUILDER_ASSETS  := $(KUBEBUILDER_DIR)/bin
KUBEBUILDER         := $(KUBEBUILDER_ASSETS)/kubebuilder

KUBECTL_VERSION  := 1.18.6
SKAFFOLD_VERSION := 1.14.0

CONTROLLER_GEN := $(abspath $(BIN_DIR)/controller-gen)
TYPE_SCAFFOLD  := $(abspath $(BIN_DIR)/type-scaffold)
KIND           := $(abspath $(BIN_DIR)/kind)
KUBECTL        := $(abspath $(BIN_DIR)/kubectl)-$(KUBECTL_VERSION)
SKAFFOLD       := $(abspath $(BIN_DIR)/skaffold)-$(SKAFFOLD_VERSION)
KPT            := $(abspath $(BIN_DIR)/kpt)
DELVE          := $(abspath $(BIN_DIR)/dlv)
GOFUMPT        := $(abspath $(BIN_DIR)/gofumpt)
GOLANGCI_LINT  := $(abspath $(BIN_DIR)/golangci-lint)

KIND_NODE_VERSION := 1.18.6
KIND_CLUSTER_NAME := bootes

BUILD_TOOLS := cd $(TOOLS_DIR) && go build -o

.PHONY: kubebuilder
kubebuilder: $(KUBEBUILDER)
$(KUBEBUILDER):
	@curl -sL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	@mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) $(KUBEBUILDER_DIR)

controlller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen

type-scaffold: $(TYPE_SCAFFOLD)
$(TYPE_SCAFFOLD): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(TYPE_SCAFFOLD) sigs.k8s.io/controller-tools/cmd/type-scaffold

kind: $(KIND)
$(KIND): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(KIND) sigs.k8s.io/kind

kubectl: $(KUBECTL)
$(KUBECTL):
	@curl -Lso $(KUBECTL) https://storage.googleapis.com/kubernetes-release/release/v$(KUBECTL_VERSION)/bin/$(GOOS)/$(GOARCH)/kubectl
	@chmod +x $(KUBECTL)

skaffold: $(SKAFFOLD)
$(SKAFFOLD):
	@curl -Lso $(SKAFFOLD) https://storage.googleapis.com/skaffold/releases/v$(SKAFFOLD_VERSION)/skaffold-$(GOOS)-$(GOARCH)
	@chmod +x $(SKAFFOLD)

kpt: $(KPT)
$(KPT): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(KPT) github.com/GoogleContainerTools/kpt

delve: $(DELVE)
$(DELVE): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(DELVE) github.com/go-delve/delve/cmd/dlv

gofumpt: $(GOFUMPT)
$(GOFUMPT): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(GOFUMPT) mvdan.cc/gofumpt

golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(GOLANGCI_LINT) github.com/golangci/golangci-lint/cmd/golangci-lint

# .PHONY: manifests
# manifests: $(CONTROLLER_GEN)
#         @$(CONTROLLER_GEN) crd paths=./internal/k8s/api/... output:crd:dir=./kubernetes/kpt output:stdout

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
	# NOTE: since skaffold is using kind and kubectl from PATH directly, override PATH to use project local executables.
	@$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME)
	@PATH=$${PWD}/dev/bin:$${PATH} $(SKAFFOLD) dev --filename=./dev/skaffold/skaffold.yaml

.PHONY: run-debug
run-debug: $(SKAFFOLD)
	# NOTE: since skaffold is using kind and kubectl from PATH directly, override PATH to use project local executables.
	@$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME)
	@PATH=$${PWD}/dev/bin:$${PATH} $(SKAFFOLD) debug --filename=./dev/skaffold/skaffold.yaml --port-forward=true

.PHONY: debug
debug: $(DELVE)
	@$(DELVE) connect --init=./dev/delve/init localhost:56268

.PHONY: fmt
fmt: $(GOFUMPT)
	@! $(GOFUMPT) -s -d ./ | grep -E '^'

.PHONY: lint
lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run --config ./.golangci.yml ./...

.PHONY: test
test:
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) go test -count=1 -race --tags=test ./...

.PHONY: kind-apply-manifests
kind-apply-manifests: $(KUBECTL)
	@$(KUBECTL) apply -f ./kubernetes/kpt/namespace.yaml
	@$(KUBECTL) apply -f ./kubernetes/kpt/crd/
	@$(KUBECTL) apply -f ./kubernetes/kpt/role/
	@$(KUBECTL) apply -f ./dev/kind/namespace.yaml
	sleep 15 # wait for namespace booting
	@$(KUBECTL) apply -f ./dev/kind/manifest/
