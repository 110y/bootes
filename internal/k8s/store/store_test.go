package store_test

import (
	"context"
	"testing"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/k8s/testutils"
)

func TestGetCluster(t *testing.T) {
	tests := map[string]struct {
		cluster   *apiv1.Cluster
		name      string
		namespace string
	}{
		"": {
			name:      "foo",
			namespace: "bar",
		},
	}

	cli, teardown := testutils.SetupEnvtest(t)
	defer teardown()

	s := store.NewStore(cli)

	ctx := context.Background()
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			_, err := s.GetCluster(ctx, test.name, test.namespace)
			if err != nil {
				t.Errorf("failed %s", err)
			}
		})
	}
}

func TestListCluster(t *testing.T) {
	tests := map[string]struct {
		cluster   *apiv1.Cluster
		name      string
		namespace string
	}{
		"": {
			name:      "foo",
			namespace: "bar",
		},
	}

	cli, teardown := testutils.SetupEnvtest(t)
	defer teardown()

	s := store.NewStore(cli)

	ctx := context.Background()
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			_, err := s.ListPodsByNamespace(ctx, test.namespace)
			if err != nil {
				t.Errorf("failed %s", err)
			}
		})
	}
}
