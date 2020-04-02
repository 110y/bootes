package store_test

import (
	"testing"

	"github.com/110y/bootes/internal/k8s/store"
)

func TestToNodeName(t *testing.T) {
	actual := store.ToNodeName("foo", "bar")
	expected := "foo.bar"
	if actual != expected {
		t.Errorf("want: %s, but got %s", expected, actual)
	}
}

func TestToNamespacedName(t *testing.T) {
	tests := map[string]struct {
		in                string
		expectedName      string
		expectedNamespace string
	}{
		"default namespace": {
			in:                "foo",
			expectedName:      "foo",
			expectedNamespace: "",
		},
		"namespaced": {
			in:                "foo.bar",
			expectedName:      "foo",
			expectedNamespace: "bar",
		},
		"have multiple dots": {
			in:                "foo.bar.baz",
			expectedName:      "foo",
			expectedNamespace: "bar.baz",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			name, namespace := store.ToNamespacedName(test.in)

			if test.expectedName != name {
				t.Errorf("name, want: %s, but got %s", test.expectedName, name)
			}

			if test.expectedNamespace != namespace {
				t.Errorf("namespace, want: %s, but got %s", test.expectedNamespace, namespace)
			}
		})
	}
}
