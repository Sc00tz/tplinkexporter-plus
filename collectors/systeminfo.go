package collectors

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/thelastguardian/tplinkexporter/clients"
)

type SystemInfoCollector struct {
	namespace   string
	client      clients.TPLINKSwitchClient
	systemInfo  *prometheus.GaugeVec
	networkInfo *prometheus.GaugeVec
}

func NewSystemInfoCollector(namespace string, client clients.TPLINKSwitchClient) *SystemInfoCollector {
	systemInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "system",
			Name:      "info",
			Help:      "System information including firmware and hardware versions",
		},
		[]string{"host", "device_description", "mac_address", "firmware_version", "hardware_version"},
	)

	networkInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "system",
			Name:      "network_info",
			Help:      "Network configuration information",
		},
		[]string{"host", "ip_address", "netmask", "gateway"},
	)

	return &SystemInfoCollector{
		namespace:   namespace,
		client:      client,
		systemInfo:  systemInfo,
		networkInfo: networkInfo,
	}
}

func (c *SystemInfoCollector) Collect(ch chan<- prometheus.Metric) {
	sysInfo, err := c.client.GetSystemInfo()
	if err != nil {
		log.Printf("Error while collecting system information: %v", err)
		return
	}

	// Set system info metric with labels containing the actual values
	c.systemInfo.With(prometheus.Labels{
		"host":               c.client.GetHost(),
		"device_description": sysInfo.DeviceDescription,
		"mac_address":        sysInfo.MACAddress,
		"firmware_version":   sysInfo.FirmwareVersion,
		"hardware_version":   sysInfo.HardwareVersion,
	}).Set(1)

	// Set network info metric with labels containing network configuration
	c.networkInfo.With(prometheus.Labels{
		"host":       c.client.GetHost(),
		"ip_address": sysInfo.IPAddress,
		"netmask":    sysInfo.Netmask,
		"gateway":    sysInfo.Gateway,
	}).Set(1)

	// Collect the metrics
	c.systemInfo.Collect(ch)
	c.networkInfo.Collect(ch)
}

func (c *SystemInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	c.systemInfo.Describe(ch)
	c.networkInfo.Describe(ch)
}