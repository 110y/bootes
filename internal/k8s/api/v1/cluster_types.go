package v1

import (
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*Cluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster is the Schema for the clusters API
// +k8s:openapi-gen=true
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec
}

type ClusterSpec struct {
	WorkloadSelector *WorkloadSelector `json:"workloadSelector,omitempty"`
	Config           *envoyapi.Cluster
}

type WorkloadSelector struct {
	Labels map[string]string `json:"labels"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
