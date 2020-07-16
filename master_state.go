package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type (
	slave struct {
		PID        string                     `json:"pid"`
		Hostname   string                     `json:"hostname"`
		Used       resources                  `json:"used_resources"`
		Unreserved resources                  `json:"unreserved_resources"`
		Total      resources                  `json:"resources"`
		Attributes map[string]json.RawMessage `json:"attributes"`
	}

	framework_resources struct {
		CPUs float64 `json:"cpus"`
		Disk float64 `json:"disk"`
		Mem  float64 `json:"mem"`
	}

	framework struct {
		Active    bool                `json:"active"`
		Tasks     []task              `json:"tasks"`
		Completed []task              `json:"completed_tasks"`
		Name      string              `json:"name"`
		ID		  string			  `json:"id"`
		Used      framework_resources `json:"used_resources"`
		Offered   framework_resources `json:"offered_resources"`
	}

	state struct {
		Slaves     []slave     `json:"slaves"`
		Frameworks []framework `json:"frameworks"`
	}

	masterCollector struct {
		*httpClient
		metrics map[prometheus.Collector]func(*state, prometheus.Collector)
	}
)

func newMasterStateCollector(httpClient *httpClient, slaveAttributeLabels []string) prometheus.Collector {
	labels := []string{"slave","hostname"}
	framework_labels := []string{"framework", "framework_id"}

	metrics := map[prometheus.Collector]func(*state, prometheus.Collector){
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Total slave CPUs (fractional)",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "cpus",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Total.CPUs)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Used slave CPUs (fractional)",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "cpus_used",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Used.CPUs)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Unreserved slave CPUs (fractional)",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "cpus_unreserved",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Unreserved.CPUs)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Total slave memory in bytes",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "mem_bytes",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Total.Mem * 1024)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Used slave memory in bytes",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "mem_used_bytes",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Used.Mem * 1024)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Unreserved slave memory in bytes",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "mem_unreserved_bytes",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Unreserved.Mem * 1024)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Total slave disk space in bytes",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "disk_bytes",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Total.Disk * 1024)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Used slave disk space in bytes",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "disk_used_bytes",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Used.Disk * 1024)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Unreserved slave disk in bytes",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "disk_unreserved_bytes",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(s.Unreserved.Disk * 1024)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Total slave ports",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "ports",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				size := s.Total.Ports.size()
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(float64(size))
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Used slave ports",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "ports_used",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				size := s.Used.Ports.size()
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(float64(size))
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Unreserved slave ports",
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "ports_unreserved",
		}, labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, s := range st.Slaves {
				size := s.Unreserved.Ports.size()
				c.(*prometheus.GaugeVec).WithLabelValues(s.PID, s.Hostname).Set(float64(size))
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Active framework",
			Namespace: "mesos",
			Subsystem: "framework",
			Name:      "active",
		}, framework_labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, f := range st.Frameworks {
				var active float64 = 0
				if f.Active {
					active = 1
				}
				c.(*prometheus.GaugeVec).WithLabelValues(f.Name, f.ID).Set(active)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Framework cpu used",
			Namespace: "mesos",
			Subsystem: "framework",
			Name:      "cpu_used",
		}, framework_labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, f := range st.Frameworks {
				c.(*prometheus.GaugeVec).WithLabelValues(f.Name, f.ID).Set(f.Used.CPUs)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Framework disk used",
			Namespace: "mesos",
			Subsystem: "framework",
			Name:      "disk_used",
		}, framework_labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, f := range st.Frameworks {
				c.(*prometheus.GaugeVec).WithLabelValues(f.Name, f.ID).Set(f.Used.Disk)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Framework memory used",
			Namespace: "mesos",
			Subsystem: "framework",
			Name:      "mem_used",
		}, framework_labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, f := range st.Frameworks {
				c.(*prometheus.GaugeVec).WithLabelValues(f.Name, f.ID).Set(f.Used.Mem)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Framework cpu offered",
			Namespace: "mesos",
			Subsystem: "framework",
			Name:      "cpu_offered",
		}, framework_labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, f := range st.Frameworks {
				c.(*prometheus.GaugeVec).WithLabelValues(f.Name, f.ID).Set(f.Offered.CPUs)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Framework mem offered",
			Namespace: "mesos",
			Subsystem: "framework",
			Name:      "mem_offered",
		}, framework_labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, f := range st.Frameworks {
				c.(*prometheus.GaugeVec).WithLabelValues(f.Name, f.ID).Set(f.Offered.Mem)
			}
		},
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "Framework disk offered",
			Namespace: "mesos",
			Subsystem: "framework",
			Name:      "disk_offered",
		}, framework_labels): func(st *state, c prometheus.Collector) {
			c.(*prometheus.GaugeVec).Reset()
			for _, f := range st.Frameworks {
				c.(*prometheus.GaugeVec).WithLabelValues(f.Name, f.ID).Set(f.Offered.Disk)
			}
		},
	}

	if len(slaveAttributeLabels) > 0 {
		normalisedAttributeLabels := normaliseLabelList(slaveAttributeLabels)
		slaveAttributesLabelsExport := append(labels, normalisedAttributeLabels...)

		metrics[counter("slave", "attributes", "Attributes assigned to slaves", slaveAttributesLabelsExport...)] = func(st *state, c prometheus.Collector) {
			for _, s := range st.Slaves {
				slaveAttributesExport := prometheus.Labels{
					"slave": s.PID,
				}

				// User labels
				for _, label := range normalisedAttributeLabels {
					slaveAttributesExport[label] = ""
				}
				for key, value := range s.Attributes {
					normalisedLabel := normaliseLabel(key)
					if stringInSlice(normalisedLabel, normalisedAttributeLabels) {
						if attribute, err := attributeString(value); err == nil {
							slaveAttributesExport[normalisedLabel] = attribute
						}
					}
				}
				c.(*settableCounterVec).Set(1, getLabelValuesFromMap(slaveAttributesExport, slaveAttributesLabelsExport)...)
			}
		}
	}

	return &masterCollector{
		httpClient: httpClient,
		metrics:    metrics,
	}
}

func (c *masterCollector) Collect(ch chan<- prometheus.Metric) {
	var s state
	log.WithField("url", "/state").Debug("fetching URL")
	c.fetchAndDecode("/state", &s)

	for c, set := range c.metrics {
		set(&s, c)
		c.Collect(ch)
	}
}

func (c *masterCollector) Describe(ch chan<- *prometheus.Desc) {
	for metric := range c.metrics {
		metric.Describe(ch)
	}
}

type ranges [][2]uint64

func (rs *ranges) UnmarshalJSON(data []byte) (err error) {
	if data = bytes.Trim(data, `[]"`); len(data) == 0 {
		return nil
	}

	var rng [2]uint64
	for _, r := range bytes.Split(data, []byte(",")) {
		ps := bytes.SplitN(r, []byte("-"), 2)
		if len(ps) != 2 {
			return fmt.Errorf("bad range: %s", r)
		}

		rng[0], err = strconv.ParseUint(string(bytes.TrimSpace(ps[0])), 10, 64)
		if err != nil {
			return err
		}

		rng[1], err = strconv.ParseUint(string(bytes.TrimSpace(ps[1])), 10, 64)
		if err != nil {
			return err
		}

		*rs = append(*rs, rng)
	}

	return nil
}

func (rs ranges) size() uint64 {
	var sz uint64
	for i := range rs {
		sz += 1 + (rs[i][1] - rs[i][0])
	}
	return sz
}
