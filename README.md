# SMM API

Go API for `api.smm.babiesiq.tech`, using MongoDB for persistence and Redis for rate limiting.

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https://github.com/uttamkumar009/SMM_API)

## Vercel deployment

The `api/index.go` file is a Vercel serverless demo entrypoint. Click the button above or import this repository in Vercel, then set the project domain to `api.smm.babiesiq.tech`.

For production wallet/order persistence, use the regular Go service (`main.go`) with MongoDB and Redis. Vercel functions are stateless, so use MongoDB Atlas and Redis Cloud and move the handlers from `api/index.go` to the persistent store-backed API before production.

## Local setup

```bash
cp .env.example .env
# Set a strong JWT_SECRET and your MongoDB/Redis URLs.
go mod tidy
go test ./...
go run .
```

Default demo login is configured by `DEMO_EMAIL` and `DEMO_PASSWORD`. Change both before production.

## Required Vercel variables

For the production service, configure `APP_ENV`, `JWT_SECRET`, `MONGO_URI`, `MONGO_DATABASE`, `REDIS_URL`, `FRONTEND_ORIGIN`, and `RATE_LIMIT_PER_MINUTE` in Vercel Project Settings → Environment Variables.

## Production

Put this service behind Nginx or a managed load balancer, expose only HTTPS, and use Cloudflare/WAF or a VPS firewall for volumetric DDoS protection. Application rate limiting is included, but it is not a substitute for edge-level DDoS mitigation.
