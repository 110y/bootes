package v1

import (
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

var _ EnvoyResource = (*Endpoint)(nil)

// EndpointList contains a list of Endpoint
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*Endpoint `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Endpoint is the Schema for the endpoints API
// +k8s:openapi-gen=true
type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EndpointSpec
}

type EndpointSpec struct {
	WorkloadSelector *WorkloadSelector `json:"workloadSelector,omitempty"`
	Config           *endpoint.ClusterLoadAssignment
}

func (l *Endpoint) GetWorkloadSelector() *WorkloadSelector {
	return l.Spec.WorkloadSelector
}

func init() {
	SchemeBuilder.Register(&Endpoint{}, &EndpointList{})
}
