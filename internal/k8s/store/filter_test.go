package store_test

import (
	"testing"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterClusters(t *testing.T) {
	cluster1 := &api.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-1"},
		Spec: api.ClusterSpec{
			WorkloadSelector: &api.WorkloadSelector{
				Labels: map[string]string{
					"app": "envoy",
				},
			},
		},
	}
	cluster2 := &api.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-2"},
		Spec: api.ClusterSpec{
			WorkloadSelector: &api.WorkloadSelector{
				Labels: map[string]string{
					"app": "envoy",
				},
			},
		},
	}
	cluster3 := &api.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-2"},
		Spec: api.ClusterSpec{
			WorkloadSelector: &api.WorkloadSelector{
				Labels: map[string]string{
					"app":    "envoy",
					"number": "3",
				},
			},
		},
	}
	cluster4 := &api.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-2"},
		Spec: api.ClusterSpec{
			WorkloadSelector: &api.WorkloadSelector{
				Labels: map[string]string{
					"number": "4",
				},
			},
		},
	}
	cluster5 := &api.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-2"},
		Spec: api.ClusterSpec{
			WorkloadSelector: nil,
		},
	}

	tests := map[string]struct {
		clusters  []*api.Cluster
		podLabels map[string]string
		expected  []*api.Cluster
	}{
		"case 1": {
			clusters: []*api.Cluster{
				cluster1,
				cluster2,
				cluster3,
				cluster4,
			},
			podLabels: map[string]string{
				"app": "envoy",
			},
			expected: []*api.Cluster{
				cluster1,
				cluster2,
			},
		},
		"case 2": {
			clusters: []*api.Cluster{
				cluster1,
				cluster2,
				cluster3,
				cluster4,
			},
			podLabels: map[string]string{
				"app":    "envoy",
				"number": "3",
			},
			expected: []*api.Cluster{
				cluster1,
				cluster2,
				cluster3,
			},
		},
		"case 3": {
			clusters: []*api.Cluster{
				cluster1,
				cluster2,
				cluster3,
				cluster4,
			},
			podLabels: map[string]string{
				"number": "4",
			},
			expected: []*api.Cluster{
				cluster4,
			},
		},
		"case 4": {
			clusters: []*api.Cluster{
				cluster1,
				cluster2,
				cluster3,
				cluster4,
			},
			podLabels: map[string]string{},
			expected:  []*api.Cluster{},
		},
		"case 5": {
			clusters: []*api.Cluster{
				cluster1,
				cluster2,
				cluster3,
				cluster4,
				cluster5,
			},
			podLabels: map[string]string{},
			expected: []*api.Cluster{
				cluster5,
			},
		},
		"case 6": {
			clusters: []*api.Cluster{
				cluster1,
				cluster2,
				cluster3,
				cluster4,
				cluster5,
			},
			podLabels: map[string]string{
				"app": "envoy",
			},
			expected: []*api.Cluster{
				cluster1,
				cluster2,
				cluster5,
			},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := store.FilterClustersByLabels(test.clusters, test.podLabels)
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}
