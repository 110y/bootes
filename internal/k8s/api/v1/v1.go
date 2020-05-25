package v1

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
)

var (
	errWorkloadSelectorNotFound = errors.New("workloadSelector not found")
	unmarshaler                 *protojson.UnmarshalOptions
)

func extractSpecFromObject(object map[string]interface{}) (map[string]interface{}, error) {
	spec, ok := object["spec"]
	if !ok {
		return nil, fmt.Errorf("spec not found")
	}

	s, ok := spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec form")
	}

	return s, nil
}

func unmarshalEnvoyConfig(spec map[string]interface{}) ([]byte, error) {
	config, ok := spec["config"]
	if !ok {
		return nil, fmt.Errorf("spec.config not found")
	}

	j, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec.config: %w", err)
	}

	return j, nil
}

func unmarshalWorkloadSelector(spec map[string]interface{}) (*WorkloadSelector, error) {
	selector, ok := spec["workloadSelector"]
	if !ok {
		return nil, errWorkloadSelectorNotFound
	}

	j, err := json.Marshal(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec.workloadSelector: %w", err)
	}

	var ws WorkloadSelector
	if err := json.Unmarshal(j, &ws); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec.workloadSelector: %w", err)
	}

	return &ws, nil
}

func init() {
	unmarshaler = &protojson.UnmarshalOptions{
		AllowPartial:   false,
		DiscardUnknown: false,
	}
}
