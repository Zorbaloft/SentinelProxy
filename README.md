# Sentinel Proxy

A high-performance reverse proxy + security gateway that sits in front of upstream services, logs complete request/response transactions to MongoDB, and provides a web dashboard for traffic monitoring and security management.

## Architecture

Sentinel consists of 4 main services:

1. **sentinel-engine** (Go) - Reverse proxy with guard, rate limiting, and async logging
2. **sentinel-api** (Go) - REST API for dashboard operations
3. **sentinel-brain** (Python) - Rule evaluator and AI heuristics
4. **sentinel-dashboard** (Next.js) - Web UI for logs, rules, incidents, and IP actions

## Quick Start

1. Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
# Edit .env with your settings
```

2. Start all services:

```bash
docker-compose up -d
```

3. Access the dashboard at http://localhost:3000

4. Test the proxy:

```bash
# Basic test
curl http://localhost:9090/

# Test routes that require specific Host headers (e.g., Symfony routes)
# For API routes (api.3cket.local):
curl -H "Host: api.3cket.local" http://localhost:9090/promoters

# For app routes (app.3cket.local):
curl -H "Host: app.3cket.local" http://localhost:9090/promoters
```

**Note:** Sentinel preserves the incoming Host header if it matches the domain pattern (`*.3cket.local` or `3cket.local`). This allows Symfony routes that depend on the Host header to work correctly. If no Host header is provided or it doesn't match the pattern, Sentinel uses the default `UPSTREAM_HOST` value.

## Configuration

### Environment Variables

- `UPSTREAM_URL` - Target upstream URL (default: http://nginx:....)
- `LOG_TTL_DAYS` - MongoDB log retention in days (default: 4)
- `LOG_BODY_MAX_BYTES` - Maximum body size to log (default: 1048576 = 1MB)
- `SENTINEL_ADMIN_TOKEN` - Admin token for API access (default: changeme)
- `POLL_INTERVAL_SEC` - Brain polling interval (default: 5)
- `AI_AUTOBLOCK` - Enable AI auto-blocking (default: false)

## Features

### Request Pipeline

1. **Guard** - Checks Redis blocklist/redirect lists
2. **Rate Counters** - Increments Redis counters for rule evaluation
3. **Proxy** - Forwards requests to upstream
4. **Logger** - Async MongoDB logging (non-blocking)

### IP Detection

Client IP is determined in this order:
1. `CF-Connecting-IP` (Cloudflare)
2. First IP in `X-Forwarded-For`
3. `X-Real-IP`
4. Remote address

### Logging

- Complete request/response logging to MongoDB
- Body truncation for large payloads (>1MB)
- Header redaction (Authorization, Cookie, Set-Cookie)
- Automatic TTL cleanup via MongoDB index

### Rules Engine

Rules support AND conditions:
- Path matching (exact/prefix/regex)
- HTTP method
- User agent (substring/regex)
- Status code/class/range
- Threshold and time window

### Manual Actions

- Block/unblock IPs
- Redirect IPs to custom URLs
- View and manage incidents

## API Endpoints

All endpoints require `X-Sentinel-Admin-Token` header.

### Logs
- `GET /logs` - Query logs with filters (ip, path, status, from, to, limit, cursor)

### Incidents
- `GET /incidents?status=open|closed` - List incidents
- `POST /incidents/:id/close` - Close incident

### Rules
- `GET /rules` - List all rules
- `POST /rules` - Create rule
- `PUT /rules/:id` - Update rule
- `DELETE /rules/:id` - Delete rule

### Actions
- `POST /block` - Block IP `{ip, ttlSec, reason}`
- `POST /unblock` - Unblock IP `{ip}`
- `POST /redirect` - Redirect IP `{ip, targetUrl, ttlSec, reason}`
- `POST /unredirect` - Remove redirect `{ip}`

## Example Rule

```json
{
  "name": "Login brute force",
  "enabled": true,
  "conditions": {
    "path": { "type": "regex", "value": "^/login$" },
    "method": "POST",
    "userAgent": { "type": "regex", "value": "curl|bot" },
    "status": { "type": "range", "min": 400, "max": 499 }
  },
  "threshold": 4,
  "windowSec": 60,
  "action": {
    "type": "block",
    "ttlSec": 3600,
    "reason": "rule:login_bruteforce"
  }
}
```

## Testing

### Acceptance Tests

1. **Proxy Test**
   ```bash
   curl http://localhost:9090/
   ```
   Should return content from upstream.

2. **MongoDB Logging**
   ```bash
   curl http://localhost:9090/test
   # Check MongoDB logs collection
   docker exec sentinel-mongo mongosh sentinel --eval "db.logs.find().limit(1).pretty()"
   ```

3. **Blocklist Test**
   ```bash
   # Block IP via API
   curl -X POST http://localhost:8090/block \
     -H "X-Sentinel-Admin-Token: changeme" \
     -H "Content-Type: application/json" \
     -d '{"ip":"192.168.1.100","ttlSec":3600,"reason":"test"}'
   
   # Test blocked (use the blocked IP)
   curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:9090/
   # Should return 403
   ```

4. **Redirect Test**
   ```bash
   curl -X POST http://localhost:8090/redirect \
     -H "X-Sentinel-Admin-Token: changeme" \
     -H "Content-Type: application/json" \
     -d '{"ip":"192.168.1.101","targetUrl":"https://example.com","ttlSec":3600,"reason":"test"}'
   
   curl -H "X-Forwarded-For: 192.168.1.101" http://localhost:9090/
   # Should return 302 redirect
   ```

## Development

### Building Services

```bash
# Engine
cd sentinel-engine && go build ./cmd/engine

# API
cd sentinel-api && go build ./cmd/api

# Brain
cd sentinel-brain && pip install -r requirements.txt

# Dashboard
cd sentinel-dashboard && npm install && npm run build
```

### Health Checks

- Engine: http://localhost:9090/health
- API: http://localhost:8090/health
- Dashboard: http://localhost:3000

## License

MIT
