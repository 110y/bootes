package v1

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
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
	Config           *route.RouteConfiguration
}

func (c *Route) GetWorkloadSelector() *WorkloadSelector {
	return c.Spec.WorkloadSelector
}

func init() {
	SchemeBuilder.Register(&Route{}, &RouteList{})
}
