manifests: bin/controller-gen
	@./bin/controller-gen crd paths=./internal/k8s/api/v1/... output:crd:dir=./kubernetes/crd/bases output:stdout

bin/controller-gen: go.mod go.sum
	@go build -o ./bin/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen
