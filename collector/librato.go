package collector

import (
	"strconv"
	"sync"

	"github.com/mtulio/librato-exporter/libratoAPI"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	libratoCollectorSubsystem string
)

type LibratoCollector struct {
	Metrics []libratoCollectorMetric
}

type libratoCollectorMetric struct {
	LibratoMetric   *libratoAPI.LibratoMetric
	PromMetricErr   *prometheus.Desc
	PromMetricTs    *prometheus.Desc
	PromMetricValue *prometheus.Desc
}

func initLibratoCollector() {
	libratoCollectorSubsystem = globalLibAPI.GetAccName()
	registerCollector(libratoCollectorSubsystem, defaultEnabled, NewLibratoCollector)
}

//NewStkTestCollector is a Status Cake Test Collector
func NewLibratoCollector() (Collector, error) {
	var metrics []libratoCollectorMetric
	for m := range globalLibAPI.Metrics {
		newMetric := libratoCollectorMetric{
			LibratoMetric: globalLibAPI.Metrics[m],
			PromMetricTs: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, libratoCollectorSubsystem, "timestamp"),
				"Librato metric last timestamp",
				[]string{"name"}, nil,
			),
			PromMetricValue: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, libratoCollectorSubsystem, "value"),
				"Librato metric value",
				[]string{"name"}, nil,
			),
			PromMetricErr: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, libratoCollectorSubsystem, "up"),
				"Librato metric status for the last collect",
				[]string{"name"}, nil,
			),
		}
		metrics = append(metrics, newMetric)
	}

	return &LibratoCollector{
		Metrics: metrics,
	}, nil
}

// Update implements Collector and exposes related metrics
func (c *LibratoCollector) Update(ch chan<- prometheus.Metric) error {
	if err := c.updateLibratoMetrics(ch); err != nil {
		return err
	}
	return nil
}

func (c *LibratoCollector) updateLibratoMetrics(ch chan<- prometheus.Metric) error {

	if len(c.Metrics) < 1 {
		return nil
	}
	wg := sync.WaitGroup{}
	wg.Add(len(c.Metrics))
	for m := range c.Metrics {
		go func(ch chan<- prometheus.Metric, metric *libratoCollectorMetric) {
			if !metric.LibratoMetric.LatestCollectSuccess {
				ch <- prometheus.MustNewConstMetric(
					metric.PromMetricErr,
					prometheus.GaugeValue,
					1,
					metric.LibratoMetric.Name,
				)
				return
			}
			metricValue, err := strconv.ParseFloat(metric.LibratoMetric.Value, 64)
			if err != nil {
				ch <- prometheus.MustNewConstMetric(
					metric.PromMetricErr,
					prometheus.GaugeValue,
					1,
					metric.LibratoMetric.Name,
				)
				return
			}
			ch <- prometheus.MustNewConstMetric(
				metric.PromMetricTs,
				prometheus.CounterValue,
				float64(metric.LibratoMetric.LatestTimestamp),
				metric.LibratoMetric.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				metric.PromMetricValue,
				prometheus.GaugeValue,
				metricValue,
				metric.LibratoMetric.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				metric.PromMetricErr,
				prometheus.GaugeValue,
				1,
				metric.LibratoMetric.Name,
			)
			wg.Done()
		}(ch, &c.Metrics[m])

	}
	wg.Wait()

	return nil
}
