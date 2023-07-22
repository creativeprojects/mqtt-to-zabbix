package main

import (
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func notifyReady() {
	_, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	if err != nil {
		ErrorLog.Printf("cannot notify systemd: %s", err)
	}
}

func notifyLeaving() {
	_, _ = daemon.SdNotify(false, daemon.SdNotifyStopping)
}

func setupWatchdog(client MQTT.Client) {
	interval, err := daemon.SdWatchdogEnabled(false)
	if err != nil {
		ErrorLog.Printf("cannot verify if systemd watchdog is enabled: %s", err)
		return
	}
	if interval == 0 {
		// watchdog not enabled
		return
	}
	for {
		if client.IsConnected() {
			// Everything is running
			_, err := daemon.SdNotify(false, daemon.SdNotifyWatchdog)
			if err != nil {
				ErrorLog.Printf("cannot notify systemd watchdog: %s", err)
			}
		}
		time.Sleep(interval / 3)
	}
}
