package collector

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/mtulio/librato-exporter/libratoAPI"
	"github.com/prometheus/client_golang/prometheus"
)

// MasterCollector implements the prometheus.Collector interface.
type MasterCollector struct {
	Collectors map[string]Collector
	LibratoAPI *libratoAPI.LibratoAPI
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

const (
	// Namespace defines the common namespace to be used by all metrics.
	namespace       = "librato"
	defaultEnabled  = true
	defaultDisabled = false
)

var (
	globalLibAPI       *libratoAPI.LibratoAPI
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"node_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"master_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
	factories      = make(map[string]func() (Collector, error))
	collectorState = make(map[string]*bool)
)

func registerCollector(collector string, isDefaultEnabled bool, factory func() (Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)

	flag := flag.Bool(flagName, isDefaultEnabled, flagHelp)
	collectorState[collector] = flag

	factories[collector] = factory
}

// NewMasterCollector creates a new NodeCollector.
func NewMasterCollector(libAPI *libratoAPI.LibratoAPI, filters ...string) (*MasterCollector, error) {
	globalLibAPI = libAPI
	initLibratoCollector()

	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		if *enabled {
			collector, err := factories[key]()
			if err != nil {
				return nil, err
			}
			if len(f) == 0 || f[key] {
				collectors[key] = collector
			}
		}
	}

	return &MasterCollector{
		Collectors: collectors,
		LibratoAPI: libAPI,
		// Metrics:    initLibratoMetrics(libAPI),
	}, nil
}

//Describe is a Prometheus implementation to be called by collector.
//It essentially writes all descriptors to the prometheus desc channel.
func (mc *MasterCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

//Collect implements required collect function for all promehteus collectors
func (mc *MasterCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(mc.Collectors))
	for name, c := range mc.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		log.Errorf("ERROR: %s collector failed after %fs: %s", name, duration.Seconds(), err)
		success = 0
	} else {
		log.Debugf("OK: %s collector succeeded after %fs.", name, duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (d *typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

// func initLibratoMetrics(libAPI *libratoAPI.LibratoAPI) []*libratoAPI.LibratoMetric {
// 	libAPI.SetCollectMetrics("static-cdn.responsetime")
// 	return libAPI.GetMetrics()
// 	// get metrics filters
// 	// create the prom metrics
// 	// return to the struct
// }
