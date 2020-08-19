package main

import (
	"io"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

// Configuration contains all information from the configuration file
type Configuration struct {
	MQTT        MQTTConfiguration         `yaml:"mqtt"`
	Zabbix      ZabbixConfiguration       `yaml:"zabbix"`
	Conversions []ConversionConfiguration `yaml:"conversions"`
}

// MQTTConfiguration contains MQTT server configuration
type MQTTConfiguration struct {
	ClientID  string   `yaml:"client-id"`
	ServerURL string   `yaml:"server"`
	Topics    []string `yaml:"topics"`
	QOS       int      `yaml:"qos"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

// ZabbixConfiguration contains zabbix server configuration
type ZabbixConfiguration struct {
	Server string `yaml:"server"`
}

// ConversionConfiguration contains a conversion entry
type ConversionConfiguration struct {
	Topic    string `yaml:"topic"`
	Hostname string `yaml:"hostname"`
	Key      string `yaml:"key"`
}

// newConfiguration creates an empty configuration object with a default configuration
func newConfiguration() *Configuration {
	config := &Configuration{}
	hostname, _ := os.Hostname()
	config.MQTT.ClientID = hostname + strconv.Itoa(time.Now().Second())
	config.MQTT.QOS = 1
	return config
}

// LoadFileConfiguration loads the configuration from the file
func LoadFileConfiguration(fileName string) (*Configuration, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return loadConfiguration(file)
}

// loadConfiguration from a io.ReadCloser
func loadConfiguration(reader io.ReadCloser) (*Configuration, error) {
	defer reader.Close()
	decoder := yaml.NewDecoder(reader)
	config := newConfiguration()
	err := decoder.Decode(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
