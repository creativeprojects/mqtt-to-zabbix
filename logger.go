package main

import (
	"log"

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
	stderr := clog.Stderr()
	stdout := clog.Stdout()
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
