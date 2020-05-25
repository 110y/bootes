package validator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	api "github.com/110y/bootes/internal/k8s/api/v1"
)

var (
	_ admission.Handler         = (*RouteValidator)(nil)
	_ admission.DecoderInjector = (*RouteValidator)(nil)
)

type RouteValidator struct {
	decoder *admission.Decoder
	logger  logr.Logger
}

func NewRouteValidator(l logr.Logger) *RouteValidator {
	return &RouteValidator{
		logger: l,
	}
}

func (v *RouteValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionv1.Create, admissionv1.Update:
		route := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       api.RouteKind,
				"apiVersion": api.GroupVersion.String(),
			},
		}

		err := v.decoder.Decode(req, route)
		if err != nil {
			v.logger.Error(err, "failed to decode Route resource")
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("Failed to decode Route resource"))
		}

		_, err = api.UnmarshalRouteObject(route.Object)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		return admission.Allowed("")
	default: // Delete, Connect
		return admission.Allowed("")
	}
}

func (v *RouteValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
