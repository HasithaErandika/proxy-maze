# ProxyMaze'26
**Real-Time Proxy Intelligence — Torch Labs Engineering Challenge**

This is a production-ready, asynchronous proxy monitoring system built in Go. It tracks proxy health in real-time, features a strict alert lifecycle with automated webhooks, and directly integrates with Slack and Discord.

## Quick Start

Run the application locally:
```bash
go mod tidy
go run main.go
```
By default, the API will be available at `http://localhost:8080`.

## Docker

To run the application via Docker:
```bash
docker compose up --build
```

## Environment Variables

- `PORT`: The HTTP port for the application to listen on (default is `8080`).

## Architecture Overview

ProxyMaze features an asynchronous architecture built around Go's concurrency model (`goroutines` and `channels/mutexes`). A background ticker continuously sweeps through a thread-safe proxy pool, firing parallel HTTP probes to measure uptime and latency. If the calculated failure rate exceeds a constant threshold (20%), the Alert Manager state machine kicks in—misting a unique `alert_id` and dispatching real-time notifications to registered webhooks and integrations (Slack/Discord) via a resilient, retry-backed worker queue. This guarantees API responsiveness regardless of pool size or webhook delivery failures.

## Alert Lifecycle

```text
  [Healthy Pool]
        |
   (Failure Rate >= 20%)
        |
        v
  [ALERT FIRED] -------> Dispatch Webhooks & Integrations
        |                  (Status: active)
        |
   (Failure Rate < 20%)
        |
        v
 [ALERT RESOLVED] -----> Dispatch Webhooks & Integrations
        |                  (Status: resolved)
        v
  [Healthy Pool]
```

## API Endpoints

| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Application health check |
| `POST` | `/config` | Update background checker config |
| `GET` | `/config` | View background checker config |
| `POST` | `/proxies` | Add or replace proxy URLs in the pool |
| `GET` | `/proxies` | View all proxies and pool failure rate |
| `GET` | `/proxies/{id}` | View details for a single proxy |
| `GET` | `/proxies/{id}/history` | View the last 100 check records for a proxy |
| `DELETE`| `/proxies` | Clear the proxy pool completely |
| `GET` | `/alerts` | View the complete alert history archive |
| `POST` | `/webhooks` | Register a new webhook receiver |
| `POST` | `/integrations` | Register a new Slack or Discord integration |
| `GET` | `/metrics` | View system metrics (total checks, deliveries) |

