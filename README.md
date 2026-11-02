# SMM API

Go API for `api.smm.babiesiq.tech`, using MongoDB for persistence and Redis for rate limiting.

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https://github.com/uttamkumar009/SMM_API)

## Vercel deployment

The Vercel demo uses `api/index.js` as a native Node.js serverless function. The full Go service is in `backend/` and remains available for VPS or container deployment with MongoDB and Redis. Vercel functions are stateless, so use MongoDB Atlas and Redis Cloud before production.

## Local setup

```bash
cp .env.example .env
# Set a strong JWT_SECRET and your MongoDB/Redis URLs.
cd backend
go mod tidy
go test ./...
go run .
```

Default demo login is configured by `DEMO_EMAIL` and `DEMO_PASSWORD`. Change both before production.

## Required Vercel variables

For the production service, configure `APP_ENV`, `JWT_SECRET`, `MONGO_URI`, `MONGO_DATABASE`, `REDIS_URL`, `FRONTEND_ORIGIN`, and `RATE_LIMIT_PER_MINUTE` in Vercel Project Settings → Environment Variables.

## Production

Put this service behind Nginx or a managed load balancer, expose only HTTPS, and use Cloudflare/WAF or a VPS firewall for volumetric DDoS protection. Application rate limiting is included, but it is not a substitute for edge-level DDoS mitigation.
