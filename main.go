package main

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/RobustPerception/azure_metrics_exporter/config"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	sc = &config.SafeConfig{
		C: &config.Config{},
	}
	ac                    = NewAzureClient()
	configFile            = kingpin.Flag("config.file", "Azure exporter configuration file.").Default("azure.yml").String()
	listenAddress         = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9276").String()
	listMetricDefinitions = kingpin.Flag("list.definitions", "List available metric definitions for the given resources and exit.").Bool()
	invalidMetricChars    = regexp.MustCompile("[^a-zA-Z0-9_:]")
	memcache              = cache.New(10*time.Minute, 60*time.Minute)
)

func init() {
	prometheus.MustRegister(version.NewCollector("azure_exporter"))
}

// Collector generic collector type
type Collector struct{}

// Describe implemented with dummy data to satisfy interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (c *Collector) processMetricData(target config.Target, resourceID string, metricData insights.Response, err error, ch chan<- prometheus.Metric) {
	if err != nil {
		log.Printf("Failed to get metrics for target %s: %v", resourceID, err)
		return
	}

	if len(*metricData.Value) == 0 || len(*(*metricData.Value)[0].Timeseries) == 0 {
		log.Printf("Metrics not found for target %v\n", resourceID)
		return
	}
	if len(*(*(*metricData.Value)[0].Timeseries)[0].Data) == 0 {
		log.Printf("No metric data returned for target %v\n", resourceID)
		return
	}

	for _, value := range *metricData.Value {
		// Ensure Azure metric names conform to Prometheus metric name conventions
		metricName := strings.Replace(*value.Name.Value, " ", "_", -1)
		metricName = strings.ToLower(metricName + "_" + string(value.Unit))
		metricName = strings.Replace(metricName, "/", "_per_", -1)
		metricName = invalidMetricChars.ReplaceAllString(metricName, "_")
		if len(*value.Timeseries) > 0 && len(*(*value.Timeseries)[0].Data) > 0 {
			metricValue := (*(*value.Timeseries)[0].Data)[len(*(*value.Timeseries)[0].Data)-1]
			labels := CreateResourceLabels(*value.ID)

			if hasAggregation(target, "Total") && metricValue.Total != nil {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(metricName+"_total", metricName+"_total", nil, labels),
					prometheus.GaugeValue,
					*metricValue.Total,
				)
			}

			if hasAggregation(target, "Average") && metricValue.Average != nil {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(metricName+"_average", metricName+"_average", nil, labels),
					prometheus.GaugeValue,
					*metricValue.Average,
				)
			}

			if hasAggregation(target, "Minimum") && metricValue.Minimum != nil {

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(metricName+"_min", metricName+"_min", nil, labels),
					prometheus.GaugeValue,
					*metricValue.Minimum,
				)
			}

			if hasAggregation(target, "Maximum") && metricValue.Minimum != nil {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(metricName+"_max", metricName+"_max", nil, labels),
					prometheus.GaugeValue,
					*metricValue.Maximum,
				)
			}
		}
	}
}

// Collect - collect results from Azure Montior API and create Prometheus metrics.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Get metric values for all defined metrics
	for _, target := range sc.C.Targets {
		metrics := []string{}
		for _, metric := range target.Metrics {
			metrics = append(metrics, metric.Name)
		}
		var resourceIdList []string
		var err error
		fmt.Printf("Resource: %v Search: %v", target.Resource, target.Search)
		if len(target.Resource) > 0 {
			fmt.Println("Use one resource")
			resourceId := fmt.Sprintf("/subscriptions/%s%s", sc.C.Credentials.SubscriptionID, target.Resource)
			resourceIdList = make([]string, 1)
			resourceIdList[0] = resourceId
		} else if len(target.Search) > 0 {
			fmt.Println("Searching for resources...")
			resourceIdList, err = ac.getResources(target.Search)
			if err != nil {
				fmt.Errorf("Failed to get resources for search pattern \"%v\": %v ", target.Search, err)
			}
		} else {
			continue
		}
		fmt.Printf("Resources: %v", len(resourceIdList))
		for _, resourceID := range resourceIdList {
			metricValueData, err := ac.getMetricValue(metrics, target, resourceID)
			c.processMetricData(target, resourceID, metricValueData, err, ch)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()
	collector := &Collector{}
	registry.MustRegister(collector)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}

	err := ac.getAuthorizer()
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	// Print list of available metric definitions for each resource to console if specified.
	if *listMetricDefinitions {
		results, err := ac.getMetricDefinitions()
		if err != nil {
			log.Fatalf("Failed to fetch metric definitions: %v", err)
		}

		for k, v := range results {
			log.Printf("Resource: %s\n\nAvailable Metrics:\n", strings.Split(k, "/")[6])
			for _, r := range *v.Value {
				log.Printf("- %s\n", *r.Name.Value)
			}
		}
		os.Exit(0)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head>
            <title>Azure Exporter</title>
            </head>
            <body>
            <h1>Azure Exporter</h1>
						<p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})

	http.HandleFunc("/metrics", handler)
	log.Printf("azure_metrics_exporter listening on port %v", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
		os.Exit(1)
	}

}
