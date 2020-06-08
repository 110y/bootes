package v1

import (
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

var _ EnvoyResource = (*Listener)(nil)

// ListenerList contains a list of Listener
type ListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*Listener `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Listener is the Schema for the listeners API
// +k8s:openapi-gen=true
type Listener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ListenerSpec
}

type ListenerSpec struct {
	WorkloadSelector *WorkloadSelector `json:"workloadSelector,omitempty"`
	Config           *listener.Listener
}

func (l *Listener) GetWorkloadSelector() *WorkloadSelector {
	return l.Spec.WorkloadSelector
}

func init() {
	SchemeBuilder.Register(&Listener{}, &ListenerList{})
}
