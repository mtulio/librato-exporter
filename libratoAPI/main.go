package libratoAPI

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mtulio/go-librato/librato"
)

type LibratoMetric struct {
	Name                 string
	translateToName      string
	translateToLabels    string
	metricType           string
	metricUnit           string
	Value                string
	LatestTimestamp      int32
	LatestCollectSuccess bool
}

type LibratoAPI struct {
	client             *librato.Client
	accountName        string
	collectIntervalSec uint32
	metricsResolution  uint32
	metricsOffset      uint32
	Metrics            []*LibratoMetric
}

func NewLibratoAPI(accName string, email string, token string) (*LibratoAPI, error) {

	// connect to the Librato API
	c := librato.NewClient(email, token)

	return &LibratoAPI{
		client:             c,
		accountName:        accName,
		collectIntervalSec: 300,
		metricsOffset:      120,
	}, nil
}

// get/set
func (libAPI *LibratoAPI) SetCollectMetrics(metric ...string) error {
	for _, m := range metric {
		if m == "" {
			continue
		}
		libMetric := LibratoMetric{
			Name:                 m,
			LatestCollectSuccess: false,
		}
		libAPI.Metrics = append(libAPI.Metrics, &libMetric)
	}
	return nil
}

func (libAPI *LibratoAPI) GetMetrics() []*LibratoMetric {
	return libAPI.Metrics
}

func (libAPI *LibratoAPI) GetAccName() string {
	return libAPI.accountName
}

func (libAPI *LibratoAPI) SetCollectInterval(sec uint32) {
	libAPI.collectIntervalSec = sec
}

func (libAPI *LibratoAPI) SetMetricsResolution(res uint32) {
	libAPI.metricsResolution = res
}

func (libAPI *LibratoAPI) SetMetricsOffset(offset uint32) {
	libAPI.metricsOffset = offset
}

// gather functions
func (libAPI *LibratoAPI) GatherAll() error {
	go libAPI.gatherMetrics()
	return nil
}

type DataResult struct {
	Timestamp int32  `json:"timestamp"`
	Value     string `json:"value"`
}

type DataResults []DataResult

func (libAPI *LibratoAPI) gatherMetrics() {
	for {
		if len(libAPI.Metrics) <= 0 {
			time.Sleep(time.Second * 10)
			continue
		}
		for m := range libAPI.Metrics {
			tOffset := libAPI.metricsOffset
			tStart := time.Now().Unix() - int64(tOffset)
			resolutionSec := libAPI.metricsResolution

			mr := librato.MeasurementRetrievals{
				Name:       libAPI.Metrics[m].Name,
				StartTime:  int(tStart),
				Resolution: int(resolutionSec),
			}
			ms, r, err := libAPI.dsGetMeasurement(&mr)
			if err != nil {
				log.Println("GetMeasurement() error ", err)
				log.Println("GetMeasurement() return code ", uint(r.StatusCode))
				continue
			}
			// parse results data
			// TODO: supporting differents backends other than statsd
			tsData, err := dsLibratoParserStatsdLatest(ms)
			if err != nil {
				fmt.Printf(" Error getting metric %s: %v", libAPI.Metrics[m].Name, err)
			}
			if tsData.Timestamp > 0 {
				libAPI.Metrics[m].Value = tsData.Value
				libAPI.Metrics[m].LatestTimestamp = tsData.Timestamp
				libAPI.Metrics[m].LatestCollectSuccess = true
			} else {
				libAPI.Metrics[m].LatestCollectSuccess = false
			}
		}

		time.Sleep(time.Second * time.Duration(libAPI.collectIntervalSec))
	}
}

// Retrieve metric mensurements
func (libAPI *LibratoAPI) dsGetMeasurement(lm *librato.MeasurementRetrievals) (*librato.Metric, *http.Response, error) {
	return libAPI.client.Metrics.Get("", lm)
}

func (libAPI *LibratoAPI) dsGetMeasurementByName(name string) (*librato.Metric, *http.Response, error) {
	return libAPI.client.Metrics.Get(name, nil)
}

// Return the latest series from StatusD
func dsLibratoParserStatsdLatest(metrics_data *librato.Metric) (*DataResult, error) {

	// var rts DataResults
	type StatsdPayload struct {
		Statsd []librato.GaugeMeasurement `json:"statsd"`
	}
	var statsd StatsdPayload

	// Fill payload results with data source values
	// Remap (Marshal/Unmarshal) statsd payload to structs
	b, err := json.Marshal(metrics_data.Measurements)
	if err != nil {
		return nil, err
	}

	// fmt.Println(statsd)
	err2 := json.Unmarshal(b, &statsd)
	if err2 != nil {
		return nil, err2
	}

	// Filter statsd time series to API results response
	latestSeries := DataResult{
		Timestamp: 0,
		Value:     "",
	}
	for i := range statsd.Statsd {
		// Getting only the latest series
		if latestSeries.Timestamp > int32(*statsd.Statsd[i].MeasureTime) {
			fmt.Println("maiooor")
			continue
		}
		latestSeries.Timestamp = int32(*statsd.Statsd[i].MeasureTime)
		latestSeries.Value = fmt.Sprintf("%f", *statsd.Statsd[i].Value)
	}

	return &latestSeries, nil
}
