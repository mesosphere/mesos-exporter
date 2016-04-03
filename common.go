package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metricMap map[string]float64

var (
	notFoundInMap = errors.New("Couldn't find key in map")
)

func gauge(subsystem, name, help string, labels ...string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "mesos",
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, labels)
}

func counter(subsystem, name, help string, labels ...string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "mesos",
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, labels)
}

type metricCollector struct {
	*http.Client
	url     string
	metrics map[prometheus.Collector]func(metricMap, prometheus.Collector) error
}

func newMetricCollector(url string, timeout time.Duration, metrics map[prometheus.Collector]func(metricMap, prometheus.Collector) error) *metricCollector {
	return &metricCollector{
		url:     url,
		Client:  &http.Client{Timeout: timeout},
		metrics: metrics,
	}
}

func findLeader(url string) (string, error) {
	if !strings.ContainsAny(url, ",") {
		return url, nil
	}
	for _, testUrl := range strings.Split(url, ",") {
		resp, err := http.Get(testUrl + "/state.json")
		if err != nil {
			log.Print(err)
			continue
		}
		defer resp.Body.Close()
		var s state
		if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
			log.Print(err)
			continue
		}
		if s.Leader == "master@" + testUrl[7:len(testUrl)] {
			return testUrl, nil
		}
	}
	return "", errors.New("Unable to find leader")
}

func (c *metricCollector) Collect(ch chan<- prometheus.Metric) {
	leaderUrl, err := findLeader(c.url)
	if err != nil {
		log.Print(err)
		errorCounter.Inc()
		return
	}
	res, err := c.Get(leaderUrl + "/metrics/snapshot")
	if err != nil {
		log.Print(err)
		errorCounter.Inc()
		return
	}
	defer res.Body.Close()

	var m metricMap
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		log.Print(err)
		errorCounter.Inc()
		return
	}

	for cm, f := range c.metrics {
		if err := f(m, cm); err != nil {
			if err == notFoundInMap {
				ch := make(chan *prometheus.Desc, 1)
				cm.Describe(ch)
				log.Printf("Couldn't find fields required to update %s\n", <-ch)
			} else {
				log.Println(err)
			}
			errorCounter.Inc()
			continue
		}
		cm.Collect(ch)
	}
}

func (c *metricCollector) Describe(ch chan<- *prometheus.Desc) {
	for m, _ := range c.metrics {
		m.Describe(ch)
	}
}
