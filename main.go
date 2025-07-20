package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/thelastguardian/tplinkexporter/clients"
	"github.com/thelastguardian/tplinkexporter/collectors"
)

func main() {
	var (
		host     = kingpin.Flag("host", "Host of target tplink easysmart switch.").Required().String()
		username = kingpin.Flag("username", "Username for switch GUI login").Default("admin").String()
		password = kingpin.Flag("password", "Password for switch GUI login").Required().String()
		port     = kingpin.Flag("port", "Metrics port to listen on for prometheus scrapes").Default("9717").Int()
	)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// Create the TP-Link switch client
	tplinkSwitch := clients.NewTPLinkSwitch(*host, *username, *password)

	// Create all collectors
	trafficCollector := collectors.NewTrafficCollector("tplinkexporter", tplinkSwitch)
	systemInfoCollector := collectors.NewSystemInfoCollector("tplinkexporter", tplinkSwitch)
	portConfigCollector := collectors.NewPortConfigCollector("tplinkexporter", tplinkSwitch)

	// Register all collectors with Prometheus
	prometheus.MustRegister(trafficCollector)
	prometheus.MustRegister(systemInfoCollector)
	prometheus.MustRegister(portConfigCollector)

	// Set up HTTP handler for metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	
	// Optional: Add a simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Optional: Add a root endpoint with basic info
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
			<head><title>TP-Link Switch Exporter - PLUS</title></head>
			<body>
				<h1>TP-Link Switch Exporter - PLUS</h1>
				<p>Enhanced Prometheus exporter for TP-Link EasySmart switches</p>
				<ul>
					<li><a href="/metrics">Metrics</a></li>
					<li><a href="/health">Health Check</a></li>
				</ul>
				<h3>Available Metrics:</h3>
				<ul>
					<li><strong>Port Statistics:</strong> tplinkexporter_portstats_*</li>
					<li><strong>Port Configuration:</strong> tplinkexporter_portconfig_*</li>
					<li><strong>System Information:</strong> tplinkexporter_system_*</li>
				</ul>
				<p>Target Switch: ` + *host + `</p>
			</body>
			</html>
		`))
	})

	log.Printf("TP-Link Switch Exporter - PLUS starting on port :%d", *port)
	log.Printf("Target switch: %s", *host)
	log.Printf("Metrics available at http://localhost:%d/metrics", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}