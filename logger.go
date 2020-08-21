package main

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/creativeprojects/clog"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	prefixMQTT = "MQTT "
	prefixApp  = "APP  "
)

var (
	// DebugLog is the default debug logger
	DebugLog *clog.StandardLogger
	// ErrorLog is the default error logger
	ErrorLog *clog.StandardLogger
)

func setupLogger(verbose, veryVerbose bool) {
	stderr := newLogWriter(os.Stderr)
	stdout := newLogWriter(os.Stdout)
	MQTT.CRITICAL = clog.NewStandardLogger(
		clog.LevelError,
		clog.NewStandardLogHandler(stderr, prefixMQTT, log.LstdFlags|log.Lmsgprefix),
	)
	MQTT.ERROR = MQTT.CRITICAL
	if verbose || veryVerbose {
		MQTT.WARN = clog.NewStandardLogger(
			clog.LevelWarning,
			clog.NewStandardLogHandler(stdout, prefixMQTT, log.LstdFlags|log.Lmsgprefix),
		)
	}
	if veryVerbose {
		MQTT.DEBUG = clog.NewStandardLogger(
			clog.LevelDebug,
			clog.NewStandardLogHandler(stdout, prefixMQTT, log.LstdFlags|log.Lmsgprefix),
		)
	}

	ErrorLog = clog.NewStandardLogger(
		clog.LevelError,
		clog.NewStandardLogHandler(stderr, prefixApp, log.LstdFlags|log.Lmsgprefix),
	)
	if verbose {
		DebugLog = clog.NewStandardLogger(
			clog.LevelDebug,
			clog.NewStandardLogHandler(stdout, prefixApp, log.LstdFlags|log.Lmsgprefix),
		)
	} else {
		DebugLog = clog.NewStandardLogger(clog.LevelDebug, &clog.DiscardHandler{})
	}
}

// logWriter is used to safe write to the logger
type logWriter struct {
	mutex  sync.Mutex
	writer io.Writer
}

func newLogWriter(writer io.Writer) *logWriter {
	return &logWriter{
		writer: writer,
	}
}

// Safe write to the underlying writer
func (h *logWriter) Write(p []byte) (int, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.writer.Write(p)
}

// Check interfaces
var (
	_ io.Writer = &logWriter{}
)
