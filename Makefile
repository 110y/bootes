CONTROLLER_GEN := bin/controller-gen

manifests: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) crd paths=./internal/k8s/api/... output:crd:dir=./kubernetes/crd/bases output:stdout

deepcopy: $(CONTROLLER_GEN)
	@$(CONTROLLER_GEN) object paths=./internal/k8s/api/...

$(CONTROLLER_GEN): go.sum
	@go build -o $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen

bin/type-scaffold: go.sum
	@go build -o ./bin/type-scaffold sigs.k8s.io/controller-tools/cmd/type-scaffold
