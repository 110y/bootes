// +build tools

package tools

import (
	_ "github.com/GoogleContainerTools/kpt"
	_ "github.com/go-delve/delve/cmd/dlv"
	_ "mvdan.cc/gofumpt"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "sigs.k8s.io/kind"
)
