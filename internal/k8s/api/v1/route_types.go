package v1

import (
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

var _ EnvoyResource = (*Route)(nil)

// RouteList contains a list of Route
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*Route `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Route is the Schema for the routes API
// +k8s:openapi-gen=true
type Route struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RouteSpec
}

type RouteSpec struct {
	WorkloadSelector *WorkloadSelector `json:"workloadSelector,omitempty"`
	Config           *envoyapi.RouteConfiguration
}

func (c *Route) GetWorkloadSelector() *WorkloadSelector {
	return c.Spec.WorkloadSelector
}

func init() {
	SchemeBuilder.Register(&Route{}, &RouteList{})
}
