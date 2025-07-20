# TPLink EasySmart Switch Exporter - PLUS

Enhanced Prometheus exporter for TPLink's EasySmart Switches with comprehensive monitoring capabilities.

## Features

- **Port Statistics**: Traffic counters (TX/RX good/bad packets)
- **Port Configuration**: Speed settings, duplex mode, flow control, LAG membership
- **System Information**: Firmware version, hardware version, MAC address, network configuration
- **Proper Port Sorting**: Zero-padded port numbers (Port 01, Port 02, etc.)
- **Multi-Architecture**: Docker images for amd64 and arm64
- **Health Monitoring**: Built-in health check endpoint

## Tested Hardware

**TPLink EasySmart Gigabit 8 Port Switch (TL-SG108E):**
- v3 Firmware: 1.0.0 Build 20171214 Rel.70905
- v4 Firmware: 1.0.0 Build 20181120 Rel.40749

**TPLink EasySmart Gigabit 16 Port Switch (TL-SG116E):**
- Firmware: 1.0.0 Build 20180523 Rel.52122

Should work on other switches in the same EasySmart family, but testing is recommended.

## Quick Start

### Using Docker (Recommended)

```bash
docker run -d \
  --name tplink-exporter \
  -p 9717:9717 \
  thelastguardian/tplinkexporter-plus \
  --host 192.168.1.100 \
  --username admin \
  --password admin
```

### Using Go

```bash
go run main.go --host 192.168.1.100 --username admin --password admin
```

### Build from Source

```bash
git clone https://github.com/thelastguardian/tplinkexporter-plus
cd tplinkexporter-plus
go build -o tplinkexporter-plus
./tplinkexporter-plus --host 192.168.1.100 --username admin --password admin
```

## Command Line Options

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `--host` | IP address or hostname of the switch | - | Yes |
| `--username` | Web GUI username | `admin` | No |
| `--password` | Web GUI password | - | Yes |
| `--port` | Port for Prometheus metrics server | `9717` | No |

## Metrics Overview

### Port Statistics (`tplinkexporter_portstats_*`)

| Metric | Description | Labels |
|--------|-------------|--------|
| `tplinkexporter_portstats_state` | Port enabled/disabled status (0/1) | `portnum`, `host` |
| `tplinkexporter_portstats_linkstatus` | Link status (0=down, 6=1000MF, 5=100MF, etc.) | `portnum`, `host` |
| `tplinkexporter_portstats_txgoodpkt` | Transmitted good packets (counter) | `portnum`, `host` |
| `tplinkexporter_portstats_txbadpkt` | Transmitted bad packets (counter) | `portnum`, `host` |
| `tplinkexporter_portstats_rxgoodpkt` | Received good packets (counter) | `portnum`, `host` |
| `tplinkexporter_portstats_rxbadpkt` | Received bad packets (counter) | `portnum`, `host` |

### Port Configuration (`tplinkexporter_portconfig_*`)

| Metric | Description | Labels |
|--------|-------------|--------|
| `tplinkexporter_portconfig_state` | Port configuration state (0=disabled, 1=enabled) | `portnum`, `host` |
| `tplinkexporter_portconfig_speedconfig` | Configured speed (1=auto, 6=1000MF, etc.) | `portnum`, `host` |
| `tplinkexporter_portconfig_speedactual` | Actual negotiated speed | `portnum`, `host` |
| `tplinkexporter_portconfig_flowcontrolconfig` | Flow control configuration (0=off, 1=on) | `portnum`, `host` |
| `tplinkexporter_portconfig_flowcontrolactual` | Actual flow control state | `portnum`, `host` |
| `tplinkexporter_portconfig_trunkinfo` | LAG group membership (0=none, 1-8=LAG number) | `portnum`, `host` |
| `tplinkexporter_portconfig_info` | Port config with human-readable labels | `portnum`, `host`, `state`, `speed_config`, `speed_actual`, `flow_control_config`, `flow_control_actual`, `trunk_info` |

### System Information (`tplinkexporter_system_*`)

| Metric | Description | Labels |
|--------|-------------|--------|
| `tplinkexporter_system_info` | System hardware/software information | `host`, `device_description`, `mac_address`, `firmware_version`, `hardware_version` |
| `tplinkexporter_system_network_info` | Network configuration | `host`, `ip_address`, `netmask`, `gateway` |

## Speed/Link Status Values

| Value | Meaning |
|-------|---------|
| 0 | Link Down |
| 1 | Auto |
| 2 | 10M Half |
| 3 | 10M Full |
| 4 | 100M Half |
| 5 | 100M Full |
| 6 | 1000M Full |

## Endpoints

- **`/metrics`** - Prometheus metrics endpoint
- **`/health`** - Health check endpoint (returns `OK`)
- **`/`** - Information page with links and current configuration

## Grafana Dashboard

Enhanced Grafana dashboard available at: https://grafana.com/grafana/dashboards/12517

The dashboard includes:
- Port status overview
- Traffic rate graphs
- Port configuration details
- System information panels

## Docker Compose Example

```yaml
version: '3.8'
services:
  tplink-exporter:
    image: thelastguardian/tplinkexporter-plus:latest
    container_name: tplink-exporter
    ports:
      - "9717:9717"
    command:
      - --host=192.168.1.100
      - --username=admin
      - --password=admin
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9717/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Prometheus Configuration

Add this job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'tplink-switches'
    static_configs:
      - targets: ['localhost:9717']
    scrape_interval: 30s
    metrics_path: /metrics
```

## Example Queries

### Traffic Rate (packets per second)
```promql
rate(tplinkexporter_portstats_rxgoodpkt[5m])
```

### Ports with Link Down
```promql
tplinkexporter_portstats_linkstatus == 0
```

### Ports with Speed Mismatch (config vs actual)
```promql
tplinkexporter_portconfig_speedconfig != tplinkexporter_portconfig_speedactual
```

### System Information
```promql
tplinkexporter_system_info
```

## Building Multi-Architecture Docker Images

The project includes automated builds via Forgejo Actions that create images for both `linux/amd64` and `linux/arm64` platforms.

To build manually:
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t thelastguardian/tplinkexporter-plus:latest .
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source. See LICENSE file for details.

## Troubleshooting

### Common Issues

1. **Connection refused**: Verify the switch IP address and ensure the web interface is accessible
2. **Authentication failed**: Check username/password, default is usually `admin`/`admin`
3. **No metrics**: Ensure the switch web interface is enabled and accessible from the exporter host
4. **Wrong port numbers**: This version uses zero-padded port numbers (01, 02, etc.) for proper sorting

### Debug Mode

Add verbose logging by modifying the source code or check the container logs:
```bash
docker logs tplink-exporter
```

### Supported Switches

This exporter is designed for TP-Link EasySmart switches. It may work with other models, but testing is recommended. If you have success with other models, please contribute to the documentation.