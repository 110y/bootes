// +build test

package testutils

import (
	"path"
	"path/filepath"
	"runtime"
	"testing"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var s = k8sRuntime.NewScheme()

func SetupEnvtest(t *testing.T) client.Client {
	t.Helper()

	if err := scheme.AddToScheme(s); err != nil {
		t.Fatalf("failed to create new scheme: %s", err)
	}

	if err := apiv1.AddToScheme(s); err != nil {
		t.Fatalf("faileld to add bootes scheme: %s", err)
	}

	_, file, _, _ := runtime.Caller(0)
	testEnv := envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join(path.Dir(file), "..", "..", "..", "kubernetes", "crd", "bases")},
	}

	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("faileld to start test env: %s", err)
	}

	cli, err := client.New(cfg, client.Options{
		Scheme: s,
	})
	if err != nil {
		t.Errorf("faileld to create controller-runtime client: %s", err)

		if err := testEnv.Stop(); err != nil {
			t.Errorf("failed to stop test env: %s", err)
		}

		t.FailNow()
	}

	t.Cleanup(func() {
		if err := testEnv.Stop(); err != nil {
			t.Fatalf("failed to stop test env: %s", err)
		}
	})

	return cli
}
