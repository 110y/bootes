package controller_test

import (
	"testing"

	"github.com/110y/bootes/internal/k8s/internal/controller"
)

func TestToNodeName(t *testing.T) {
	actual := controller.ToNodeName("foo", "bar")
	expected := "foo.bar"
	if actual != expected {
		t.Errorf("want: %s, but got %s", expected, actual)
	}
}
