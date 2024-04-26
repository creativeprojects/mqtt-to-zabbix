package main

import (
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/datadope-io/go-zabbix/v2"
)

var (
	zabbixSender       *zabbix.Sender
	zabbixMessageQueue *messageQueue
	debounced          func(func())
)

type messageQueue struct {
	mu    sync.Mutex
	queue []*zabbix.Metric
}

func (q *messageQueue) add(metric *zabbix.Metric) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, metric)
}

func (q *messageQueue) get() []*zabbix.Metric {
	q.mu.Lock()
	defer func() {
		q.queue = q.queue[:0]
		q.mu.Unlock()
	}()
	return q.queue
}

func init() {
	zabbixMessageQueue = &messageQueue{
		queue: make([]*zabbix.Metric, 0, 10),
	}
	debounced = debounce.New(200 * time.Millisecond)
}

func transferMessage(topic string, message []byte) {
	for _, conversion := range config.Conversions {
		if conversion.Topic == topic {
			zabbixMetric(conversion.Hostname, conversion.Key, string(message))
		}
	}
}

func zabbixMetric(hostname, key, value string) {
	zabbixMessageQueue.add(zabbix.NewMetric(hostname, key, value, false, time.Now().Unix()))
	debounced(zabbixSend)
}

func zabbixSend() {
	packet := zabbix.NewPacket(zabbixMessageQueue.get(), false)

	resp, err := zabbixSender.Send(packet)
	if err != nil {
		ErrorLog.Printf("zabbix sender: %s", err)
	}
	DebugLog.Printf("zabbix response: %s", resp)
}
