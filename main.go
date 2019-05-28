package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func init() {
	flag.Parse()

	parsedLevel, err := log.ParseLevel(*rawLevel)
	if err != nil {
		log.Fatal(err)
	}
	logLevel = parsedLevel
}

var logLevel = log.InfoLevel
var bindAddr = flag.String("bind-addr", ":9189", "bind address for the metrics server")
var metricsPath = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
var rawLevel = flag.String("log-level", "info", "log level")
var metadataEndpoint = flag.String("metadata-endpoint", "http://169.254.169.254/latest/meta-data/", "metadata endpoint to query")

func main() {
	log.SetLevel(logLevel)
	log.Info("Starting spot-termination-exporter")

	log.Debug("registering term exporter")
	prometheus.MustRegister(NewTerminationCollector(*metadataEndpoint))

	go serveMetrics()

	exitChannel := make(chan os.Signal)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitSignal := <-exitChannel
	log.WithFields(log.Fields{"signal": exitSignal}).Infof("Caught %s signal, exiting", exitSignal)
}

func serveMetrics() {
	log.Infof("Starting metric http endpoint on %s", *bindAddr)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html>
		<head><title>Spot Termination Exporter</title></head>
		<body>
		<h1>Spot Termination Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
}
