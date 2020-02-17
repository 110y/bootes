CONTROLLER_GEN := bin/controller-gen
KIND := bin/kind

$(CONTROLLER_GEN): go.sum
	@go build -o $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen

bin/type-scaffold: go.sum
	@go build -o bin/type-scaffold sigs.k8s.io/controller-tools/cmd/type-scaffold

$(KIND): go.sum
	@go build -o $(KIND) sigs.k8s.io/kind

.PHONY: manifests
manifests: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) crd paths=./internal/k8s/api/... output:crd:dir=./kubernetes/crd/bases output:stdout

.PHONY: deepcopy
deepcopy: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) object paths=./internal/k8s/api/...

.PHONY: kind-cluster
kind-cluster:
ifneq ($(shell docker ps --filter name=bootes-control-plane --quiet),)
endif
	# id=$$(docker ps --filter name=bootes-control-plane --quiet);
	# @$(KIND) create cluster --name bootes --image kindest/node:v1.16.3
