package collectors

import (
	"fmt"
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/thelastguardian/tplinkexporter/clients"
)

type PortConfigCollector struct {
	namespace     string
	client        clients.TPLINKSwitchClient
	configMetrics map[string](*prometheus.GaugeVec)
	portInfo      *prometheus.GaugeVec
}

var portConfigMetricsFields = []string{
	"State",
	"SpeedConfig",
	"SpeedActual", 
	"FlowControlConfig",
	"FlowControlActual",
	"TrunkInfo",
}

// formatPortNumber pads port numbers with leading zeros for proper sorting
func formatPortNumberConfig(portnum int) string {
	return fmt.Sprintf("%02d", portnum)
}

// speedToString converts speed numeric values to human readable strings
func speedToString(speed int) string {
	switch speed {
	case 0:
		return "Link Down"
	case 1:
		return "Auto"
	case 2:
		return "10MH"
	case 3:
		return "10MF"
	case 4:
		return "100MH"
	case 5:
		return "100MF"
	case 6:
		return "1000MF"
	default:
		return "Unknown"
	}
}

// stateToString converts state numeric values to human readable strings
func stateToString(state int) string {
	switch state {
	case 0:
		return "Disabled"
	case 1:
		return "Enabled"
	default:
		return "Unknown"
	}
}

// flowControlToString converts flow control numeric values to human readable strings
func flowControlToString(fc int) string {
	switch fc {
	case 0:
		return "Off"
	case 1:
		return "On"
	default:
		return "Unknown"
	}
}

// trunkToString converts trunk info to human readable strings
func trunkToString(trunk int) string {
	if trunk == 0 {
		return "None"
	}
	return fmt.Sprintf("LAG%d", trunk)
}

func NewPortConfigCollector(namespace string, client clients.TPLINKSwitchClient) *PortConfigCollector {
	configMetrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range portConfigMetricsFields {
		configMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "portconfig",
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Port configuration metric for '%s'", name),
			},
			[]string{"portnum", "host"},
		)
	}

	// Info metric with descriptive labels
	portInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "portconfig",
			Name:      "info",
			Help:      "Port configuration information with descriptive labels",
		},
		[]string{"portnum", "host", "state", "speed_config", "speed_actual", "flow_control_config", "flow_control_actual", "trunk_info"},
	)

	return &PortConfigCollector{
		namespace:     namespace,
		client:        client,
		configMetrics: configMetrics,
		portInfo:      portInfo,
	}
}

func (c *PortConfigCollector) Collect(ch chan<- prometheus.Metric) {
	configs, err := c.client.GetPortConfig()
	if err != nil {
		log.Printf("Error while collecting port configuration: %v", err)
		return
	}

	for _, config := range configs {
		paddedPortNum := formatPortNumberConfig(config.Port)

		// Set numeric metrics
		c.configMetrics["State"].With(prometheus.Labels{
			"portnum": paddedPortNum,
			"host":    c.client.GetHost(),
		}).Set(float64(config.State))

		c.configMetrics["SpeedConfig"].With(prometheus.Labels{
			"portnum": paddedPortNum,
			"host":    c.client.GetHost(),
		}).Set(float64(config.SpeedConfig))

		c.configMetrics["SpeedActual"].With(prometheus.Labels{
			"portnum": paddedPortNum,
			"host":    c.client.GetHost(),
		}).Set(float64(config.SpeedActual))

		c.configMetrics["FlowControlConfig"].With(prometheus.Labels{
			"portnum": paddedPortNum,
			"host":    c.client.GetHost(),
		}).Set(float64(config.FlowControlCfg))

		c.configMetrics["FlowControlActual"].With(prometheus.Labels{
			"portnum": paddedPortNum,
			"host":    c.client.GetHost(),
		}).Set(float64(config.FlowControlAct))

		c.configMetrics["TrunkInfo"].With(prometheus.Labels{
			"portnum": paddedPortNum,
			"host":    c.client.GetHost(),
		}).Set(float64(config.TrunkInfo))

		// Set info metric with human-readable labels
		c.portInfo.With(prometheus.Labels{
			"portnum":             paddedPortNum,
			"host":                c.client.GetHost(),
			"state":               stateToString(config.State),
			"speed_config":        speedToString(config.SpeedConfig),
			"speed_actual":        speedToString(config.SpeedActual),
			"flow_control_config": flowControlToString(config.FlowControlCfg),
			"flow_control_actual": flowControlToString(config.FlowControlAct),
			"trunk_info":          trunkToString(config.TrunkInfo),
		}).Set(1)
	}

	// Collect all metrics
	for name := range c.configMetrics {
		c.configMetrics[name].Collect(ch)
	}
	c.portInfo.Collect(ch)
}

func (c *PortConfigCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, name := range portConfigMetricsFields {
		c.configMetrics[name].Describe(ch)
	}
	c.portInfo.Describe(ch)
}