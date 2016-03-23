package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type (
	executor struct {
		ID          string      `json:"executor_id"`
		Name        string      `json:"executor_name"`
		FrameworkID string      `json:"framework_id"`
		Source      string      `json:"source"`
		Statistics  *statistics `json:"statistics"`
	}

	statistics struct {
		CpusLimit             float64 `json:"cpus_limit"`
		CpusSystemTimeSecs    float64 `json:"cpus_system_time_secs"`
		CpusUserTimeSecs      float64 `json:"cpus_user_time_secs"`
		CpusThrottledTimeSecs float64 `json:"cpus_throttled_time_secs"`

		MemLimitBytes float64 `json:"mem_limit_bytes"`
		MemRssBytes   float64 `json:"mem_rss_bytes"`

		NetRxBytes   float64 `json:"net_rx_bytes"`
		NetRxDropped float64 `json:"net_rx_dropped"`
		NetRxErrors  float64 `json:"net_rx_errors"`
		NetRxPackets float64 `json:"net_rx_packets"`
		NetTxBytes   float64 `json:"net_tx_bytes"`
		NetTxDropped float64 `json:"net_tx_dropped"`
		NetTxErrors  float64 `json:"net_tx_errors"`
		NetTxPackets float64 `json:"net_tx_packets"`
	}

	slaveCollector struct {
		*http.Client
		url     string
		metrics map[*prometheus.Desc]metric
	}

	metric struct {
		valueType prometheus.ValueType
		get       func(*statistics) float64
	}
)

func newSlaveMonitorCollector(url string, timeout time.Duration) *slaveCollector {
	labels := []string{"id", "framework_id", "source"}

	return &slaveCollector{
		Client: &http.Client{Timeout: timeout},
		url:    url,
		metrics: map[*prometheus.Desc]metric{
			// CPU
			prometheus.NewDesc(
				"cpus_limit",
				"Current limit of CPUs for task",
				labels, nil,
			): metric{prometheus.GaugeValue, func(s *statistics) float64 { return s.CpusLimit }},
			prometheus.NewDesc(
				"cpu_system_seconds_total",
				"Total system CPU seconds",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.CpusSystemTimeSecs }},
			prometheus.NewDesc(
				"cpu_user_seconds_total",
				"Total user CPU seconds",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.CpusUserTimeSecs }},
			prometheus.NewDesc(
				"cpu_throttled_seconds_total",
				"Total time CPU was throttled",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.CpusThrottledTimeSecs }},

			// Memory
			prometheus.NewDesc(
				"mem_limit_bytes",
				"Current memory limit in bytes",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.MemLimitBytes }},
			prometheus.NewDesc(
				"mem_rss_bytes",
				"Current rss memory usage",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.MemRssBytes }},

			// Network
			// - RX
			prometheus.NewDesc(
				"network_receive_bytes_total",
				"Total bytes received",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetRxBytes }},
			prometheus.NewDesc(
				"network_receive_dropped_total",
				"Total packets dropped while receiving",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetRxDropped }},
			prometheus.NewDesc(
				"network_receive_errors_total",
				"Total errors while receiving",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetRxBytes }},
			prometheus.NewDesc(
				"network_receive_packets_total",
				"Total packets received",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetRxBytes }},
			// - TX
			prometheus.NewDesc(
				"network_transmit_bytes_total",
				"Total bytes transmitted",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetTxBytes }},
			prometheus.NewDesc(
				"network_transmit_dropped_total",
				"Total packets dropped while transmitting",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetTxDropped }},
			prometheus.NewDesc(
				"network_transmit_errors_total",
				"Total errors while transmitting",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetTxBytes }},
			prometheus.NewDesc(
				"network_transmit_packets_total",
				"Total packets transmitted",
				labels, nil,
			): metric{prometheus.CounterValue, func(s *statistics) float64 { return s.NetTxBytes }},
		},
	}
}

func (c *slaveCollector) Collect(ch chan<- prometheus.Metric) {
	res, err := http.Get(c.url + "/monitor/statistics")
	if err != nil {
		log.Print(err)
		return
	}
	defer res.Body.Close()

	stats := []executor{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		log.Print(err)
		return
	}

	for _, exec := range stats {
		for desc, m := range c.metrics {
			ch <- prometheus.MustNewConstMetric(desc, m.valueType, m.get(exec.Statistics), exec.ID, exec.FrameworkID, exec.Source)
		}
	}
}

func (c *slaveCollector) Describe(ch chan<- *prometheus.Desc) {
	for metric := range c.metrics {
		ch <- metric
	}
}
