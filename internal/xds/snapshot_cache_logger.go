package xds

import (
	"fmt"

	"github.com/go-logr/logr"
)

type snapshotCacheLogger struct {
	logger logr.Logger
}

func (l *snapshotCacheLogger) Debugf(format string, args ...interface{}) {
	l.info("debug", format, args...)
}

func (l *snapshotCacheLogger) Infof(format string, args ...interface{}) {
	l.info("info", format, args...)
}

func (l *snapshotCacheLogger) Warnf(format string, args ...interface{}) {
	l.info("warn", format, args...)
}

func (l *snapshotCacheLogger) Errorf(format string, args ...interface{}) {
	l.info("error", format, args...)
}

func (l *snapshotCacheLogger) info(level, format string, args ...interface{}) {
	l.logger.WithName(level).Info(fmt.Sprintf(format, args...))
}
