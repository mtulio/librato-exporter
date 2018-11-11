package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mtulio/librato-exporter/collector"
	"github.com/mtulio/librato-exporter/libratoAPI"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
)

// Flags setup
func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func initFlags() error {
	fListenAddress := flag.String("web.listen-address", config.listenAddress, "Address on which to expose metrics and web interface.")
	fMetricsPath := flag.String("web.telemetry-path", config.metricsPath, "Path under which to expose metrics.")

	fLibAccName := flag.String("librato.account", defaultAccName, "Librato alias for Account Name.")
	fLibEmail := flag.String("librato.email", "", "Librato Email account owner of token.")
	fLibToken := flag.String("librato.token", "", "Librato API Token created by Email.")
	fLibInterval := flag.Int("librato.interval", defaultInterval, "Interval in seconds to retrieve metrics from API")

	fLibMetricsFilter := flag.String("metrics.filter", "", "List of metrics sepparated by comma.")
	fLibMetricsResolution := flag.Int("metrics.resolution", defaultResolution, "Metrics resolution in seconds.")
	fLibMetricsOffset := flag.Int("metrics.offset", defaultOffset, "Time offset in seconds to define the start timestamp.")

	fVersion := flag.Bool("v", false, "prints current version")
	flag.Usage = usage
	flag.Parse()

	if *fVersion {
		fmt.Println(config.version)
		os.Exit(0)
	}

	if *fListenAddress != config.listenAddress {
		config.listenAddress = *fListenAddress
	}

	if *fMetricsPath != config.metricsPath {
		config.metricsPath = *fMetricsPath
	}

	if *fLibEmail == "" {
		log.Errorln("Librato Email must be provided.")
		os.Exit(1)
	} else {
		config.LibratoEmail = *fLibEmail
	}

	if *fLibToken == "" {
		log.Errorln("Librato Token must be provided.")
		os.Exit(1)
	} else {
		config.LibratoToken = *fLibToken
	}

	if *fLibMetricsFilter == "" {
		log.Errorln("Librato Metrics must be provided.")
		os.Exit(1)
	} else {
		config.LibratoMetrics = *fLibMetricsFilter
	}

	config.LibratoAccName = *fLibAccName
	config.LibratoInterval = uint32(*fLibInterval)
	config.LibratoMetRes = uint32(*fLibMetricsResolution)
	config.LibratoMetOffset = uint32(*fLibMetricsOffset)

	return nil
}

// Prometheus
func initPrometheusApp() {
	version.Version = VersionFull
	version.Revision = VersionTag
	prometheus.MustRegister(version.NewCollector(exporterName))
}

func initPrometheus(filters ...string) error {
	var err error
	err = nil

	prom.MC, err = collector.NewMasterCollector(libAPI, filters...)
	if err != nil {
		log.Warnln("Init Prom: Couldn't create collector: ", err)
		return err
	}

	prom.Registry = prometheus.NewRegistry()
	err = prom.Registry.Register(prom.MC)
	if err != nil {
		log.Errorln("Init Prom: Couldn't register collector:", err)
		return err
	}

	prom.Gatherers = &prometheus.Gatherers{
		prometheus.DefaultGatherer,
		prom.Registry,
	}
	return nil
}

// Librato API setup
func initAPI() error {
	var err error
	err = nil

	log.Info("Initializing Librato client...")

	libAPI, err = libratoAPI.NewLibratoAPI(
		config.LibratoAccName, config.LibratoEmail, config.LibratoToken,
	)
	if err != nil {
		log.Errorln("Init Librato API: ", err)
		return err
	}
	log.Infoln("Success")

	err = libAPI.SetCollectMetrics(strings.Split(config.LibratoMetrics, ",")...)
	if err != nil {
		log.Warnln("Init Prom: Couldn't set metrics: ", err)
		return err
	}
	libAPI.SetCollectInterval(config.LibratoInterval)
	libAPI.SetMetricsResolution(config.LibratoMetRes)
	libAPI.SetMetricsOffset(config.LibratoMetOffset)

	err = libAPI.GatherAll()
	if err != nil {
		log.Warnln("Init Prom: Couldn't create collector: ", err)
		return err
	}

	return nil
}
