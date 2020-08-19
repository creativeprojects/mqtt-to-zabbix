package main

import (
	"crypto/tls"
	"flag"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/blacked/go-zabbix"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	configFile  string
	verbose     bool
	veryVerbose bool
	config      *Configuration
)

func main() {
	var err error
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	flag.StringVar(&configFile, "config", "config.yaml", "configuration file")
	flag.BoolVar(&verbose, "v", false, "display debugging information")
	flag.BoolVar(&veryVerbose, "vv", false, "display even more debugging information")
	flag.Parse()

	setupLogger(verbose, veryVerbose)

	config, err = LoadFileConfiguration(configFile)
	if err != nil {
		ErrorLog.Printf("configuration file '%s' not found: %s", configFile, err)
		return
	}

	if len(config.MQTT.Topics) == 0 {
		ErrorLog.Print("no topic defined in the configuration")
		return
	}

	if !strings.Contains(config.Zabbix.Server, ":/") {
		config.Zabbix.Server = "tcp://" + config.Zabbix.Server
	}
	zabbixURL, err := url.Parse(config.Zabbix.Server)
	if err != nil {
		ErrorLog.Printf("zabbix server url not valid '%s': %s", config.Zabbix.Server, err)
		return
	}
	DebugLog.Printf("%+v", zabbixURL)
	zabbixHost := zabbixURL.Hostname()
	zabbixPort, err := strconv.Atoi(zabbixURL.Port())
	if err != nil || zabbixPort == 0 {
		zabbixPort = 10051
	}
	zabbixSender = zabbix.NewSender(zabbixHost, zabbixPort)
	DebugLog.Printf("Zabbix server: %s port %d", zabbixHost, zabbixPort)

	connOpts := MQTT.NewClientOptions().AddBroker(config.MQTT.ServerURL).SetClientID(config.MQTT.ClientID).SetCleanSession(true)
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	connOpts.SetTLSConfig(tlsConfig)

	connOpts.OnConnect = func(c MQTT.Client) {
		for _, topic := range config.MQTT.Topics {
			if token := c.Subscribe(topic, byte(config.MQTT.QOS), onMessageReceived); token.Wait() && token.Error() != nil {
				ErrorLog.Printf("error subscribing to topic '%s': %v", topic, token.Error())
			}
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		ErrorLog.Printf("error connecting to MQTT server %s: %v", config.MQTT.ServerURL, token.Error())
		return
	}

	DebugLog.Printf("connected to MQTT server %s", config.MQTT.ServerURL)

	<-c
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	DebugLog.Printf("received: '%s' = '%s'", message.Topic(), message.Payload())
	transferMessage(message.Topic(), message.Payload())
}
