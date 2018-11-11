package main

import (
	"github.com/mtulio/librato-exporter/collector"
	"github.com/mtulio/librato-exporter/libratoAPI"
	"github.com/prometheus/client_golang/prometheus"
)

type globalConf struct {
	listenAddress    string
	metricsPath      string
	version          string
	versionCm        string
	versionTag       string
	versionEnv       string
	LibratoEmail     string
	LibratoAccName   string
	LibratoToken     string
	LibratoMetrics   string
	LibratoInterval  uint32
	LibratoMetRes    uint32
	LibratoMetOffset uint32
}

type globalProm struct {
	MC        *collector.MasterCollector
	Registry  *prometheus.Registry
	Gatherers *prometheus.Gatherers
}

const (
	exporterName        = "librato_exporter"
	exporterDescription = "Librato Exporter"
	defaultListenPort   = ":9800"
	defaultMetricsPath  = "/metrics"
	defaultInterval     = 300
	defaultResolution   = 30
	defaultOffset       = 120
	defaultAccName      = "app"
)

var (
	// VersionCommit is a compiler exporterd var
	VersionCommit string
	VersionTag    string
	VersionFull   string
	VersionEnv    string

	// Global vars
	config = globalConf{
		defaultListenPort,
		defaultMetricsPath,
		VersionFull,
		VersionCommit,
		VersionTag,
		VersionEnv,
		defaultAccName,
		"",
		"",
		"",
		defaultInterval,
		defaultResolution,
		defaultOffset,
	}
	libAPI *libratoAPI.LibratoAPI
	prom   globalProm
)
