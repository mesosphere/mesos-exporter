// Scrape the /slave(1)/state endpoint to get information on the tasks running
// on executors. Information scraped at this point:
//
// * Labels of running tasks ("mesos_slave_task_labels" series)
// * Attributes of mesos slaves ("mesos_slave_attributes")
package main

import (
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type (
	slaveState struct {
		Attributes map[string]json.RawMessage `json:"attributes"`
		Frameworks []slaveFramework           `json:"frameworks"`
	}
	slaveFramework struct {
		ID        string               `json:"ID"`
		Executors []slaveStateExecutor `json:"executors"`
	}
	slaveStateExecutor struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Source string `json:"source"`
		Tasks  []task `json:"tasks"`
	}

	slaveStateCollector struct {
		*httpClient
		metrics map[*prometheus.Desc]slaveMetric
	}
	slaveMetric struct {
		valueType prometheus.ValueType
		value     func(*slaveState) []metricValue
	}
	metricValue struct {
		result float64
		labels []string
	}
)

func newSlaveStateCollector(httpClient *httpClient, userTaskLabelList []string, slaveAttributeLabelList []string) *slaveStateCollector {
	c := slaveStateCollector{httpClient, make(map[*prometheus.Desc]slaveMetric)}

	defaultTaskLabels := []string{"source", "framework_id", "executor_id", "task_id", "task_name"}
	normalisedUserTaskLabelList := normaliseLabelList(userTaskLabelList)
	taskLabelList := append(defaultTaskLabels, normalisedUserTaskLabelList...)

	c.metrics[prometheus.NewDesc(
		prometheus.BuildFQName("mesos", "slave", "task_labels"),
		"Labels assigned to tasks running on slaves",
		taskLabelList,
		nil)] = slaveMetric{prometheus.CounterValue,
		func(st *slaveState) []metricValue {
			res := []metricValue{}
			for _, f := range st.Frameworks {
				for _, e := range f.Executors {
					for _, t := range e.Tasks {
						//Default labels
						taskLabels := prometheus.Labels{
							"source":       e.Source,
							"framework_id": f.ID,
							"executor_id":  e.ID,
							"task_id":      t.ID,
							"task_name":    t.Name,
						}

						// User labels
						for _, label := range normalisedUserTaskLabelList {
							taskLabels[label] = ""
						}
						for _, label := range t.Labels {
							normalisedLabel := normaliseLabel(label.Key)
							// Ignore labels not explicitly whitelisted by user
							if stringInSlice(normalisedLabel, normalisedUserTaskLabelList) {
								taskLabels[normalisedLabel] = label.Value
							}
						}

						res = append(res, metricValue{1, getLabelValuesFromMap(taskLabels, taskLabelList)})
					}
				}
			}
			return res
		},
	}

	c.metrics[prometheus.NewDesc(
		prometheus.BuildFQName("mesos", "slave", "task_cpus"),
		"CPUs allocated to tasks running on slaves",
		taskLabelList,
		nil)] = slaveMetric{prometheus.GaugeValue,
		func(st *slaveState) []metricValue {
			res := []metricValue{}
			for _, f := range st.Frameworks {
				for _, e := range f.Executors {
					for _, t := range e.Tasks {
						//Default labels
						taskLabels := prometheus.Labels{
							"source":       e.Source,
							"framework_id": f.ID,
							"executor_id":  e.ID,
							"task_id":      t.ID,
							"task_name":    t.Name,
						}


						res = append(res, metricValue{t.Resources.CPUs, getLabelValuesFromMap(taskLabels, taskLabelList)})
					}
				}
			}
			return res
		},
	}

	if len(slaveAttributeLabelList) > 0 {
		normalisedAttributeLabels := normaliseLabelList(slaveAttributeLabelList)

		c.metrics[prometheus.NewDesc(
			prometheus.BuildFQName("mesos", "slave", "attributes"),
			"Attributes assigned to slaves",
			normalisedAttributeLabels,
			nil)] = slaveMetric{prometheus.CounterValue,
			func(st *slaveState) []metricValue {
				slaveAttributes := prometheus.Labels{}

				for _, label := range normalisedAttributeLabels {
					slaveAttributes[label] = ""
				}
				for key, value := range st.Attributes {
					normalisedLabel := normaliseLabel(key)
					if stringInSlice(normalisedLabel, normalisedAttributeLabels) {
						if attribute, err := attributeString(value); err == nil {
							slaveAttributes[normalisedLabel] = attribute
						}
					}
				}

				return []metricValue{{1, getLabelValuesFromMap(slaveAttributes, normalisedAttributeLabels)}}
			},
		}
	}
	return &c
}

func (c *slaveStateCollector) Collect(ch chan<- prometheus.Metric) {
	var s slaveState
	log.WithField("url", "/slave(1)/state").Debug("fetching URL")
	c.fetchAndDecode("/slave(1)/state", &s)
	for d, cm := range c.metrics {
		for _, m := range cm.value(&s) {
			ch <- prometheus.MustNewConstMetric(d, cm.valueType, m.result, m.labels...)
		}
	}
}

func (c *slaveStateCollector) Describe(ch chan<- *prometheus.Desc) {
	for d := range c.metrics {
		ch <- d
	}
}
