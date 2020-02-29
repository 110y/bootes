package controller_test

import (
	"path/filepath"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// func TestClusterReconciler_Reconcile(t *testing.T) {
//     cli, teardown := setup(t)
//     defer teardown()

//     tests := map[string]struct{}{}

//     for name, test := range tests {
//         test := test
//         t.Run(name, func(t *testing.T) {
//         })
//     }
// }

func setup(t *testing.T) (client.Client, func()) {
	t.Helper()

	testEnv := envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "..", "kubernetes", "crd", "bases")},
	}

	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("faileld to start test env: %s", err)
	}

	cli, err := client.New(cfg, client.Options{})
	if err != nil {
		t.Errorf("faileld to create controller-runtime client: %s", err)

		if err := testEnv.Stop(); err != nil {
			t.Errorf("failed to stop test env: %s", err)
		}

		t.FailNow()
	}

	return cli, func() {
		if err := testEnv.Stop(); err != nil {
			t.Fatalf("failed to stop test env: %s", err)
		}
	}
}
