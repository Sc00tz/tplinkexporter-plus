package collectors

import (
	"fmt"
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/thelastguardian/tplinkexporter/clients"
)

type TrafficCollector struct {
	namespace     string
	client        clients.TPLINKSwitchClient
	pktMetrics    map[string](*prometheus.GaugeVec)
	statusMetrics map[string](*prometheus.GaugeVec)
	configMetrics map[string](*prometheus.GaugeVec)

	// trafficScrapesTotalMetric              prometheus.Gauge
	// trafficScrapeErrorsTotalMetric         prometheus.Gauge
	// lastTrafficScrapeErrorMetric           prometheus.Gauge
	// lastTrafficScrapeTimestampMetric       prometheus.Gauge
	// lastTrafficScrapeDurationSecondsMetric prometheus.Gauge
}

var statusMetricsFields = []string{
	"State",
	"LinkStatus",
}

var pktMetricsFields = []string{
	"TxGoodPkt",
	"TxBadPkt",
	"RxGoodPkt",
	"RxBadPkt",
}

var configMetricsFields = []string{
	"SpeedConfig",
	"SpeedActual",
	"FlowControlConfig",
	"FlowControlActual",
	"TrunkInfo",
}

// formatPortNumber pads port numbers with leading zeros for proper sorting
func formatPortNumber(portnum int) string {
	return fmt.Sprintf("%02d", portnum+1)
}

func NewTrafficCollector(namespace string, client clients.TPLINKSwitchClient) *TrafficCollector {
	pktMetrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range pktMetricsFields {
		pktMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "portstats",
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' traffic metric from the router", name),
			},
			[]string{"portnum", "host"},
		)
	}
	
	statusMetrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range statusMetricsFields {
		statusMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "portstats",
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' status metric from the router", name),
			},
			[]string{"portnum", "host"},
		)
	}
	
	configMetrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range configMetricsFields {
		configMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "portconfig",
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' configuration metric from the router", name),
			},
			[]string{"portnum", "host"},
		)
	}
	
	return &TrafficCollector{
		namespace:     namespace,
		client:        client,
		pktMetrics:    pktMetrics,
		statusMetrics: statusMetrics,
		configMetrics: configMetrics,

		// trafficScrapesTotalMetric:              trafficScrapesTotalMetric,
		// trafficScrapeErrorsTotalMetric:         trafficScrapeErrorsTotalMetric,
		// lastTrafficScrapeErrorMetric:           lastTrafficScrapeErrorMetric,
		// lastTrafficScrapeTimestampMetric:       lastTrafficScrapeTimestampMetric,
		// lastTrafficScrapeDurationSecondsMetric: lastTrafficScrapeDurationSecondsMetric,
	}
}

func (c *TrafficCollector) Collect(ch chan<- prometheus.Metric) {
	// var begun = time.Now()

	// Collect port statistics (existing functionality)
	stats, err := c.client.GetPortStats()
	if err != nil {
		log.Printf("Error while collecting traffic statistics: %v", err)
		// c.trafficScrapeErrorsTotalMetric.Inc()
	} else {
		for portnum := 0; portnum < len(stats); portnum++ {
			paddedPortNum := formatPortNumber(portnum)
			
			for name, value := range stats[portnum].PktCount {
				// log.Printf("portnum '%s', metricname '%s', metricvalue '%d'", paddedPortNum, name, value)
				c.pktMetrics[name].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(value))
			}
			// log.Printf("portnum '%s', state '%d', linkstatus '%d'", paddedPortNum, stats[portnum].State, stats[portnum].LinkStatus)
			c.statusMetrics["State"].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(stats[portnum].State))
			c.statusMetrics["LinkStatus"].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(stats[portnum].LinkStatus))
		}
	}

	// Collect port configuration (new functionality)
	configs, err := c.client.GetPortConfig()
	if err != nil {
		log.Printf("Error while collecting port configuration: %v", err)
	} else {
		for _, config := range configs {
			paddedPortNum := formatPortNumber(config.Port - 1) // config.Port is 1-based
			
			c.configMetrics["SpeedConfig"].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(config.SpeedConfig))
			c.configMetrics["SpeedActual"].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(config.SpeedActual))
			c.configMetrics["FlowControlConfig"].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(config.FlowControlCfg))
			c.configMetrics["FlowControlActual"].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(config.FlowControlAct))
			c.configMetrics["TrunkInfo"].With(prometheus.Labels{"portnum": paddedPortNum, "host": c.client.GetHost()}).Set(float64(config.TrunkInfo))
		}
	}

	// Collect all metrics
	for name := range c.pktMetrics {
		c.pktMetrics[name].Collect(ch)
	}
	for name := range c.statusMetrics {
		c.statusMetrics[name].Collect(ch)
	}
	for name := range c.configMetrics {
		c.configMetrics[name].Collect(ch)
	}

	// c.trafficScrapeErrorsTotalMetric.Collect(ch)

	// c.trafficScrapesTotalMetric.Inc()
	// c.trafficScrapesTootalMetric.Collect(ch)

	// c.lastTrafficScrapeErrorMetric.Set(errorMetric)
	// c.lastTrafficScrapeErrorMetric.Collect(ch)

	// c.lastTrafficScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	// c.lastTrafficScrapeTimestampMetric.Collect(ch)

	// c.lastTrafficScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	// c.lastTrafficScrapeDurationSecondsMetric.Collect(ch)
}

func (c *TrafficCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, name := range pktMetricsFields {
		c.pktMetrics[name].Describe(ch)
	}
	for _, name := range statusMetricsFields {
		c.statusMetrics[name].Describe(ch)
	}
	for _, name := range configMetricsFields {
		c.configMetrics[name].Describe(ch)
	}

	// c.trafficScrapesTotalMetric.Describe(ch)
	// c.trafficScrapeErrorsTotalMetric.Describe(ch)
	// c.lastTrafficScrapeErrorMetric.Describe(ch)
	// c.lastTrafficScrapeTimestampMetric.Describe(ch)
	// c.lastTrafficScrapeDurationSecondsMetric.Describe(ch)
}