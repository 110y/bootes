package k8s

import (
	"github.com/go-logr/logr"
)

type ManagerConfig struct {
	HealthzServerPort       int
	WebhookServerPort       int
	MetricsServerPort       int
	EnableValidatingWebhook bool
	Logger                  logr.Logger
}
