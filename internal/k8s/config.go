package k8s

import (
	"github.com/go-logr/logr"
)

type ManagerConfig struct {
	HealthzServerPort int
	WebhookServerPort int
	MetricsServerPort int
	Logger            logr.Logger
}
