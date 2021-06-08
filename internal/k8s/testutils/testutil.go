// +build test

package testutils

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/110y/bootes/internal/k8s"
)

var mu sync.Once

var (
	CmpOptProtoTransformer = cmp.FilterValues(func(x, y interface{}) bool {
		if _, ok := x.(proto.Message); ok {
			if _, ok := y.(proto.Message); ok {
				return true
			}
		}
		return false
	}, protocmp.Transform())

	// TODO: these comparers should be more restrict.
	CmpOptPodListComparer = cmp.Comparer(func(x, y corev1.PodList) bool {
		if len(x.Items) != len(y.Items) {
			return false
		}

		for i, xp := range x.Items {
			yp := y.Items[i]

			if !cmp.Equal(xp, yp, CmpOptPodComparer) {
				return false
			}
		}

		return true
	})

	CmpOptPodComparer = cmp.Comparer(func(x, y corev1.Pod) bool {
		if len(x.Spec.Containers) != len(y.Spec.Containers) {
			return false
		}

		for i, xc := range x.Spec.Containers {
			yc := y.Spec.Containers[i]

			if !cmp.Equal(xc, yc, CmpOptContainerComparer) {
				return false
			}
		}

		return x.Name == y.Name &&
			x.Namespace == y.Namespace
	})

	CmpOptContainerComparer = cmp.Comparer(func(x, y corev1.Container) bool {
		if x.Name != y.Name {
			return false
		}

		if x.Image != y.Image {
			return false
		}

		return true
	})

	s = k8sRuntime.NewScheme()
)

func TestK8SClient() (client.Client, func(), error) {
	mu.Do(func() {
		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:       false,
			Encoding:          "json",
			DisableCaller:     true,
			DisableStacktrace: true,
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "timestamp",
				LevelKey:       "level",
				NameKey:        "logger",
				MessageKey:     "message",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.EpochMillisTimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
			},
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		l, err := config.Build()
		if err != nil {
			panic(err)
		}

		ctrl.SetLogger(zapr.NewLogger(l).WithName("test"))
	})

	if err := scheme.AddToScheme(s); err != nil {
		return nil, nil, fmt.Errorf("failed to create new scheme: %w", err)
	}

	if err := k8s.SchemeBuilder.AddToScheme(s); err != nil {
		return nil, nil, fmt.Errorf("faileld to add bootes scheme: %w", err)
	}

	_, file, _, _ := runtime.Caller(0)
	testEnv := envtest.Environment{
		BinaryAssetsDirectory:    filepath.Join(path.Dir(file), "..", "..", "..", "dev", "bin"),
		CRDDirectoryPaths:        []string{filepath.Join(path.Dir(file), "..", "..", "..", "kubernetes", "kpt", "crd")},
		ControlPlaneStartTimeout: 20 * time.Second,
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: false,
		KubeAPIServerFlags: []string{
			"--advertise-address=127.0.0.1",
			"--etcd-servers={{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
			"--cert-dir={{ .CertDir }}",
			"--insecure-port={{ if .URL }}{{ .URL.Port }}{{ end }}",
			"--insecure-bind-address={{ if .URL }}{{ .URL.Hostname }}{{ end }}",
			"--secure-port={{ if .SecurePort }}{{ .SecurePort }}{{ end }}",
			// we're keeping this disabled because if enabled, default SA is missing which would force all tests to create one
			// in normal apiserver operation this SA is created by controller, but that is not run in integration environment
			"--disable-admission-plugins=ServiceAccount",
			"--service-cluster-ip-range=10.0.0.0/24",
			"--allow-privileged=true",
		},
	}

	cfg, err := testEnv.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start test env: %w", err)
	}

	cli, err := client.New(cfg, client.Options{
		Scheme: s,
	})
	if err != nil {
		err = fmt.Errorf("faileld to create controller-runtime client: %w", err)

		if nerr := testEnv.Stop(); err != nil {
			err = fmt.Errorf("failed to stop test env: %w", nerr)
		}

		return nil, nil, err
	}

	return cli, func() {
		if err := testEnv.Stop(); err != nil {
			panic(fmt.Sprintf("failed to stop envtest instance: %s", err))
		}
	}, nil
}

func NewNamespace(t *testing.T, ctx context.Context, cli client.Client) string {
	t.Helper()

	name := uuid.New().String()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := cli.Create(ctx, ns); err != nil {
		t.Fatalf("failed to create namespace: %s", err)
	}

	return name
}
