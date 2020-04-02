package xds

import (
	"fmt"

	"github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/go-logr/logr"
)

var _ log.Logger = (*snapshotCacheLogger)(nil)

func newSnapshotCacheLogger(l logr.Logger) log.Logger {
	return &snapshotCacheLogger{
		debugf: l.WithName("debug"),
		infof:  l.WithName("info"),
		warnf:  l.WithName("warn"),
		errorf: l.WithName("error"),
	}
}

type snapshotCacheLogger struct {
	debugf logr.Logger
	infof  logr.Logger
	warnf  logr.Logger
	errorf logr.Logger
}

func (l *snapshotCacheLogger) Debugf(format string, args ...interface{}) {
	l.debugf.Info(l.message(format, args...))
}

func (l *snapshotCacheLogger) Infof(format string, args ...interface{}) {
	l.debugf.Info(l.message(format, args...))
}

func (l *snapshotCacheLogger) Warnf(format string, args ...interface{}) {
	l.warnf.Info(l.message(format, args...))
}

func (l *snapshotCacheLogger) Errorf(format string, args ...interface{}) {
	l.errorf.Info(l.message(format, args...))
}

func (l *snapshotCacheLogger) message(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
