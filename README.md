Very simple implementation of a service subscribing to some MQTT topics, and sending the selected values to a Zabbix trapper

Everything is in the configuration file:

```yaml
---
mqtt:
    client-id: mqtt-to-zabbix
    server: tcp://mqtt-server:1883
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

If launched as a systemd service, it sends a watchdog message to systemd every 5 minutes
