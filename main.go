package main


import (
    "flag"
    "fmt"
    "log"
    "net/http"

    // "os"
    // "strings"
    // "time"

    "github.com/halkyon/prometheus-hetrixtools-exporter/internal/collector"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    // "github.com/prometheus/common/version"
)

const (
    defaultListenAddress = ":8080" // Port where the Prometheus metrics will be exposed
    defaultMetricsPath   = "/metrics"
    namespace            = "hetrixtools"
)

func main() {
    var (
        listenAddress = flag.String("web.listen-address", defaultListenAddress, "Address on which to expose metrics and web interface.")
        metricsPath   = flag.String("web.telemetry-path", defaultMetricsPath, "Path under which to expose metrics.")
        apiKey        = flag.String("hetrixtools.api-key", "", "HetrixTools API key for authentication.")
    )

    flag.Parse()

    if *apiKey == "" {
        log.Fatal("HetrixTools API key is required")
    }

    // Create a new instance of the Collector
    hetrixCollector := collector.New(namespace, *apiKey)
    prometheus.MustRegister(hetrixCollector)

    // This will point /metrics to the promhttp handler
    http.Handle(*metricsPath, promhttp.Handler())
    fmt.Println("Beginning to serve on port", *listenAddress)
    log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

