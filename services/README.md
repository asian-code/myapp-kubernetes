# myHealth Services

This directory contains the microservices for the myHealth Oura Ring application.

## Services

### 1. oura-collector
A CronJob service that fetches data from the Oura Ring API and sends it to the data-processor.

**Features:**
- Fetches sleep, activity, and readiness data
- Runs every 5 minutes (configured in Kubernetes CronJob)
- Sends data to data-processor service

**Environment Variables:**
- `OURA_API_KEY`: Oura Ring API key
- `PROCESSOR_URL`: URL of the data-processor service
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

### 2. data-processor
An HTTP service that receives, transforms, and stores Oura Ring metrics in PostgreSQL.

**Features:**
- HTTP server on port 8080
- REST API for ingesting and querying metrics
- PostgreSQL database integration
- Prometheus metrics endpoint

**Endpoints:**
- `POST /api/v1/ingest` - Ingest metrics data
- `GET /api/v1/metrics/{type}` - Query metrics by type (sleep, activity, readiness)
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

**Environment Variables:**
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `LOG_LEVEL`: Logging level

### 3. api-service
A REST API service that provides authenticated access to health metrics.

**Features:**
- JWT-based authentication
- Dashboard endpoint with weekly summaries
- CORS support
- Prometheus metrics

**Endpoints:**
- `POST /auth/login` - Login and get JWT token
- `GET /api/v1/dashboard` - Get dashboard with latest metrics and weekly summary
- `GET /api/v1/sleep` - Get sleep metrics
- `GET /api/v1/activity` - Get activity metrics
- `GET /api/v1/readiness` - Get readiness metrics
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

**Environment Variables:**
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `JWT_SECRET`: JWT signing secret
- `LOG_LEVEL`: Logging level

## Shared Libraries

The `shared/` directory contains common packages used by all services:

- **logger**: Structured logging with logrus
- **database**: PostgreSQL connection pool management
- **secrets**: AWS Secrets Manager integration
- **metrics**: Prometheus metrics

## Development

### Prerequisites
- Go 1.21+
- Docker
- PostgreSQL (for local development)

### Build Services

```bash
# Build all services
cd services

# Build specific service
cd oura-collector
go build -o bin/oura-collector ./cmd

cd ../data-processor
go build -o bin/data-processor ./cmd

cd ../api-service
go build -o bin/api-service ./cmd
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests for specific service
cd oura-collector
go test ./...
```

### Build Docker Images

```bash
# From the services directory
docker build -t myhealth/oura-collector:latest -f oura-collector/Dockerfile .
docker build -t myhealth/data-processor:latest -f data-processor/Dockerfile .
docker build -t myhealth/api-service:latest -f api-service/Dockerfile .
```

### Local Development

1. Start PostgreSQL:
```bash
docker run -d \
  --name myhealth-postgres \
  -e POSTGRES_DB=myhealth \
  -e POSTGRES_USER=myhealth_user \
  -e POSTGRES_PASSWORD=secret \
  -p 5432:5432 \
  postgres:14
```

2. Set environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=myhealth_user
export DB_PASSWORD=secret
export DB_NAME=myhealth
export JWT_SECRET=your-secret-key
export OURA_API_KEY=your-oura-api-key
export PROCESSOR_URL=http://localhost:8080
```

3. Run services:
```bash
# Terminal 1: data-processor
cd data-processor
go run ./cmd

# Terminal 2: api-service
cd api-service
go run ./cmd

# Terminal 3: oura-collector (manual run)
cd oura-collector
go run ./cmd
```

## Database Schema

### sleep_metrics
- `id`: Serial primary key
- `oura_id`: Unique Oura metric ID
- `day`: Date of the metric
- `score`: Sleep score
- `duration`: Sleep duration in seconds
- `created_at`: Timestamp
- `updated_at`: Timestamp

### activity_metrics
- `id`: Serial primary key
- `oura_id`: Unique Oura metric ID
- `day`: Date of the metric
- `score`: Activity score
- `active_calories`: Active calories burned
- `steps`: Step count
- `medium_activity_minutes`: Medium activity duration
- `high_activity_minutes`: High activity duration
- `created_at`: Timestamp
- `updated_at`: Timestamp

### readiness_metrics
- `id`: Serial primary key
- `oura_id`: Unique Oura metric ID
- `day`: Date of the metric
- `score`: Readiness score
- `created_at`: Timestamp
- `updated_at`: Timestamp

## CI/CD

Docker images are automatically built and pushed to ECR via GitHub Actions when changes are pushed to the main branch.

See `.github/workflows/` for CI/CD pipeline configuration.

## Monitoring

All services expose Prometheus metrics at `/metrics`:
- `http_requests_total`: Total HTTP requests
- `http_request_duration_seconds`: HTTP request duration
- `db_query_duration_seconds`: Database query duration

## License

MIT
