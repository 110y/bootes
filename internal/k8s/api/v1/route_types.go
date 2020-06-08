package v1

import (
	"errors"
	"fmt"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/golang/protobuf/proto"
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

func (r *Route) GetWorkloadSelector() *WorkloadSelector {
	return r.Spec.WorkloadSelector
}

func UnmarshalRouteObject(object map[string]interface{}) (*Route, error) {
	spec, err := extractSpecFromObject(object)
	if err != nil {
		return nil, err
	}

	config, err := unmarshalRouteConfig(spec)
	if err != nil {
		return nil, err
	}

	selector, err := unmarshalWorkloadSelector(spec)
	if err != nil && !errors.Is(err, errWorkloadSelectorNotFound) {
		return nil, err
	}

	return &Route{
		Spec: RouteSpec{
			WorkloadSelector: selector,
			Config:           config,
		},
	}, nil
}

func unmarshalRouteConfig(spec map[string]interface{}) (*route.RouteConfiguration, error) {
	config, err := unmarshalEnvoyConfig(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal envoy configuration: %w", err)
	}

	route := &route.RouteConfiguration{}
	if err := unmarshaler.Unmarshal(config, proto.MessageV2(route)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec.config: %w", err)
	}

	return route, nil
}

func init() {
	SchemeBuilder.Register(&Route{}, &RouteList{})
}
