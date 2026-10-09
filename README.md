# Takō Cloud Verification Mini App 🐙

Lightweight Go web application designed for end-to-end verification of **Takō Cloud PaaS**:
- Ingress domain routing & Let's Encrypt TLS (testing `demo.gettako.dev`)
- Container environment variables injection
- Terminal shell & diagnostic CLI tools inside containers
- Real-time system telemetry and health endpoints

## Endpoints

- `GET /` - Interactive Dark Mode Web Dashboard
- `GET /health` - Health check endpoint returning `200 OK`
- `GET /api/info` - Host header, client IP, runtime specs, and uptime
- `GET /api/env` - JSON array of injected environment variables
- `GET /api/echo?msg=test` - Echo diagnostic endpoint

## Running locally

```bash
go run main.go
```
Or via Docker:
```bash
docker build -t tako-demo-verify .
docker run -p 8080:8080 -e APP_ENV=production tako-demo-verify
```
