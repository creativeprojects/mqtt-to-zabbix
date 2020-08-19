package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	prefixCriticalMQTT = "MQTT CRITICAL"
	prefixErrorMQTT    = "MQTT ERROR   "
	prefixWarningMQTT  = "MQTT WARNING "
	prefixDebugMQTT    = "MQTT DEBUG   "
	prefixErrorApp     = "ERROR        "
	prefixDebugApp     = "DEBUG        "
)

var (
	// DebugLog is the default debug logger
	DebugLog *log.Logger
	// ErrorLog is the default error logger
	ErrorLog *log.Logger
)

func setupLogger(verbose, veryVerbose bool) {
	stderr := newLogHandler(os.Stderr)
	stdout := newLogHandler(os.Stdout)
	MQTT.CRITICAL = log.New(stderr, prefixCriticalMQTT, log.LstdFlags|log.Lmsgprefix)
	MQTT.ERROR = log.New(stderr, prefixErrorMQTT, log.LstdFlags|log.Lmsgprefix)
	if verbose {
		MQTT.WARN = log.New(stdout, prefixWarningMQTT, log.LstdFlags|log.Lmsgprefix)
	}
	if veryVerbose {
		MQTT.DEBUG = log.New(stdout, prefixDebugMQTT, log.LstdFlags|log.Lmsgprefix)
	}

	ErrorLog = log.New(stderr, prefixErrorApp, log.LstdFlags|log.Lmsgprefix)
	if verbose {
		DebugLog = log.New(stdout, prefixDebugApp, log.LstdFlags|log.Lmsgprefix)
	} else {
		// TODO: this is a waste of resources + a mutex lock/unlock on each debug entry
		DebugLog = log.New(ioutil.Discard, "", 0)
	}
}

// logHandler is used to safe write to the logger
type logHandler struct {
	mutex  sync.Mutex
	writer io.Writer
}

func newLogHandler(writer io.Writer) *logHandler {
	return &logHandler{
		writer: writer,
	}
}

// Safe write to the underlying writer
func (h *logHandler) Write(p []byte) (int, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.writer.Write(p)
}

// Check interfaces
var (
	_ io.Writer = &logHandler{}
)
