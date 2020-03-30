package v1

type EnvoyResource interface {
	GetWorkloadSelector() *WorkloadSelector
}
