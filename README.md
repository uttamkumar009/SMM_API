# SMM API

Go API for `api.smm.babiesiq.tech`, using MongoDB for persistence and Redis for rate limiting.

## Local setup

```bash
cp .env.example .env
# Set a strong JWT_SECRET and your MongoDB/Redis URLs.
go mod tidy
go test ./...
go run .
```

Default demo login is configured by `DEMO_EMAIL` and `DEMO_PASSWORD`. Change both before production.

## Production

Put this service behind Nginx or a managed load balancer, expose only HTTPS, and use Cloudflare/WAF or a VPS firewall for volumetric DDoS protection. Application rate limiting is included, but it is not a substitute for edge-level DDoS mitigation.
