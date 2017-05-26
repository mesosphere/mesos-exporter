package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type (
	resources struct {
		CPUs  float64 `json:"cpus"`
		Disk  float64 `json:"disk"`
		Mem   float64 `json:"mem"`
		Ports ranges  `json:"ports"`
	}

	task struct {
		Name        string    `json:"name"`
		ID          string    `json:"id"`
		ExecutorID  string    `json:"executor_id"`
		FrameworkID string    `json:"framework_id"`
		SlaveID     string    `json:"slave_id"`
		State       string    `json:"state"`
		Labels      []label   `json:"labels"`
		Resources   resources `json:"resources"`
		Statuses    []status  `json:"statuses"`
	}

	label struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	status struct {
		State     string  `json:"state"`
		Timestamp float64 `json:"timestamp"`
	}

	tokenResponse struct {
		Token string `json:"token"`
	}

	tokenRequest struct {
		UID string `json:"uid"`
		Token string `json:"token"`
	}

	mesosSecret struct {
		LoginEndpoint string `json:"login_endpoint"`
		PrivateKey string `json:"private_key"`
		Scheme string `json:"scheme"`
		UID string `json:"uid"`
	}
)

type metricMap map[string]float64

var (
	notFoundInMap = errors.New("Couldn't find key in map")
)

type settableCounterVec struct {
	desc   *prometheus.Desc
	values []prometheus.Metric
}

func (c *settableCounterVec) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *settableCounterVec) Collect(ch chan<- prometheus.Metric) {
	for _, v := range c.values {
		ch <- v
	}

	c.values = nil
}

func (c *settableCounterVec) Set(value float64, labelValues ...string) {
	c.values = append(c.values, prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value, labelValues...))
}

type settableCounter struct {
	desc  *prometheus.Desc
	value prometheus.Metric
}

func (c *settableCounter) Describe(ch chan<- *prometheus.Desc) {
	if c.desc == nil {
		log.Printf("NIL description: %v", c)
	}
	ch <- c.desc
}

func (c *settableCounter) Collect(ch chan<- prometheus.Metric) {
	if c.value == nil {
		log.Printf("NIL value: %v", c)
	}
	ch <- c.value
}

func (c *settableCounter) Set(value float64) {
	c.value = prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value)
}

func newSettableCounter(subsystem, name, help string) *settableCounter {
	return &settableCounter{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("mesos", subsystem, name),
			help,
			nil,
			prometheus.Labels{},
		),
	}
}

func gauge(subsystem, name, help string, labels ...string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "mesos",
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, labels)
}

func counter(subsystem, name, help string, labels ...string) *settableCounterVec {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName("mesos", subsystem, name),
		help,
		labels,
		prometheus.Labels{},
	)

	return &settableCounterVec{
		desc:   desc,
		values: nil,
	}
}

type authInfo struct {
	username string
	password string
	loginUrl string
	token string
	strictMode bool
	privateKey string
	skipSSLVerify bool
}

type httpClient struct {
	http.Client
	url  string
	auth authInfo
}

type metricCollector struct {
	*httpClient
	metrics map[prometheus.Collector]func(metricMap, prometheus.Collector) error
}

func newMetricCollector(httpClient *httpClient, metrics map[prometheus.Collector]func(metricMap, prometheus.Collector) error) prometheus.Collector {
	return &metricCollector{httpClient, metrics}
}

func (httpClient *httpClient) fetchAndDecode(endpoint string, target interface{}) bool {
	url := strings.TrimSuffix(httpClient.url, "/") + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating HTTP request to %s: %s", url, err)
		return false
	}
	if httpClient.auth.username != "" && httpClient.auth.password != "" {
		req.SetBasicAuth(httpClient.auth.username, httpClient.auth.password)
	}
	if httpClient.auth.strictMode {
		req.Header.Add("Authorization", httpClient.auth.token)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Error fetching %s: %s", url, err)
		errorCounter.Inc()
		return false
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&target); err != nil {
		log.Printf("Error decoding response body from %s: %s", url, err)
		errorCounter.Inc()
		return false
	}

	return true
}

func (c *metricCollector) Collect(ch chan<- prometheus.Metric) {
	var m metricMap
	c.fetchAndDecode("/metrics/snapshot", &m)
	for cm, f := range c.metrics {
		if err := f(m, cm); err != nil {
			if err == notFoundInMap {
				ch := make(chan *prometheus.Desc, 1)
				cm.Describe(ch)
				log.Printf("Couldn't find fields required to update %s\n", <-ch)
			} else {
				log.Printf("Error extracting metric: %s", err)
			}
			errorCounter.Inc()
			continue
		}
		cm.Collect(ch)
	}
}

func (c *metricCollector) Describe(ch chan<- *prometheus.Desc) {
	for m := range c.metrics {
		m.Describe(ch)
	}
}

var invalidLabelNameCharRE = regexp.MustCompile("(^[^a-zA-Z_])|([^a-zA-Z0-9_])")

// Sanitize label names according to https://prometheus.io/docs/concepts/data_model/
func normaliseLabel(label string) string {
	return invalidLabelNameCharRE.ReplaceAllString(label, "_")
}

func normaliseLabelList(labelList []string) []string {
	normalisedLabelList := []string{}
	for _, label := range labelList {
		normalisedLabelList = append(normalisedLabelList, normaliseLabel(label))
	}
	return normalisedLabelList
}

func stringInSlice(string string, slice []string) bool {
	for _, elem := range slice {
		if string == elem {
			return true
		}
	}
	return false
}

func getLabelValuesFromMap(labels prometheus.Labels, orderedLabelKeys []string) []string {
	labelValues := []string{}
	for _, label := range orderedLabelKeys {
		labelValues = append(labelValues, labels[label])
	}
	return labelValues
}
