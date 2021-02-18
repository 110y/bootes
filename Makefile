GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

DEV_DIR   := $(shell pwd)/dev
BIN_DIR   := $(DEV_DIR)/bin
TOOLS_DIR := $(DEV_DIR)/tools
TOOLS_SUM := $(TOOLS_DIR)/go.sum

KUBERNETES_VERSION     := 1.20.2
KUBE_APISERVER_VERSION := 1.19.8
ETCD_VERSION           := 3.4.9

KUBEBUILDER_VERSION := 2.3.2
KUBEBUILDER_DIR     := $(DEV_DIR)/kubebuilder
KUBEBUILDER_ASSETS  := $(KUBEBUILDER_DIR)/bin
KUBEBUILDER         := $(KUBEBUILDER_ASSETS)/kubebuilder

KPT_VERSION      := 0.37.1
SKAFFOLD_VERSION := 1.19.0
KIND_VERSION     := 0.10.0

CONTROLLER_GEN := $(abspath $(BIN_DIR)/controller-gen)
TYPE_SCAFFOLD  := $(abspath $(BIN_DIR)/type-scaffold)
KUBE_APISERVER := $(abspath $(BIN_DIR)/kube-apiserver)-$(KUBERNETES_VERSION)
ETCD           := $(abspath $(BIN_DIR)/etcd)-$(ETCD_VERSION)
KIND           := $(abspath $(BIN_DIR)/kind)-$(KIND_VERSION)
KUBECTL        := $(abspath $(BIN_DIR)/kubectl)-$(KUBERNETES_VERSION)
SKAFFOLD       := $(abspath $(BIN_DIR)/skaffold)-$(SKAFFOLD_VERSION)
KPT            := $(abspath $(BIN_DIR)/kpt)-$(KPT_VERSION)
DELVE          := $(abspath $(BIN_DIR)/dlv)
GOFUMPT        := $(abspath $(BIN_DIR)/gofumpt)
GOLANGCI_LINT  := $(abspath $(BIN_DIR)/golangci-lint)

KIND_CLUSTER_NAME := bootes

BUILD_TOOLS := cd $(TOOLS_DIR) && go build -o

.PHONY: kubebuilder
kubebuilder: $(KUBEBUILDER)
$(KUBEBUILDER):
	@curl -sL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	@mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) $(KUBEBUILDER_DIR)

controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen

type-scaffold: $(TYPE_SCAFFOLD)
$(TYPE_SCAFFOLD): $(TOOLS_SUM)
	@$(BUILD_TOOLS) $(TYPE_SCAFFOLD) sigs.k8s.io/controller-tools/cmd/type-scaffold

kube-apiserver: $(KUBE_APISERVER)
$(KUBE_APISERVER):
	@curl -sSL "https://dl.k8s.io/v$(KUBE_APISERVER_VERSION)/kubernetes-server-$(GOOS)-$(GOARCH).tar.gz" | tar -C /tmp -xzv kubernetes/server/bin/kube-apiserver
	@mv /tmp/kubernetes/server/bin/kube-apiserver $(KUBE_APISERVER)
	@cp $(KUBE_APISERVER) $(BIN_DIR)/kube-apiserver

etcd: $(ETCD)
$(ETCD):
	@curl -sSL "https://github.com/etcd-io/etcd/releases/download/v$(ETCD_VERSION)/etcd-v$(ETCD_VERSION)-$(GOOS)-$(GOARCH).tar.gz" | tar -C /tmp -xzv etcd-v$(ETCD_VERSION)-$(GOOS)-$(GOARCH)/etcd
	@mv /tmp/etcd-v$(ETCD_VERSION)-$(GOOS)-$(GOARCH)/etcd $(ETCD)
	@cp $(ETCD) $(BIN_DIR)/etcd

kind: $(KIND)
$(KIND):
	@curl -Lso $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(GOOS)-$(GOARCH)
	@chmod +x $(KIND)
	@cp $(KIND) $(BIN_DIR)/kind # to use this in skaffold and which implicitly picks up from $PATH

kubectl: $(KUBECTL)
$(KUBECTL):
	@curl -Lso $(KUBECTL) https://storage.googleapis.com/kubernetes-release/release/v$(KUBERNETES_VERSION)/bin/$(GOOS)/$(GOARCH)/kubectl
	@chmod +x $(KUBECTL)
	@cp $(KUBECTL) $(BIN_DIR)/kubectl # to use this in skaffold and which implicitly picks up from $PATH

skaffold: $(SKAFFOLD)
$(SKAFFOLD):
	@curl -Lso $(SKAFFOLD) https://storage.googleapis.com/skaffold/releases/v$(SKAFFOLD_VERSION)/skaffold-$(GOOS)-$(GOARCH)
	@chmod +x $(SKAFFOLD)

kpt: $(KPT)
$(KPT):
	@curl -sSL "https://github.com/GoogleContainerTools/kpt/releases/download/v$(KPT_VERSION)/kpt_$(GOOS)_$(GOARCH)-$(KPT_VERSION).tar.gz" | tar -C /tmp -xzv kpt
	@mv /tmp/kpt $(KPT)

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
	@$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --image kindest/node:v${KUBERNETES_VERSION}
	@make kind-image
	@make kind-apply-manifests

.PHONY: kind-image
kind-image: $(KIND)
	@docker build -t 110y/bootes-envoy:latest ./dev/kind/envoy
	$(KIND) load docker-image 110y/bootes-envoy:latest --name $(KIND_CLUSTER_NAME)

.PHONY: run
run: $(SKAFFOLD)
	# NOTE: since skaffold is using kind and kubectl from PATH directly, override PATH to use project local executables.
	@$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME)
	@PATH=$${PWD}/dev/bin:$${PATH} $(SKAFFOLD) dev --tail --filename=./dev/skaffold/skaffold.yaml

.PHONY: run-debug
run-debug: $(SKAFFOLD)
	# NOTE: since skaffold is using kind and kubectl from PATH directly, override PATH to use project local executables.
	@$(KUBECTL) config use-context kind-$(KIND_CLUSTER_NAME)
	@PATH=$${PWD}/dev/bin:$${PATH} $(SKAFFOLD) debug --tail --filename=./dev/skaffold/skaffold.yaml --port-forward=true

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
test: $(KUBE_APISERVER) $(ETCD)
	go test -count=1 -race --tags=test ./...

.PHONY: kind-apply-manifests
kind-apply-manifests: $(KUBECTL)
	@$(KUBECTL) apply --filename ./kubernetes/kpt/namespace.yaml
	@$(KUBECTL) apply --filename ./kubernetes/kpt/crd/
	@$(KUBECTL) apply --kustomize ./kubernetes/kpt/role/
	@$(KUBECTL) apply --kustomize ./dev/kind/manifest
