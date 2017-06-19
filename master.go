package main

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func newMasterCollector(httpClient *httpClient) prometheus.Collector {
	metrics := map[prometheus.Collector]func(metricMap, prometheus.Collector) error{
		// CPU/Disk/Mem resources in free/used
		gauge("master", "cpus", "Current CPU resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["master/cpus_total"]
			used := m["master/cpus_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("master", "cpus_revocable", "Current revocable CPU resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["master/cpus_revocable_total"]
			used := m["master/cpus_revocable_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("master", "mem", "Current memory resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["master/mem_total"]
			used := m["master/mem_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("master", "mem_revocable", "Current revocable memory resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["master/mem_revocable_total"]
			used := m["master/mem_revocable_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("master", "disk", "Current disk resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["master/disk_total"]
			used := m["master/disk_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("master", "disk_revocable", "Current disk resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			total := m["master/disk_revocable_total"]
			used := m["master/disk_revocable_used"]

			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},

		// Master stats about uptime and election state
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "master",
			Name:      "elected",
			Help:      "1 if master is elected leader, 0 if not",
		}): func(m metricMap, c prometheus.Collector) error {
			elected, ok := m["master/elected"]
			if !ok {
				return errNotFoundInMap
			}
			c.(prometheus.Gauge).Set(elected)
			return nil
		},
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "master",
			Name:      "uptime_seconds",
			Help:      "Number of seconds the master process is running.",
		}): func(m metricMap, c prometheus.Collector) error {
			uptime, ok := m["master/uptime_secs"]
			if !ok {
				return errNotFoundInMap
			}
			c.(prometheus.Gauge).Set(uptime)
			return nil
		},
		// Master stats about agents
		counter("master", "slave_registration_events_total", "Total number of registration events on this master since it booted.", "event"): func(m metricMap, c prometheus.Collector) error {
			registrations := m["master/slave_registrations"]
			reregistrations := m["master/slave_reregistrations"]

			c.(*settableCounterVec).Set(registrations, "register")
			c.(*settableCounterVec).Set(reregistrations, "reregister")
			return nil
		},

		counter("master", "slave_removal_events_total", "Total number of removal events on this master since it booted.", "event"): func(m metricMap, c prometheus.Collector) error {
			scheduled := m["master/slave_shutdowns_scheduled"]
			canceled := m["master/slave_shutdowns_canceled"]
			completed := m["master/slave_shutdowns_completed"]
			removals := m["master/slave_removals"]

			c.(*settableCounterVec).Set(scheduled, "scheduled")
			c.(*settableCounterVec).Set(canceled, "canceled")
			c.(*settableCounterVec).Set(completed, "completed")
			c.(*settableCounterVec).Set(removals-completed, "died")
			return nil
		},
		gauge("master", "slaves_state", "Current number of slaves known to the master per connection and registration state.", "state"): func(m metricMap, c prometheus.Collector) error {
			active := m["master/slaves_active"]
			inactive := m["master/slaves_inactive"]
			disconnected := m["master/slaves_disconnected"]
			connected := m["master/slaves_connected"]

			c.(*prometheus.GaugeVec).WithLabelValues("active").Set(active)
			c.(*prometheus.GaugeVec).WithLabelValues("inactive").Set(inactive)
			c.(*prometheus.GaugeVec).WithLabelValues("disconnected").Set(disconnected)
			c.(*prometheus.GaugeVec).WithLabelValues("connected").Set(connected)
			return nil
		},

		// Master stats about frameworks
		gauge("master", "frameworks_state", "Current number of frames known to the master per connection and registration state.", "state"): func(m metricMap, c prometheus.Collector) error {
			active := m["master/frameworks_active"]
			connected := m["master/frameworks_connected"]
			inactive := m["master/frameworks_inactive"]
			disconnected := m["master/frameworks_disconnected"]

			c.(*prometheus.GaugeVec).WithLabelValues("active").Set(active)
			c.(*prometheus.GaugeVec).WithLabelValues("inactive").Set(inactive)
			c.(*prometheus.GaugeVec).WithLabelValues("disconnected").Set(disconnected)
			c.(*prometheus.GaugeVec).WithLabelValues("connected").Set(connected)
			return nil
		},
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "master",
			Name:      "offers_pending",
			Help:      "Current number of offers made by the master which aren't yet accepted or declined by frameworks.",
		}): func(m metricMap, c prometheus.Collector) error {
			offers := m["master/outstanding_offers"]

			c.(prometheus.Gauge).Set(offers)
			return nil
		},
		// Master stats about tasks
		counter("master", "task_states_exit_total", "Total number of tasks processed by exit state.", "state"): func(m metricMap, c prometheus.Collector) error {
			errored := m["master/tasks_error"]
			failed := m["master/tasks_failed"]
			finished := m["master/tasks_finished"]
			killed := m["master/tasks_killed"]
			lost := m["master/tasks_lost"]

			c.(*settableCounterVec).Set(errored, "errored")
			c.(*settableCounterVec).Set(failed, "failed")
			c.(*settableCounterVec).Set(finished, "finished")
			c.(*settableCounterVec).Set(killed, "killed")
			c.(*settableCounterVec).Set(lost, "lost")
			return nil
		},
		gauge("master", "task_states_current", "Current number of tasks by state.", "state"): func(m metricMap, c prometheus.Collector) error {
			running := m["master/tasks_running"]
			staging := m["master/tasks_staging"]
			starting := m["master/tasks_starting"]
			killing := m["master/tasks_killing"]
			unreachable := m["master/tasks_unreachable"]

			c.(*prometheus.GaugeVec).WithLabelValues("running").Set(running)
			c.(*prometheus.GaugeVec).WithLabelValues("staging").Set(staging)
			c.(*prometheus.GaugeVec).WithLabelValues("starting").Set(starting)
			c.(*prometheus.GaugeVec).WithLabelValues("killing").Set(killing)
			c.(*prometheus.GaugeVec).WithLabelValues("unreachable").Set(unreachable)
			return nil
		},

		// Master stats about messages
		counter("master", "messages_outcomes_total",
			"Total number of messages by outcome of operation and direction.",
			"source", "destination", "type", "outcome"): func(m metricMap, c prometheus.Collector) error {
			frameworkToExecutorValid := m["master/valid_framework_to_executor_messages"]
			frameworkToExecutorInvalid := m["master/invalid_framework_to_executor_messages"]
			executorToFrameworkValid := m["master/valid_executor_to_framework_messages"]
			executorToFrameworkInvalid := m["master/invalid_executor_to_framework_messages"]

			// status updates are sent from framework?(FIXME) to slave
			// status update acks are sent from slave to framework?
			statusUpdateAckValid := m["master/valid_status_update_acknowledgements"]
			statusUpdateAckInvalid := m["master/invalid_status_update_acknowledgements"]
			statusUpdateValid := m["master/valid_status_updates"]
			statusUpdateInvalid := m["master/invalid_status_updates"]

			c.(*settableCounterVec).Set(frameworkToExecutorValid, "framework", "executor", "", "valid")
			c.(*settableCounterVec).Set(frameworkToExecutorInvalid, "framework", "executor", "", "invalid")

			c.(*settableCounterVec).Set(executorToFrameworkValid, "executor", "framework", "", "valid")
			c.(*settableCounterVec).Set(executorToFrameworkInvalid, "executor", "framework", "", "invalid")

			// We consider a ack message simply as a message from slave to framework
			c.(*settableCounterVec).Set(statusUpdateValid, "framework", "slave", "status_update", "valid")
			c.(*settableCounterVec).Set(statusUpdateInvalid, "framework", "slave", "status_update", "invalid")
			c.(*settableCounterVec).Set(statusUpdateAckValid, "slave", "framework", "status_update", "valid")
			c.(*settableCounterVec).Set(statusUpdateAckInvalid, "slave", "framework", "status_update", "invalid")
			return nil
		},
		counter("master", "messages_type_total", "Total number of valid messages by type.", "type"): func(m metricMap, c prometheus.Collector) error {
			for k, v := range m {
				i := strings.Index("master/messages_", k)
				if i == -1 {
					continue
				}
				// FIXME: We expose things like messages_framework_to_executor twice
				c.(*settableCounterVec).Set(v, k[i:])
			}
			return nil
		},

		// Master stats about events
		gauge("master", "event_queue_length", "Current number of elements in event queue by type", "type"): func(m metricMap, c prometheus.Collector) error {
			dispatches := m["master/event_queue_dispatches"]
			httpRequests := m["master/event_queue_http_requests"]
			messages := m["master/event_queue_messages"]

			c.(*prometheus.GaugeVec).WithLabelValues("message").Set(messages)
			c.(*prometheus.GaugeVec).WithLabelValues("http_request").Set(httpRequests)
			c.(*prometheus.GaugeVec).WithLabelValues("dispatches").Set(dispatches)
			return nil
		},
	}
	return newMetricCollector(httpClient, metrics)
}
