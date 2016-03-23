package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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

	slave struct {
		PID        string    `json:"pid"`
		Used       resources `json:"used_resources"`
		Unreserved resources `json:"unreserved_resources"`
		Total      resources `json:"resources"`
	}

	framework struct {
		Active    bool   `json:"active"`
		Tasks     []task `json:"tasks"`
		Completed []task `json:"completed_tasks"`
	}

	state struct {
		Slaves     []slave     `json:"slaves"`
		Frameworks []framework `json:"frameworks"`
	}

	masterCollector struct {
		*http.Client
		url     string
		metrics map[prometheus.Collector]func(*state, prometheus.Collector)
	}
)

func newMasterStateCollector(url string, timeout time.Duration) *masterCollector {
	labels := []string{"slave"}
	return &masterCollector{
		Client: &http.Client{Timeout: timeout},
		url:    url,
		metrics: map[prometheus.Collector]func(*state, prometheus.Collector){
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Total slave CPUs (fractional)",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "cpus",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Total.CPUs)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Used slave CPUs (fractional)",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "cpus_used",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Used.CPUs)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Unreserved slave CPUs (fractional)",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "cpus_unreserved",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Unreserved.CPUs)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Total slave memory in bytes",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "mem_bytes",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Total.Mem * 1024)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Used slave memory in bytes",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "mem_used_bytes",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Used.Mem * 1024)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Unreserved slave memory in bytes",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "mem_unreserved_bytes",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Unreserved.Mem * 1024)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Total slave disk space in bytes",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "disk_bytes",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Total.Disk * 1024)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Used slave disk space in bytes",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "disk_used_bytes",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Used.Disk * 1024)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Unreserved slave disk in bytes",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "disk_unreserved_bytes",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(s.Unreserved.Disk * 1024)
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Total slave ports",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "ports",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					size := s.Total.Ports.size()
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(float64(size))
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Used slave ports",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "ports_used",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					size := s.Used.Ports.size()
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(float64(size))
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Unreserved slave ports",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "ports_unreserved",
			}, labels): func(st *state, c prometheus.Collector) {
				for _, s := range st.Slaves {
					size := s.Unreserved.Ports.size()
					c.(*prometheus.GaugeVec).WithLabelValues(s.PID).Set(float64(size))
				}
			},
			prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Help:      "Framework tasks",
				Namespace: "mesos",
				Subsystem: "slave",
				Name:      "task_state_time",
			}, []string{"slave", "task", "executor", "name", "framework", "state"}): func(st *state, c prometheus.Collector) {
				for _, f := range st.Frameworks {
					if !f.Active {
						continue
					}
					for _, task := range f.Completed {
						values := []string{
							task.ID,
							task.SlaveID,
							task.ExecutorID,
							task.Name,
							task.FrameworkID,
							task.State,
						}
						if len(task.Statuses) > 0 {
							c.(*prometheus.GaugeVec).WithLabelValues(values...).Set(task.Statuses[0].Timestamp)
						}
					}
				}
			},
		},
	}
}

func (c *masterCollector) Collect(ch chan<- prometheus.Metric) {
	res, err := c.Get(c.url + "/state")
	if err != nil {
		log.Print(err)
		return
	}
	defer res.Body.Close()

	var s state
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		log.Print(err)
		return
	}

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
