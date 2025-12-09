# Prom-Verifier: Prometheus Alert Backtesting Tool

**Prom-Verifier** is a CLI tool designed to "Time Travel" and verify your Prometheus alerting rules against historical data.

It allows Site Reliability Engineers (SREs) to replay alert logic against past metrics to ensure that new alerts would have fired correctly (or to debug why an alert didn't fire), without needing to wait for a real outage.

## Features

- **🕰️ Time Travel Mode**: Verify alerts at any specific point in the past (e.g., "What would this alert have done last Tuesday at 2 PM?").
- **🛡️ Safety First**: Built-in guardrails prevent querying data older than 3 months or requesting massive windows (>4h) that could crash your Prometheus instance.
- **📝 Annotation Template Rendering**: Automatically renders alert annotations (e.g., `{{ $labels.foo }}`) so you can see exactly how the alert message will look when it fires.
- **developer-friendly**: Helps developers verify if their alerts work and preview how the rendered message will look, all before deploying to production.

## Requirements

**⚠️ Network Access Required**: This tool works by querying your actual Prometheus instance.
You **must** have direct network access to the Prometheus HTTP API (typically port 9090).
> If your infrastructure team has sealed off access (e.g., restricted VPC, VPN only, or internal-only Ingress), `prom-verifier` will fail to fetch the historical data.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/prom-verifier.git
cd prom-verifier

# Build the binary
go build -o prom-verifier ./cmd/prom-verifier
```

## Usage

By default, the tool looks for an `alert.yaml` file in the current directory and connects to `localhost:9090`.

### Basic Run
```bash
./prom-verifier
```

### Time Travel Example
Backtest a rule file against data from a specific date with a 15-minute window:

```bash
./prom-verifier \
  -url "http://prometheus.prod.svc:9090" \
  -file "my-alerts.yaml" \
  -at "2023-11-20 14:30" \
  -window "15m"
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-file` | Path to the alert rule file (YAML format) | `alert.yaml` |
| `-output` | Output format (`text`, `json`, `yaml`) | `text` |
| `-url` | Prometheus API URL | `http://localhost:9090` |
| `-at` | Target timestamp (`YYYY-MM-DD HH:MM`) for backtesting. | `Now` |
| `-window` | Time window size to fetch data for (max `4h`). | `30m` |

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

- `cmd/prom-verifier`: Main entry point.
- `internal/config`: Configuration parsing, flag handling, and input validation.
- `internal/client`: Prometheus API client wrapper.
- `internal/engine`: Core domain logic (Alert evaluation, Pending vs Firing state calculation).

## License

MIT
