package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/datadope-io/go-zabbix/v2"
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
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)

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
	zabbixHost := zabbixURL.Hostname()
	zabbixPort, err := strconv.Atoi(zabbixURL.Port())
	if err != nil || zabbixPort == 0 {
		zabbixPort = 10051
	}
	zabbixServer := fmt.Sprintf("%s:%d", zabbixHost, zabbixPort)
	zabbixSender = zabbix.NewSender(zabbixServer)
	DebugLog.Printf("zabbix server: %s", zabbixServer)

	connOpts := MQTT.NewClientOptions().AddBroker(config.MQTT.ServerURL).SetClientID(config.MQTT.ClientID)
	if config.MQTT.Username != "" {
		connOpts.SetUsername(config.MQTT.Username)
	}
	if config.MQTT.Password != "" {
		connOpts.SetPassword(config.MQTT.Password)
	}
	if config.MQTT.CA != "" {
		tlsConfig := getTLSConfig(config)
		if tlsConfig != nil {
			connOpts.SetTLSConfig(tlsConfig)
		}
	}

	connOpts.OnConnect = func(c MQTT.Client) {
		DebugLog.Printf("subscribing to %d topics", len(config.MQTT.Topics))
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

	DebugLog.Printf("connected to MQTT server: %s, client ID: %s", config.MQTT.ServerURL, config.MQTT.ClientID)

	// Notify systemd we're ready to serve
	notifyReady()

	// systemd watchdog
	go setupWatchdog(client)

	// Wait until we're politely asked to leave
	<-stop

	client.Disconnect(2000)

	// Notify systemd we're leaving
	notifyLeaving()
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	DebugLog.Printf("received: '%s' = '%s'", message.Topic(), message.Payload())
	transferMessage(message.Topic(), message.Payload())
}

func getTLSConfig(config *Configuration) *tls.Config {
	if config.MQTT.CA == "" {
		return nil
	}
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		caCertPool = x509.NewCertPool()
	}
	caCert, err := os.ReadFile(config.MQTT.CA)
	if err != nil {
		ErrorLog.Printf("cannot load CA certificate: %s", err)
		return nil
	}
	if !caCertPool.AppendCertsFromPEM(caCert) {
		ErrorLog.Printf("invalid certificate: %q", config.MQTT.CA)
		return nil
	}
	return &tls.Config{
		RootCAs: caCertPool,
	}
}
