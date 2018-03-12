package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

func newSlaveCollector(httpClient *httpClient) prometheus.Collector {
	metrics := map[prometheus.Collector]func(metricMap, prometheus.Collector) error{
		// CPU/Disk/Mem resources in free/used
		gauge("slave", "cpus", "Current CPU resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["slave/cpus_total"]
			used := m["slave/cpus_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "cpus_revocable", "Current revocable CPU resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["slave/cpus_revocable_total"]
			used := m["slave/cpus_revocable_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "mem", "Current memory resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["slave/mem_total"]
			used := m["slave/mem_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "mem_revocable", "Current revocable memory resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["slave/mem_revocable_total"]
			used := m["slave/mem_revocable_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "disk", "Current disk resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["slave/disk_total"]
			used := m["slave/disk_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "disk_revocable", "Current disk resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["slave/disk_revocable_total"]
			used := m["slave/disk_revocable_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},

		// Slave stats about uptime and connectivity
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "registered",
			Help:      "1 if slave is registered with master, 0 if not.",
		}): func(m metricMap, c prometheus.Collector) error {
			registered, ok := m["slave/registered"]
			if !ok {
				return errNotFoundInMap
			}
			c.(prometheus.Gauge).Set(registered)
			return nil
		},
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "uptime_seconds",
			Help:      "Number of seconds the master process is running.",
		}): func(m metricMap, c prometheus.Collector) error {
			uptime, ok := m["slave/uptime_secs"]
			if !ok {
				return errNotFoundInMap
			}
			c.(prometheus.Gauge).Set(uptime)
			return nil
		},

		// Slave stats about frameworks and executors
		gauge("slave", "executor_state", "Current number of executors by state.", "state"): func(m metricMap, c prometheus.Collector) error {
			registering := m["slave/executors_registering"]
			running := m["slave/executors_running"]
			terminating := m["slave/executors_terminating"]

			c.(*prometheus.GaugeVec).WithLabelValues("registering").Set(registering)
			c.(*prometheus.GaugeVec).WithLabelValues("running").Set(running)
			c.(*prometheus.GaugeVec).WithLabelValues("terminating").Set(terminating)
			return nil
		},
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "frameworks_active",
			Help:      "Current number of active frameworks",
		}): func(m metricMap, c prometheus.Collector) error {
			active, ok := m["slave/frameworks_active"]
			if !ok {
				return errNotFoundInMap
			}
			c.(prometheus.Gauge).Set(active)
			return nil
		},
		newSettableCounter("slave",
			"executors_terminated",
			"Total number of executor terminations."): func(m metricMap, c prometheus.Collector) error {
			terminated, ok := m["slave/executors_terminated"]
			if !ok {
				return errNotFoundInMap
			}
			c.(*settableCounter).Set(terminated)
			return nil
		},
		newSettableCounter("slave",
			"executors_preempted",
			"Total number of executor preemptions."): func(m metricMap, c prometheus.Collector) error {
			preempted, ok := m["slave/executors_preempted"]
			if !ok {
				return errNotFoundInMap
			}
			c.(*settableCounter).Set(preempted)
			return nil
		},

		// Slave stats about tasks
		counter("slave", "task_states_exit_total", "Total number of tasks processed by exit state.", "state"): func(m metricMap, c prometheus.Collector) error {
			errored := m["slave/tasks_error"]
			failed := m["slave/tasks_failed"]
			finished := m["slave/tasks_finished"]
			killed := m["slave/tasks_killed"]
			lost := m["slave/tasks_lost"]

			c.(*settableCounterVec).Set(errored, "errored")
			c.(*settableCounterVec).Set(failed, "failed")
			c.(*settableCounterVec).Set(finished, "finished")
			c.(*settableCounterVec).Set(killed, "killed")
			c.(*settableCounterVec).Set(lost, "lost")
			return nil
		},
		counter("slave", "task_states_current", "Current number of tasks by state.", "state"): func(m metricMap, c prometheus.Collector) error {
			running := m["slave/tasks_running"]
			staging := m["slave/tasks_staging"]
			starting := m["slave/tasks_starting"]

			c.(*settableCounterVec).Set(running, "running")
			c.(*settableCounterVec).Set(staging, "staging")
			c.(*settableCounterVec).Set(starting, "starting")
			return nil
		},

		// Slave stats about messages
		counter("slave", "messages_outcomes_total",
			"Total number of messages by outcome of operation",
			"type", "outcome"): func(m metricMap, c prometheus.Collector) error {

			frameworkMessagesValid := m["slave/valid_framework_messages"]
			frameworkMessagesInvalid := m["slave/invalid_framework_messages"]
			statusUpdateValid := m["slave/valid_status_updates"]
			statusUpdateInvalid := m["slave/invalid_status_updates"]

			c.(*settableCounterVec).Set(frameworkMessagesValid, "framework", "valid")
			c.(*settableCounterVec).Set(frameworkMessagesInvalid, "framework", "invalid")
			c.(*settableCounterVec).Set(statusUpdateValid, "status", "valid")
			c.(*settableCounterVec).Set(statusUpdateInvalid, "status", "invalid")

			return nil
		},
	}
	return newMetricCollector(httpClient, metrics)
}
