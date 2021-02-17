OS   := $(shell uname | awk '{print tolower($$0)}')
ARCH := $(shell case $$(arch) in (x86_64) echo -n amd64 ;; (aarch64) echo -n arm64 ;; (*) echo -n $$(arch) ;; esac)

BIN_DIR := bin

KUBECTL_VERSION   := 1.20.2
KUSTOMIZE_VERSION := 3.10.0

KUBECTL   := $(abspath $(BIN_DIR)/kubectl)-$(KUBECTL_VERSION)
KUSTOMIZE := $(abspath $(BIN_DIR)/kustomize)-$(KUSTOMIZE_VERSION)

kubectl: $(KUBECTL)
$(KUBECTL):
	@curl -Lso $(KUBECTL) https://storage.googleapis.com/kubernetes-release/release/v$(KUBECTL_VERSION)/bin/$(OS)/$(ARCH)/kubectl
	@chmod +x $(KUBECTL)

kustomize: $(KUSTOMIZE)
$(KUSTOMIZE):
	@curl -sSL "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v$(KUSTOMIZE_VERSION)/kustomize_v$(KUSTOMIZE_VERSION)_$(OS)_$(ARCH).tar.gz" | tar -C $(BIN_DIR) -xzv kustomize > /dev/null 2>&1
	@mv $(BIN_DIR)/kustomize $(KUSTOMIZE)

.PHONY: install
install: $(KUBECTL) $(KUSTOMIZE)
	@$(KUSTOMIZE) build . | $(KUBECTL) apply --filename -
