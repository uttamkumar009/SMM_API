# SMM API

Go API for `api.smm.babiesiq.tech`, using MongoDB for persistence and Redis for rate limiting.

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https://github.com/uttamkumar009/SMM_API)

## Vercel deployment

The Vercel demo uses `api/index.js` as a native Node.js serverless function. The full Go service is in `backend/` and remains available for VPS or container deployment with MongoDB and Redis. Vercel functions are stateless, so use MongoDB Atlas and Redis Cloud before production.

The Vercel project must use the repository root (`.`), not `backend/`. The root `vercel.json` selects the Other framework preset so Vercel deploys `api/index.js`; the Go service is only for VPS or container deployment.

## Provider API

The Go service proxies the provider API from the server side. Configure `SMM_PROVIDER_API_URL` and `SMM_PROVIDER_API_KEY` in the backend environment; never put the provider key in frontend code. The frontend reads the live catalog from `GET /api/services`, and orders are submitted through the backend.

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

## Service catalog API

The frontend loads the catalog from the backend:

```bash
curl http://localhost:8080/api/categories
curl 'http://localhost:8080/api/services?category=Telegram%20Members'
```

Place the provider response (an object containing a `services` array) in a JSON file and set `SERVICES_CATALOG_FILE` to use the full catalog. Categories and rates are never hardcoded in the frontend, and order creation validates `serviceId`, quantity limits, and the server-side rate.

## Required Vercel variables

For the production service, configure `APP_ENV`, `JWT_SECRET`, `MONGO_URI`, `MONGO_DATABASE`, `REDIS_ADDRESS`, `REDIS_USERNAME`, `REDIS_PASSWORD`, `FRONTEND_ORIGIN`, and `RATE_LIMIT_PER_MINUTE` in Vercel Project Settings → Environment Variables. `REDIS_URL` can be used instead and takes precedence.

## Production

Put this service behind Nginx or a managed load balancer, expose only HTTPS, and use Cloudflare/WAF or a VPS firewall for volumetric DDoS protection. Application rate limiting is included, but it is not a substitute for edge-level DDoS mitigation.
