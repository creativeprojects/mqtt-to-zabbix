## Introduction

Very simple implementation of a service subscribing to some MQTT topics, and sending the selected values to a Zabbix trapper.

The service supports TLS with a dedicated CA certificate for MQTT.

## Configuration

Everything is in the configuration file:

```yaml
---
mqtt:
    client-id: mqtt-to-zabbix
    # server: tcp://mqtt-server:1883
    server: ssl://mqtt-server:8883
    ca: ca_root.pem
    topics:
      - "homie/#"

zabbix:
    server: tcp://zabbix-server:10051

conversions:
  -
    topic: homie/pizero/rpi/temperature
    hostname: pizero
    key: mqtt.pizero.rpi.temperature
  -
    topic: homie/pizero/bmp280/temperature
    hostname: pizero
    key: mqtt.pizero.bmp280.temperature
  -
    topic: homie/pizero/bmp280/pressure
    hostname: pizero
    key: mqtt.pizero.bmp280.pressure

```

## Systemd

If launched as a systemd service, it sends a watchdog message to systemd every 5 minutes

Here's an example of a systemd unit file with the watchdog:

```
[Unit]
Description=MQTT to Zabbix
After=network.target

[Service]
Type=notify
WorkingDirectory=/opt/mqtt-to-zabbix/
ExecStart=/opt/mqtt-to-zabbix/mqtt-to-zabbix
WatchdogSec=900s
Restart=on-failure

[Install]
WantedBy=multi-user.target
```