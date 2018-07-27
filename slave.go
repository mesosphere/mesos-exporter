package main

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func newSlaveCollector(httpClient *httpClient) prometheus.Collector {
	metrics := map[prometheus.Collector]func(metricMap, prometheus.Collector) error{
		// CPU/Disk/Mem resources in free/used
		gauge("slave", "cpus", "Current CPU resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/cpus_percent"]
			if !ok {
				log.WithField("metric", "slave/cpus_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/cpus_total"]
			if !ok {
				log.WithField("metric", "slave/cpus_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/cpus_used"]
			if !ok {
				log.WithField("metric", "slave/cpus_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "cpus_revocable", "Current revocable CPU resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/cpus_revocable_percent"]
			if !ok {
				log.WithField("metric", "slave/cpus_revocable_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/cpus_revocable_total"]
			if !ok {
				log.WithField("metric", "slave/cpus_revocable_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/cpus_revocable_used"]
			if !ok {
				log.WithField("metric", "slave/cpus_revocable_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "mem", "Current memory resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/mem_percent"]
			if !ok {
				log.WithField("metric", "slave/mem_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/mem_total"]
			if !ok {
				log.WithField("metric", "slave/mem_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/mem_used"]
			if !ok {
				log.WithField("metric", "slave/mem_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "mem_revocable", "Current revocable memory resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/mem_revocable_percent"]
			if !ok {
				log.WithField("metric", "slave/mem_revocable_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/mem_revocable_total"]
			if !ok {
				log.WithField("metric", "slave/mem_revocable_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/mem_revocable_used"]
			if !ok {
				log.WithField("metric", "slave/mem_revocable_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "gpus", "Current GPU resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/gpus_percent"]
			if !ok {
				log.WithField("metric", "slave/gpus_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/gpus_total"]
			if !ok {
				log.WithField("metric", "slave/gpus_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/gpus_used"]
			if !ok {
				log.WithField("metric", "slave/gpus_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "gpus_revocable", "Current revocable GPUS resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/gpus_revocable_percent"]
			if !ok {
				log.WithField("metric", "slave/gpus_revocable_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/gpus_revocable_total"]
			if !ok {
				log.WithField("metric", "slave/gpus_revocable_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/gpus_revocable_used"]
			if !ok {
				log.WithField("metric", "slave/gpus_revocable_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "disk", "Current disk resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/disk_percent"]
			if !ok {
				log.WithField("metric", "slave/disk_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/disk_total"]
			if !ok {
				log.WithField("metric", "slave/disk_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/disk_used"]
			if !ok {
				log.WithField("metric", "slave/disk_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			return nil
		},
		gauge("slave", "disk_revocable", "Current disk resources in cluster.", "type"): func(m metricMap, c prometheus.Collector) error {
			percent, ok := m["slave/disk_revocable_percent"]
			if !ok {
				log.WithField("metric", "slave/disk_revocable_percent").Warn(LogErrNotFoundInMap)
			}
			total, ok := m["slave/disk_revocable_total"]
			if !ok {
				log.WithField("metric", "slave/disk_revocable_total").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["slave/disk_revocable_used"]
			if !ok {
				log.WithField("metric", "slave/disk_revocable_used").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("percent").Set(percent)
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
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
				log.WithField("metric", "slave/registered").Warn(LogErrNotFoundInMap)
			}
			c.(prometheus.Gauge).Set(registered)
			return nil
		},
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "uptime_seconds",
			Help:      "Number of seconds the slave process is running.",
		}): func(m metricMap, c prometheus.Collector) error {
			uptime, ok := m["slave/uptime_secs"]
			if !ok {
				log.WithField("metric", "slave/uptime_seconds").Warn(LogErrNotFoundInMap)
			}
			c.(prometheus.Gauge).Set(uptime)
			return nil
		},
		newSettableCounter("slave",
			"recovery_errors",
			"Total number of recovery errors"): func(m metricMap, c prometheus.Collector) error {
			errors, ok := m["slave/recovery_errors"]
			if !ok {
				log.WithField("metric", "slave/recovery_errors").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(errors)
			return nil
		},
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "recovery_time_secs",
			Help:      "Agent recovery time in seconds",
		}): func(m metricMap, c prometheus.Collector) error {
			age, ok := m["slave/recovery_time_secs"]
			if !ok {
				log.WithField("metric", "slave/recovery_time_secs").Warn(LogErrNotFoundInMap)
			}
			c.(prometheus.Gauge).Set(age)
			return nil
		},
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "executor_directory_max_allowed_age_secs",
			Help:      "Max allowed age of the executor directory",
		}): func(m metricMap, c prometheus.Collector) error {
			age, ok := m["slave/executor_directory_max_allowed_age_secs"]
			if !ok {
				log.WithField("metric", "slave/executor_directory_max_allowed_age_secs").Warn(LogErrNotFoundInMap)
			}
			c.(prometheus.Gauge).Set(age)
			return nil
		},

		// Slave stats about frameworks and executors
		gauge("slave", "executor_state", "Current number of executors by state.", "state"): func(m metricMap, c prometheus.Collector) error {
			registering, ok := m["slave/executors_registering"]
			if !ok {
				log.WithField("metric", "slave/executors_registering").Warn(LogErrNotFoundInMap)
			}
			running, ok := m["slave/executors_running"]
			if !ok {
				log.WithField("metric", "slave/executors_running").Warn(LogErrNotFoundInMap)
			}
			terminating, ok := m["slave/executors_terminating"]
			if !ok {
				log.WithField("metric", "slave/executors_terminating").Warn(LogErrNotFoundInMap)
			}
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
				log.WithField("metric", "slave/frameworks_active").Warn(LogErrNotFoundInMap)
			}
			c.(prometheus.Gauge).Set(active)
			return nil
		},
		newSettableCounter("slave",
			"executors_terminated",
			"Total number of executor terminations."): func(m metricMap, c prometheus.Collector) error {
			terminated, ok := m["slave/executors_terminated"]
			if !ok {
				log.WithField("metric", "slave/executors_terminated").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(terminated)
			return nil
		},
		newSettableCounter("slave",
			"executors_preempted",
			"Total number of executor preemptions."): func(m metricMap, c prometheus.Collector) error {
			preempted, ok := m["slave/executors_preempted"]
			if !ok {
				log.WithField("metric", "slave/executors_preempted").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(preempted)
			return nil
		},

		// Slave stats about tasks
		counter("slave", "task_states_exit_total", "Total number of tasks processed by exit state.", "state"): func(m metricMap, c prometheus.Collector) error {
			errored, ok := m["slave/tasks_error"]
			if !ok {
				log.WithField("metric", "slave/tasks_error").Warn(LogErrNotFoundInMap)
			}
			failed, ok := m["slave/tasks_failed"]
			if !ok {
				log.WithField("metric", "slave/tasks_failed").Warn(LogErrNotFoundInMap)
			}
			finished, ok := m["slave/tasks_finished"]
			if !ok {
				log.WithField("metric", "slave/tasks_finished").Warn(LogErrNotFoundInMap)
			}
			gone, ok := m["slave/tasks_gone"]
			if !ok {
				log.WithField("metric", "slave/tasks_gone").Warn(LogErrNotFoundInMap)
			}
			killed, ok := m["slave/tasks_killed"]
			if !ok {
				log.WithField("metric", "slave/tasks_killed").Warn(LogErrNotFoundInMap)
			}
			killing, ok := m["slave/tasks_killing"]
			if !ok {
				log.WithField("metric", "slave/tasks_killing").Warn(LogErrNotFoundInMap)
			}
			lost, ok := m["slave/tasks_lost"]
			if !ok {
				log.WithField("metric", "slave/tasks_lost").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounterVec).Set(errored, "errored")
			c.(*settableCounterVec).Set(failed, "failed")
			c.(*settableCounterVec).Set(finished, "finished")
			c.(*settableCounterVec).Set(gone, "gone")
			c.(*settableCounterVec).Set(killed, "killed")
			c.(*settableCounterVec).Set(killing, "killing")
			c.(*settableCounterVec).Set(lost, "lost")
			return nil
		},
		counter("slave", "task_states_current", "Current number of tasks by state.", "state"): func(m metricMap, c prometheus.Collector) error {
			running, ok := m["slave/tasks_running"]
			if !ok {
				log.WithField("metric", "slave/tasks_running").Warn(LogErrNotFoundInMap)
			}
			staging, ok := m["slave/tasks_staging"]
			if !ok {
				log.WithField("metric", "slave/tasks_staging").Warn(LogErrNotFoundInMap)
			}
			starting, ok := m["slave/tasks_starting"]
			if !ok {
				log.WithField("metric", "slave/tasks_starting").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounterVec).Set(running, "running")
			c.(*settableCounterVec).Set(staging, "staging")
			c.(*settableCounterVec).Set(starting, "starting")
			return nil
		},

		counter("slave", "task_state_counts_by_source_reason", "Number of task states by source and reason", "state", "source", "reason"): func(m metricMap, c prometheus.Collector) error {
			re, err := regexp.Compile("slave/task_(.*?)/source_(.*?)/reason_(.*?)$")
			if err != nil {
				log.WithFields(log.Fields{
					"regex":  "slave/task_(.*?)/source_(.*?)/reason_(.*?)$",
					"metric": "slave_task_state_counts_by_source_reason",
					"error":  err,
				}).Error("could not compile regex")
				return fmt.Errorf("could not compile slave_task_state_counts_by_source_reason regex: %s", err)
			}
			for metric, value := range m {
				matches := re.FindStringSubmatch(metric)
				if len(matches) != 4 {
					continue
				}
				state := matches[1]
				source := matches[2]
				reason := matches[3]
				c.(*settableCounterVec).Set(value, state, source, reason)
			}
			return nil
		},

		// Slave stats about messages
		counter("slave", "messages_outcomes_total",
			"Total number of messages by outcome of operation",
			"type", "outcome"): func(m metricMap, c prometheus.Collector) error {

			frameworkMessagesValid, ok := m["slave/valid_framework_messages"]
			if !ok {
				log.WithField("metric", "slave/valid_framework_messages").Warn(LogErrNotFoundInMap)
			}
			frameworkMessagesInvalid, ok := m["slave/invalid_framework_messages"]
			if !ok {
				log.WithField("metric", "slave/invalid_framework_messages").Warn(LogErrNotFoundInMap)
			}
			statusUpdateValid, ok := m["slave/valid_status_updates"]
			if !ok {
				log.WithField("metric", "slave/valid_status_updates").Warn(LogErrNotFoundInMap)
			}
			statusUpdateInvalid, ok := m["slave/invalid_status_updates"]
			if !ok {
				log.WithField("metric", "slave/invalid_status_updates").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounterVec).Set(frameworkMessagesValid, "framework", "valid")
			c.(*settableCounterVec).Set(frameworkMessagesInvalid, "framework", "invalid")
			c.(*settableCounterVec).Set(statusUpdateValid, "status", "valid")
			c.(*settableCounterVec).Set(statusUpdateInvalid, "status", "invalid")

			return nil
		},

		// GC information
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mesos",
			Subsystem: "slave",
			Name:      "gc_path_removals_pending",
			Help:      "Number of sandbox paths that are currently pending agent garbage collection",
		}): func(m metricMap, c prometheus.Collector) error {
			pending, ok := m["gc/path_removals_pending"]
			if !ok {
				log.WithField("metric", "gc/path_removals_pending").Warn(LogErrNotFoundInMap)
			}
			c.(prometheus.Gauge).Set(pending)
			return nil
		},
		counter("slave", "gc_path_removals_outcome",
			"Number of sandbox paths the agent removed",
			"outcome"): func(m metricMap, c prometheus.Collector) error {

			succeeded, ok := m["gc/path_removals_succeeded"]
			if !ok {
				log.WithField("metric", "gc/path_removals_succeeded").Warn(LogErrNotFoundInMap)
			}
			failed, ok := m["gc/path_removals_failed"]
			if !ok {
				log.WithField("metric", "gc/path_removals_failed").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounterVec).Set(succeeded, "success")
			c.(*settableCounterVec).Set(failed, "failed")

			return nil
		},

		// Container / Containerizer information
		newSettableCounter("slave",
			"container_launch_errors",
			"Total number of container launch errors"): func(m metricMap, c prometheus.Collector) error {
			errors, ok := m["slave/container_launch_errors"]
			if !ok {
				log.WithField("metric", "slave/container_launch_errors").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(errors)
			return nil
		},
		newSettableCounter("slave",
			"containerizer_filesystem_containers_new_rootfs",
			"Number of containers changing root filesystem"): func(m metricMap, c prometheus.Collector) error {
			newRootfs, ok := m["containerizer/mesos/filesystem/containers_new_rootfs"]
			if !ok {
				log.WithField("metric", "containerizer/mesos/filesystem/containers_new_rootfs").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(newRootfs)
			return nil
		},
		newSettableCounter("slave",
			"containerizer_provisioner_bind_remove_rootfs_errors",
			"Number of errors from the containerizer attempting to bind the rootfs"): func(m metricMap, c prometheus.Collector) error {
			errors, ok := m["containerizer/mesos/provisioner/bind/remove_rootfs_errors"]
			if !ok {
				log.WithField("metric", "containerizer/mesos/provisioner/bind/remove_rootfs_errors").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(errors)
			return nil
		},
		newSettableCounter("slave",
			"containerizer_provisioner_remove_container_errors",
			"Number of errors from the containerizer attempting to remove a container"): func(m metricMap, c prometheus.Collector) error {
			errors, ok := m["containerizer/mesos/provisioner/remove_container_errors"]
			if !ok {
				log.WithField("metric", "containerizer/mesos/provisioner/remove_container_errors").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(errors)
			return nil
		},
		newSettableCounter("slave",
			"containerizer_container_destroy_errors",
			"Number of containers destroyed due to launch errors"): func(m metricMap, c prometheus.Collector) error {
			errors, ok := m["containerizer/mesos/container_destroy_errors"]
			if !ok {
				log.WithField("metric", "containerizer/mesos/container_destroy_errors").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounter).Set(errors)
			return nil
		},
		counter("slave", "containerizer_fetcher_task_fetches",
			"Total number of containerizer fetcher tasks by outcome",
			"outcome"): func(m metricMap, c prometheus.Collector) error {

			succeeded, ok := m["containerizer/fetcher/task_fetches_succeeded"]
			if !ok {
				log.WithField("metric", "containerizer/fetcher/task_fetches_succeeded").Warn(LogErrNotFoundInMap)
			}
			failed, ok := m["containerizer/fetcher/task_fetches_failed"]
			if !ok {
				log.WithField("metric", "containerizer/fetcher/task_fetches_failed").Warn(LogErrNotFoundInMap)
			}
			c.(*settableCounterVec).Set(succeeded, "success")
			c.(*settableCounterVec).Set(failed, "failed")

			return nil
		},
		gauge("slave", "containerizer_fetcher_cache_size", "Containerizer fetcher cache sizes in bytes", "type"): func(m metricMap, c prometheus.Collector) error {
			total, ok := m["containerizer/fetcher/cache_size_total_bytes"]
			if !ok {
				log.WithField("metric", "containerizer/fetcher/cache_size_total_bytes").Warn(LogErrNotFoundInMap)
			}
			used, ok := m["containerizer/fetcher/cache_size_used_bytes"]
			if !ok {
				log.WithField("metric", "containerizer/fetcher/cache_size_used_bytes").Warn(LogErrNotFoundInMap)
			}
			c.(*prometheus.GaugeVec).WithLabelValues("total").Set(total)
			c.(*prometheus.GaugeVec).WithLabelValues("used").Set(used)
			c.(*prometheus.GaugeVec).WithLabelValues("free").Set(total - used)
			return nil
		},

		// END
	}
	return newMetricCollector(httpClient, metrics)
}
