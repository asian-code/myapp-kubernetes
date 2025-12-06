# myHealth - Oura Ring Kubernetes Microservice Application
## Project Plan & Architecture

**Tech Stack:** EKS, Terraform, Golang, Helm, ArgoCD, Prometheus, Grafana, Istio, GitHub Actions

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Technology Stack](#technology-stack)
3. [Project Structure](#project-structure)
4. [Infrastructure Components](#infrastructure-components)
5. [Microservices Design](#microservices-design)
6. [Helm Chart Strategy](#helm-chart-strategy)
7. [CI/CD Pipeline](#cicd-pipeline)
8. [Monitoring & Observability](#monitoring--observability)
9. [Security & Authentication](#security--authentication)
10. [Implementation Roadmap](#implementation-roadmap)
11. [Prerequisites](#prerequisites)

---

## Architecture Overview

### High-Level Architecture Diagram
```
┌─────────────────────────────────────────────────────────────────────┐
│                           AWS Cloud                                  │
│                                                                      │
│  ┌──────────────┐         ┌─────────────────────────────────┐      │
│  │  CloudFront  │────────▶│  S3 (Frontend Static Assets)    │      │
│  └──────────────┘         └─────────────────────────────────┘      │
│         │                                                            │
│         │                                                            │
│  ┌──────▼──────────┐                                                │
│  │  API Gateway    │                                                │
│  └──────┬──────────┘                                                │
│         │                                                            │
│  ┌──────▼─────────────────────────────────────────────────────┐    │
│  │              EKS Cluster (VPC)                              │    │
│  │                                                              │    │
│  │  ┌────────────────────────────────────────────────────┐    │    │
│  │  │        Istio Service Mesh                           │    │    │
│  │  │                                                      │    │    │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐ │    │    │
│  │  │  │ oura-        │  │ data-        │  │ api-     │ │    │    │
│  │  │  │ collector    │─▶│ processor    │─▶│ service  │ │    │    │
│  │  │  │ (CronJob)    │  │              │  │          │ │    │    │
│  │  │  └──────────────┘  └──────────────┘  └──────────┘ │    │    │
│  │  │                                                      │    │    │
│  │  │  ┌──────────────┐  ┌──────────────┐                │    │    │
│  │  │  │ Prometheus   │  │ Grafana      │                │    │    │
│  │  │  └──────────────┘  └──────────────┘                │    │    │
│  │  └────────────────────────────────────────────────────┘    │    │
│  │                                                              │    │
│  │  ┌─────────────┐    ┌──────────────┐                       │    │
│  │  │ ArgoCD      │    │ PostgreSQL   │                       │    │
│  │  │             │    │ (RDS)        │                       │    │
│  │  └─────────────┘    └──────────────┘                       │    │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌────────────────────┐  ┌────────────────────┐                    │
│  │ AWS Secrets Manager│  │  ECR (Container    │                    │
│  │                    │  │  Registry)         │                    │
│  └────────────────────┘  └────────────────────┘                    │
└─────────────────────────────────────────────────────────────────────┘

External:
┌──────────────┐         ┌──────────────┐
│ Oura Ring API│         │ GitHub       │
└──────────────┘         │ (Source Code)│
                         └──────────────┘
```

### Data Flow
1. **Scheduled Data Collection**: CronJob triggers `oura-collector` every 5 minutes
2. **Data Ingestion**: Collector fetches data from Oura API → sends to `data-processor`
3. **Data Processing**: Processor transforms, validates → stores in PostgreSQL RDS
4. **API Access**: Frontend/Users → CloudFront → API Gateway → Istio Ingress → `api-service`
5. **Metrics**: All services expose Prometheus metrics → Prometheus scrapes → Grafana visualizes
6. **GitOps**: Git push → GitHub Actions → ECR → ArgoCD detects changes → deploys to EKS

---

## Technology Stack

### Infrastructure Layer
| Component | Technology | Purpose |
|-----------|------------|---------|
| Container Orchestration | AWS EKS (Kubernetes 1.34+) | Manages containerized workloads |
| Infrastructure as Code | Terraform (v1.13+) | Provisions AWS resources |
| Service Mesh | Istio (1.23+) | Traffic management, security, observability |
| API Gateway | AWS API Gateway | External API entry point, rate limiting, auth |
| CDN | AWS CloudFront | Static content delivery |
| Storage | AWS S3 | Frontend static assets |
| Database | AWS RDS PostgreSQL (14+) | Persistent data storage |
| Secrets Management | AWS Secrets Manager | API keys, DB credentials |
| Container Registry | AWS ECR | Docker image storage |

### Application Layer
| Component | Technology | Purpose |
|-----------|------------|---------|
| Microservices | Go (1.21+) | Backend services |
| Package Manager | Helm (v3.13+) | Kubernetes application packaging |
| GitOps | ArgoCD (2.9+) | Continuous deployment |
| Monitoring | Prometheus + Grafana | Metrics & dashboards |
| CI/CD | GitHub Actions | Build, test, deploy automation |

### Development Tools
- **Docker**: Containerization
- **kubectl**: Kubernetes CLI
- **helm**: Helm CLI
- **aws-cli**: AWS operations
- **terraform**: Infrastructure management
- **istioctl**: Istio management

---

## Project Structure

```
myapp-kubernetes/
├── .github/
│   └── workflows/
│       ├── terraform-plan.yml           # Infrastructure validation
│       ├── terraform-apply.yml          # Infrastructure deployment
│       ├── build-oura-collector.yml     # Build & push collector service
│       ├── build-data-processor.yml     # Build & push processor service
│       ├── build-api-service.yml        # Build & push API service
│       ├── helm-lint.yml                # Helm chart validation
│       └── deploy-argocd.yml            # ArgoCD sync trigger
│
├── terraform/
│   ├── environments/
│   │   ├── dev/
│   │   │   ├── main.tf
│   │   │   ├── terraform.tfvars
│   │   │   └── backend.tf
│   │   ├── staging/
│   │   └── prod/
│   ├── modules/
│   │   ├── eks/
│   │   │   ├── main.tf                  # EKS cluster configuration
│   │   │   ├── variables.tf
│   │   │   ├── outputs.tf
│   │   │   ├── node-groups.tf
│   │   │   ├── irsa.tf                  # IAM Roles for Service Accounts
│   │   │   └── addons.tf                # EKS add-ons (EBS CSI, etc.)
│   │   ├── networking/
│   │   │   ├── main.tf                  # VPC, subnets, NAT
│   │   │   ├── security-groups.tf
│   │   │   └── outputs.tf
│   │   ├── rds/
│   │   │   ├── main.tf                  # PostgreSQL RDS
│   │   │   ├── variables.tf
│   │   │   └── outputs.tf
│   │   ├── api-gateway/
│   │   │   ├── main.tf                  # API Gateway + VPC Link
│   │   │   ├── authorizer.tf            # Lambda authorizer
│   │   │   └── outputs.tf
│   │   ├── secrets-manager/
│   │   │   ├── main.tf                  # Secrets for Oura API, DB
│   │   │   └── outputs.tf
│   │   ├── ecr/
│   │   │   ├── main.tf                  # ECR repositories
│   │   │   └── outputs.tf
│   │   ├── cloudfront/
│   │   │   ├── main.tf                  # CloudFront + S3
│   │   │   ├── s3.tf
│   │   │   └── outputs.tf
│   │   └── istio/
│   │       ├── main.tf                  # Istio installation via Helm
│   │       ├── gateway.tf
│   │       └── outputs.tf
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── backend.tf                       # S3 backend for state
│   └── versions.tf
│
├── services/
│   ├── oura-collector/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── client/                  # Oura API client
│   │   │   │   ├── oura.go
│   │   │   │   └── oura_test.go
│   │   │   ├── processor/               # Data transformation
│   │   │   │   └── processor.go
│   │   │   └── config/
│   │   │       └── config.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── README.md
│   │
│   ├── data-processor/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/                 # HTTP/gRPC handlers
│   │   │   │   └── handler.go
│   │   │   ├── repository/              # Database layer
│   │   │   │   ├── postgres.go
│   │   │   │   └── repository.go
│   │   │   ├── service/                 # Business logic
│   │   │   │   └── service.go
│   │   │   ├── models/
│   │   │   │   └── models.go
│   │   │   └── metrics/                 # Prometheus metrics
│   │   │       └── metrics.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── README.md
│   │
│   ├── api-service/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/                 # REST API handlers
│   │   │   │   ├── health.go
│   │   │   │   ├── stats.go
│   │   │   │   └── middleware.go
│   │   │   ├── repository/
│   │   │   │   └── repository.go
│   │   │   ├── service/
│   │   │   │   └── service.go
│   │   │   ├── auth/                    # JWT validation
│   │   │   │   └── auth.go
│   │   │   └── metrics/
│   │   │       └── metrics.go
│   │   ├── api/
│   │   │   └── openapi.yaml             # OpenAPI spec
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── README.md
│   │
│   └── shared/                          # Shared Go packages
│       ├── logger/
│       ├── database/
│       ├── secrets/
│       └── metrics/
│
├── helm/
│   ├── myhealth/
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   ├── values-dev.yaml
│   │   ├── values-staging.yaml
│   │   ├── values-prod.yaml
│   │   ├── charts/                      # Dependency charts
│   │   │   ├── prometheus/              # (pulled as dependency)
│   │   │   └── grafana/                 # (pulled as dependency)
│   │   ├── templates/
│   │   │   ├── _helpers.tpl
│   │   │   ├── NOTES.txt
│   │   │   ├── serviceaccount.yaml
│   │   │   ├── configmap.yaml
│   │   │   ├── secrets.yaml             # External secrets operator
│   │   │   │
│   │   │   ├── oura-collector/
│   │   │   │   ├── cronjob.yaml
│   │   │   │   └── servicemonitor.yaml
│   │   │   │
│   │   │   ├── data-processor/
│   │   │   │   ├── deployment.yaml
│   │   │   │   ├── service.yaml
│   │   │   │   ├── hpa.yaml
│   │   │   │   └── servicemonitor.yaml
│   │   │   │
│   │   │   ├── api-service/
│   │   │   │   ├── deployment.yaml
│   │   │   │   ├── service.yaml
│   │   │   │   ├── hpa.yaml
│   │   │   │   ├── virtualservice.yaml   # Istio routing
│   │   │   │   ├── destinationrule.yaml
│   │   │   │   └── servicemonitor.yaml
│   │   │   │
│   │   │   ├── istio/
│   │   │   │   ├── gateway.yaml
│   │   │   │   ├── authorizationpolicy.yaml
│   │   │   │   └── peerauthentication.yaml
│   │   │   │
│   │   │   ├── prometheus/
│   │   │   │   ├── prometheus.yaml
│   │   │   │   └── serviceaccount.yaml
│   │   │   │
│   │   │   └── grafana/
│   │   │       ├── deployment.yaml
│   │   │       ├── service.yaml
│   │   │       ├── configmap-dashboards.yaml
│   │   │       └── configmap-datasources.yaml
│   │   │
│   │   └── dashboards/                  # Grafana JSON dashboards
│   │       ├── oura-overview.json
│   │       ├── sleep-metrics.json
│   │       ├── activity-metrics.json
│   │       └── service-health.json
│   │
│   └── argocd/
│       ├── application.yaml             # ArgoCD Application CRD
│       ├── appproject.yaml              # ArgoCD AppProject
│       └── values-override.yaml
│
├── scripts/
│   ├── setup-cluster.sh                 # Initial cluster setup
│   ├── install-argocd.sh                # ArgoCD installation
│   ├── install-istio.sh                 # Istio installation
│   ├── create-secrets.sh                # AWS Secrets Manager setup
│   ├── db-migrate.sh                    # Database migrations
│   └── local-dev.sh                     # Local development setup
│
├── frontend/                            # (Optional - React/Vue for custom UI)
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── Dockerfile
│
├── database/
│   ├── migrations/
│   │   ├── 001_initial_schema.sql
│   │   ├── 002_add_sleep_metrics.sql
│   │   └── 003_add_activity_metrics.sql
│   └── seeds/
│       └── dev_data.sql
│
├── docs/
│   ├── architecture.md
│   ├── api-documentation.md
│   ├── deployment-guide.md
│   ├── monitoring-guide.md
│   └── troubleshooting.md
│
├── .gitignore
├── .dockerignore
├── Makefile                             # Common commands
├── README.md
└── PROJECT_PLAN.md                      # This file
```

---

## Infrastructure Components

### 1. AWS EKS Cluster
**Purpose**: Kubernetes cluster to run microservices

**Terraform Module**: `terraform/modules/eks/`

**Key Features**:
- Multi-AZ deployment (3 availability zones)
- Managed node groups with auto-scaling
- IRSA (IAM Roles for Service Accounts) for fine-grained permissions
- EKS add-ons: VPC CNI, CoreDNS, kube-proxy, EBS CSI driver
- Cluster autoscaler
- Version: 1.28+

**Configuration**:
```hcl
# Minimum 2 nodes, max 10 nodes
# Instance types: t3.medium (dev), t3.large (prod)
# Spot instances for cost optimization (non-prod)
```

### 2. VPC & Networking
**Purpose**: Isolated network for EKS cluster

**Terraform Module**: `terraform/modules/networking/`

**Components**:
- VPC with CIDR: `10.0.0.0/16`
- Public subnets (3): For load balancers, NAT gateways
- Private subnets (3): For EKS nodes, RDS
- NAT Gateways: High availability (one per AZ)
- Internet Gateway
- Route tables with proper routing
- VPC endpoints for AWS services (S3, ECR, Secrets Manager)

### 3. AWS API Gateway
**Purpose**: Public API entry point with authentication

**Terraform Module**: `terraform/modules/api-gateway/`

**Features**:
- REST API with custom domain
- VPC Link to connect to EKS internal load balancer
- Lambda Authorizer for JWT validation
- Request/response validation
- Rate limiting (100 req/min per client)
- API key management
- CloudWatch logging
- WAF integration (optional)

### 4. AWS RDS PostgreSQL
**Purpose**: Persistent data storage for Oura metrics

**Terraform Module**: `terraform/modules/rds/`

**Configuration**:
- Engine: PostgreSQL 14.x
- Instance class: db.t3.medium (dev), db.r6g.large (prod)
- Multi-AZ for high availability (prod)
- Automated backups (7-day retention)
- Encryption at rest (KMS)
- Private subnet group (no public access)
- Enhanced monitoring enabled

**Schema**:
```sql
Tables:
- users (id, email, oura_user_id, created_at)
- sleep_metrics (id, user_id, date, score, duration, efficiency, ...)
- activity_metrics (id, user_id, date, steps, calories, distance, ...)
- readiness_metrics (id, user_id, date, score, hrv, resting_hr, ...)
- raw_data (id, user_id, data_type, payload, created_at)
```

### 5. AWS Secrets Manager
**Purpose**: Secure storage for sensitive credentials

**Terraform Module**: `terraform/modules/secrets-manager/`

**Secrets**:
- `myhealth/oura-api-key`: Oura Ring API credentials
- `myhealth/db-credentials`: PostgreSQL username/password
- `myhealth/jwt-secret`: JWT signing key for API authentication
- `myhealth/grafana-admin`: Grafana admin credentials

**Access**:
- IRSA policies for K8s pods
- External Secrets Operator to sync to K8s Secrets

### 6. AWS ECR
**Purpose**: Docker container registry

**Terraform Module**: `terraform/modules/ecr/`

**Repositories**:
- `myhealth/oura-collector`
- `myhealth/data-processor`
- `myhealth/api-service`

**Features**:
- Image scanning on push
- Lifecycle policies (keep last 10 images)
- Encryption at rest
- Cross-region replication (optional)

### 7. CloudFront + S3
**Purpose**: Frontend static asset delivery

**Terraform Module**: `terraform/modules/cloudfront/`

**Configuration**:
- S3 bucket with website hosting
- CloudFront distribution
- Origin Access Identity (OAI)
- SSL/TLS certificate (ACM)
- Custom domain name
- Cache policies optimized for SPA

### 8. Istio Service Mesh
**Purpose**: Traffic management, security, observability

**Installation**: Helm chart via Terraform module

**Components**:
- Istio Control Plane (istiod)
- Istio Ingress Gateway
- Istio Egress Gateway (for Oura API calls)

**Features**:
- Mutual TLS (mTLS) between services
- Traffic routing and splitting
- Circuit breaking and retries
- Distributed tracing (Jaeger integration)
- Request metrics

---

## Microservices Design

### 1. oura-collector
**Purpose**: Scheduled job to fetch data from Oura Ring API

**Type**: Kubernetes CronJob (runs every 5 minutes)

**Responsibilities**:
- Authenticate with Oura API using OAuth2
- Fetch daily summaries, sleep, activity, readiness data
- Send data to `data-processor` via HTTP/gRPC
- Handle API rate limits and retries
- Log collection status

**API Endpoints** (Internal):
- Health check: `GET /health`
- Metrics: `GET /metrics` (Prometheus)

**Environment Variables**:
```yaml
OURA_API_KEY: (from AWS Secrets Manager)
PROCESSOR_URL: http://data-processor:8080
LOG_LEVEL: info
```

**Dependencies**:
- Oura Ring API (external)
- data-processor service

### 2. data-processor
**Purpose**: Process, validate, and store Oura metrics

**Type**: Kubernetes Deployment (HTTP service)

**Responsibilities**:
- Receive data from collector
- Validate and transform data
- Store in PostgreSQL RDS
- Provide internal API for queries
- Cache frequently accessed data (Redis - optional)

**API Endpoints** (Internal):
```
POST   /api/v1/ingest         - Receive data from collector
GET    /api/v1/metrics/{type} - Query stored metrics
GET    /health                - Health check
GET    /metrics               - Prometheus metrics
```

**Database Access**:
- Connection pooling (pgx)
- Automatic migrations on startup
- Read replicas support (future)

**Scaling**:
- Horizontal Pod Autoscaler (2-5 replicas)
- CPU-based scaling (70% threshold)

### 3. api-service
**Purpose**: Public-facing REST API for frontend/users

**Type**: Kubernetes Deployment (HTTP service)

**Responsibilities**:
- Serve aggregated Oura metrics
- User authentication (JWT)
- Rate limiting per user
- API documentation (Swagger/OpenAPI)
- CORS handling

**API Endpoints** (Public via API Gateway):
```
POST   /auth/login                    - User authentication
GET    /api/v1/dashboard              - Dashboard summary
GET    /api/v1/sleep                  - Sleep metrics
GET    /api/v1/activity               - Activity metrics
GET    /api/v1/readiness              - Readiness metrics
GET    /api/v1/trends/{metric}        - Historical trends
GET    /health                        - Health check
GET    /metrics                       - Prometheus metrics
```

**Authentication Flow**:
1. User logs in → JWT token issued
2. API Gateway validates JWT with Lambda Authorizer
3. Istio VirtualService routes to api-service
4. Service verifies token and processes request

**Scaling**:
- Horizontal Pod Autoscaler (3-10 replicas)
- Request-based scaling (100 req/sec threshold)

---

## Helm Chart Strategy

### Chart Structure
**Single Helm Chart**: `myhealth` (umbrella chart)

### Dependencies (Chart.yaml)
```yaml
dependencies:
  - name: prometheus
    version: "25.8.0"
    repository: "https://prometheus-community.github.io/helm-charts"
    condition: prometheus.enabled
  
  - name: grafana
    version: "7.0.0"
    repository: "https://grafana.github.io/helm-charts"
    condition: grafana.enabled
  
  - name: external-secrets
    version: "0.9.0"
    repository: "https://charts.external-secrets.io"
    condition: externalSecrets.enabled
```

### Values Structure (values.yaml)
```yaml
global:
  environment: dev
  region: us-east-1
  domain: myhealth.example.com

ouraCollector:
  enabled: true
  image:
    repository: <account-id>.dkr.ecr.us-east-1.amazonaws.com/myhealth/oura-collector
    tag: latest
  schedule: "*/5 * * * *"  # Every 5 minutes
  resources:
    requests:
      cpu: 100m
      memory: 128Mi

dataProcessor:
  enabled: true
  image:
    repository: <account-id>.dkr.ecr.us-east-1.amazonaws.com/myhealth/data-processor
    tag: latest
  replicaCount: 2
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5

apiService:
  enabled: true
  image:
    repository: <account-id>.dkr.ecr.us-east-1.amazonaws.com/myhealth/api-service
    tag: latest
  replicaCount: 3
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10

database:
  host: myhealth-db.xxxxx.us-east-1.rds.amazonaws.com
  port: 5432
  name: myhealth

istio:
  enabled: true
  gateway:
    hosts:
      - api.myhealth.example.com

prometheus:
  enabled: true
  server:
    retention: 15d
  alertmanager:
    enabled: true

grafana:
  enabled: true
  adminPassword: (from secret)
  dashboardProviders:
    enabled: true
  datasources:
    enabled: true
```

### Environment-Specific Values
- `values-dev.yaml`: Lower resources, relaxed limits
- `values-staging.yaml`: Production-like setup
- `values-prod.yaml`: High availability, auto-scaling

---

## CI/CD Pipeline

### GitHub Actions Workflows

#### 1. Terraform Workflows

**terraform-plan.yml** (On PR)
```yaml
Triggers: Pull request to main
Steps:
  1. Checkout code
  2. Setup Terraform
  3. Terraform init
  4. Terraform fmt -check
  5. Terraform validate
  6. Terraform plan
  7. Comment plan output on PR
```

**terraform-apply.yml** (On merge to main)
```yaml
Triggers: Push to main (terraform/* changes)
Steps:
  1. Checkout code
  2. Setup Terraform
  3. Terraform init
  4. Terraform apply -auto-approve
  5. Export outputs (EKS cluster name, RDS endpoint, etc.)
  6. Update GitHub secrets/variables
```

#### 2. Microservice Build Workflows

**build-oura-collector.yml**
```yaml
Triggers: 
  - Push to main (services/oura-collector/* changes)
  - Manual workflow dispatch
  
Steps:
  1. Checkout code
  2. Setup Go 1.21
  3. Run tests (go test -v ./...)
  4. Run linter (golangci-lint)
  5. Build Docker image
  6. Scan image (Trivy/Grype)
  7. Tag image (git SHA + latest)
  8. Push to ECR
  9. Update Helm values with new image tag
  10. Commit changes (triggers ArgoCD)
```

**build-data-processor.yml** (Similar to above)

**build-api-service.yml** (Similar to above)

#### 3. Helm Chart Workflow

**helm-lint.yml** (On PR)
```yaml
Triggers: Pull request (helm/* changes)
Steps:
  1. Checkout code
  2. Setup Helm
  3. Helm lint
  4. Helm template (dry-run)
  5. kubeval validation
  6. Comment results on PR
```

#### 4. ArgoCD Deployment

**deploy-argocd.yml** (On merge to main)
```yaml
Triggers: Push to main (helm/* changes)
Steps:
  1. Checkout code
  2. Configure kubectl (EKS)
  3. ArgoCD sync (if auto-sync disabled)
  4. Wait for sync completion
  5. Health check
  6. Slack/Discord notification
```

### Deployment Flow
```
Code Push → GitHub Actions → Build & Test → Push to ECR → 
Update Helm Values → ArgoCD Detects Change → Sync to EKS → 
Health Checks → Notifications
```

---

## Monitoring & Observability

### 1. Prometheus
**Purpose**: Metrics collection and storage

**Configuration**:
- Scrape interval: 30s
- Retention: 15 days
- ServiceMonitor CRDs for automatic discovery
- Persistent volume: 50GB (EBS gp3)

**Metrics Collected**:
- Application metrics (custom business metrics)
- HTTP request rates, latencies, errors
- Database connection pool stats
- Oura API call success/failure rates
- Resource usage (CPU, memory, network)

**Alerting Rules**:
```yaml
- High error rate (> 5% for 5min)
- API latency > 1s (p95)
- Pod crash loop
- Database connection failures
- Oura API rate limit approaching
```

### 2. Grafana
**Purpose**: Visualization and dashboards

**Dashboards**:

**1. Oura Overview Dashboard**
- Sleep score trends (7d, 30d, 90d)
- Activity summary (steps, calories)
- Readiness score
- HRV trends

**2. Sleep Metrics Dashboard**
- Total sleep time
- Sleep efficiency
- Deep/REM/Light sleep breakdown
- Sleep latency
- Nighttime movement

**3. Activity Metrics Dashboard**
- Daily steps
- Active calories
- Training frequency/volume
- Recovery time

**4. Service Health Dashboard**
- Request rate per service
- Latency percentiles (p50, p95, p99)
- Error rates
- Pod status and restarts
- Database query performance

**5. Kubernetes Cluster Dashboard**
- Node CPU/Memory usage
- Pod resource consumption
- Network I/O
- PVC usage

### 3. Istio Observability
**Kiali**: Service mesh visualization (traffic flow, health)
**Jaeger**: Distributed tracing (request path across services)

### 4. Logging
**Solution**: AWS CloudWatch Logs or Fluentd → OpenSearch

**Log Aggregation**:
- Application logs (structured JSON)
- Istio access logs
- Audit logs
- Error tracking (Sentry integration - optional)

---

## Security & Authentication

### 1. Authentication Flow

**User Authentication** (JWT-based):
```
1. User → POST /auth/login (email, password)
2. API Service validates credentials
3. Issues JWT token (expiry: 24h)
4. User includes token in Authorization header
5. API Gateway Lambda Authorizer validates JWT
6. Request forwarded to EKS if valid
```

**Oura API Authentication** (OAuth2):
```
1. Admin completes OAuth2 flow (one-time)
2. Access token + refresh token stored in Secrets Manager
3. Collector uses access token for API calls
4. Automatically refreshes when expired
```

### 2. Network Security

**Security Groups**:
- EKS nodes: Allow only necessary ports (443, 10250, etc.)
- RDS: Only accessible from EKS security group
- API Gateway: Public (with WAF)

**Network Policies** (Kubernetes):
- Default deny all ingress/egress
- Explicit allow rules per service
- Namespace isolation

**Istio Security**:
- mTLS enforced between all services (STRICT mode)
- Authorization policies (role-based access)
- JWT validation at ingress gateway

### 3. Secrets Management

**AWS Secrets Manager** → **External Secrets Operator** → **Kubernetes Secrets**

**Rotation**:
- Database credentials: 90 days
- JWT signing keys: 180 days
- Oura API tokens: Auto-refresh

**Access Control**:
- IRSA policies (least privilege)
- Secrets encrypted at rest (KMS)
- Audit logging enabled

### 4. Image Security
- Scan all images before deployment (Trivy)
- Use minimal base images (distroless/alpine)
- Regular vulnerability patching
- Signed images (Cosign - optional)

### 5. RBAC
**Kubernetes RBAC**:
- Namespace-based isolation
- Service accounts with minimal permissions
- No cluster-admin for applications

**API Authorization**:
- Role-based access (admin, user)
- Scope-based permissions (read:sleep, write:profile)

---

## Implementation Roadmap

### Phase 1: Infrastructure Setup (Week 1-2)

**Sprint 1.1: AWS Foundation**
- [ ] Set up Terraform backend (S3 + DynamoDB)
- [ ] Create VPC and networking
- [ ] Provision EKS cluster
- [ ] Set up ECR repositories
- [ ] Configure AWS Secrets Manager
- [ ] Create RDS PostgreSQL instance

**Sprint 1.2: Kubernetes Add-ons**
- [ ] Install Istio service mesh
- [ ] Install External Secrets Operator
- [ ] Install Metrics Server
- [ ] Install Cluster Autoscaler
- [ ] Configure IRSA for workloads

**Sprint 1.3: GitOps Setup**
- [ ] Install ArgoCD
- [ ] Configure ArgoCD Application/AppProject
- [ ] Set up GitHub webhooks
- [ ] Test sync workflow

**Deliverables**:
- Fully functional EKS cluster
- ArgoCD operational
- All AWS resources provisioned

---

### Phase 2: Microservices Development (Week 3-5)

**Sprint 2.1: Shared Libraries**
- [ ] Create Go module structure
- [ ] Implement logger package
- [ ] Implement database connection package
- [ ] Implement secrets fetching package
- [ ] Implement Prometheus metrics package
- [ ] Write unit tests (80% coverage target)

**Sprint 2.2: oura-collector Service**
- [ ] Implement Oura API client (OAuth2)
- [ ] Create data fetching logic
- [ ] Implement retry mechanism
- [ ] Add Prometheus metrics
- [ ] Write Dockerfile (multi-stage build)
- [ ] Create unit tests + integration tests
- [ ] Document API usage

**Sprint 2.3: data-processor Service**
- [ ] Create HTTP/gRPC server
- [ ] Implement data validation logic
- [ ] Create PostgreSQL repository layer
- [ ] Add database migrations
- [ ] Implement caching (if needed)
- [ ] Add Prometheus metrics
- [ ] Write comprehensive tests
- [ ] Load testing

**Sprint 2.4: api-service**
- [ ] Create REST API server (Gin/Echo framework)
- [ ] Implement authentication middleware
- [ ] Create all API endpoints
- [ ] Add rate limiting
- [ ] Generate OpenAPI spec
- [ ] Add Prometheus metrics
- [ ] Write API integration tests
- [ ] Performance testing

**Deliverables**:
- Three microservices with 80%+ test coverage
- Docker images in ECR
- API documentation

---

### Phase 3: Helm Chart Development (Week 6)

**Sprint 3.1: Base Chart**
- [ ] Create Helm chart structure
- [ ] Define Chart.yaml with dependencies
- [ ] Create values.yaml schema
- [ ] Implement helper templates
- [ ] Create ConfigMaps and Secrets templates

**Sprint 3.2: Service Templates**
- [ ] CronJob for oura-collector
- [ ] Deployment for data-processor
- [ ] Deployment for api-service
- [ ] Services and ServiceMonitors
- [ ] HPA configurations
- [ ] PodDisruptionBudgets

**Sprint 3.3: Istio Configuration**
- [ ] Gateway template
- [ ] VirtualService templates
- [ ] DestinationRule templates
- [ ] AuthorizationPolicy templates
- [ ] PeerAuthentication (mTLS)

**Sprint 3.4: Observability**
- [ ] Prometheus configuration
- [ ] Grafana dashboards (as ConfigMaps)
- [ ] ServiceMonitor CRDs
- [ ] AlertManager rules
- [ ] Test chart installation

**Deliverables**:
- Production-ready Helm chart
- Multiple environment configurations
- Validated on dev cluster

---

### Phase 4: CI/CD Pipeline (Week 7)

**Sprint 4.1: Terraform Workflows**
- [ ] Create terraform-plan.yml
- [ ] Create terraform-apply.yml
- [ ] Set up AWS credentials in GitHub
- [ ] Test workflows

**Sprint 4.2: Build Workflows**
- [ ] Create build workflows for each service
- [ ] Integrate Go testing
- [ ] Add golangci-lint
- [ ] Add Trivy image scanning
- [ ] Configure ECR push

**Sprint 4.3: Deployment Workflows**
- [ ] Create helm-lint workflow
- [ ] Create ArgoCD sync workflow
- [ ] Add health checks
- [ ] Configure notifications (Slack/Discord)
- [ ] End-to-end test

**Deliverables**:
- Fully automated CI/CD pipeline
- Zero-touch deployments to dev/staging

---

### Phase 5: Monitoring & Dashboards (Week 8)

**Sprint 5.1: Prometheus Setup**
- [ ] Configure scrape configs
- [ ] Create alerting rules
- [ ] Set up AlertManager
- [ ] Configure notification channels
- [ ] Test alerts

**Sprint 5.2: Grafana Dashboards**
- [ ] Create Oura metrics dashboards
- [ ] Create service health dashboards
- [ ] Create Kubernetes cluster dashboard
- [ ] Import Istio dashboards
- [ ] Configure data sources

**Sprint 5.3: Logging & Tracing**
- [ ] Set up CloudWatch Logs
- [ ] Configure Fluentd (optional)
- [ ] Enable Jaeger tracing
- [ ] Test distributed tracing

**Deliverables**:
- Complete observability stack
- Pre-configured dashboards
- Working alerting

---

### Phase 6: API Gateway & Frontend (Week 9-10)

**Sprint 6.1: API Gateway**
- [ ] Provision API Gateway via Terraform
- [ ] Create VPC Link to EKS
- [ ] Configure routes
- [ ] Create Lambda Authorizer
- [ ] Set up rate limiting
- [ ] Configure custom domain
- [ ] Test end-to-end

**Sprint 6.2: CloudFront + S3**
- [ ] Provision S3 bucket
- [ ] Configure CloudFront distribution
- [ ] Set up SSL certificate
- [ ] Configure custom domain
- [ ] Test static hosting

**Sprint 6.3: Frontend (Optional)**
- [ ] Create React/Vue app
- [ ] Integrate with API
- [ ] Build dashboard UI
- [ ] Add authentication flow
- [ ] Deploy to S3/CloudFront

**Deliverables**:
- Public API accessible via API Gateway
- Frontend deployed (if implemented)

---

### Phase 7: Security Hardening (Week 11)

**Sprint 7.1: Security Audit**
- [ ] Enable Pod Security Standards
- [ ] Review RBAC policies
- [ ] Audit network policies
- [ ] Review Istio AuthZ policies
- [ ] Scan for vulnerabilities

**Sprint 7.2: Compliance**
- [ ] Enable AWS Config
- [ ] Set up CloudTrail
- [ ] Configure GuardDuty
- [ ] Enable VPC Flow Logs
- [ ] Document security controls

**Sprint 7.3: DR & Backup**
- [ ] Configure RDS automated backups
- [ ] Set up Velero for K8s backups
- [ ] Document disaster recovery procedures
- [ ] Test backup restoration

**Deliverables**:
- Security hardened infrastructure
- Backup and recovery procedures

---

### Phase 8: Testing & Launch (Week 12)

**Sprint 8.1: Load Testing**
- [ ] Create load test scenarios (k6/Locust)
- [ ] Test API endpoints under load
- [ ] Verify auto-scaling
- [ ] Identify bottlenecks
- [ ] Optimize performance

**Sprint 8.2: End-to-End Testing**
- [ ] Test complete data flow (Oura → API)
- [ ] Verify authentication
- [ ] Test failure scenarios
- [ ] Verify monitoring/alerts
- [ ] Chaos engineering (optional)

**Sprint 8.3: Documentation**
- [ ] Complete README
- [ ] API documentation
- [ ] Deployment guide
- [ ] Runbook for operations
- [ ] Architecture diagrams

**Sprint 8.4: Production Launch**
- [ ] Deploy to production
- [ ] Smoke tests
- [ ] Monitor for 24h
- [ ] Handoff to operations

**Deliverables**:
- Production-ready application
- Complete documentation
- Operational runbooks

---

## Prerequisites

### Required Tools
```bash
# Infrastructure
terraform >= 1.6.0
aws-cli >= 2.0

# Kubernetes
kubectl >= 1.28
helm >= 3.13
istioctl >= 1.20
argocd >= 2.9

# Development
go >= 1.21
docker >= 24.0
git >= 2.40

# Optional
k9s              # Kubernetes TUI
kubectx/kubens   # Context switching
```

### AWS Resources
- AWS Account with sufficient quotas
- IAM user/role with admin permissions
- Route53 hosted zone (for custom domains)
- ACM certificate (for HTTPS)

### External Services
- GitHub account (for repository + Actions)
- Oura Ring account + API access
- Domain name (optional)

### Knowledge Requirements
- Kubernetes fundamentals
- Terraform basics
- Go programming
- Docker containerization
- CI/CD concepts
- Basic AWS services

---

## Cost Estimation (Monthly - US East 1)

### Development Environment
| Service | Configuration | Cost |
|---------|---------------|------|
| EKS Cluster | Control plane | $73 |
| EC2 Nodes | 2x t3.medium | $60 |
| RDS PostgreSQL | db.t3.medium | $50 |
| NAT Gateway | 1x | $32 |
| API Gateway | 1M requests | $3.50 |
| ECR | 10GB storage | $1 |
| Secrets Manager | 5 secrets | $2 |
| CloudWatch Logs | 10GB | $5 |
| **Total** | | **~$227/month** |

### Production Environment
| Service | Configuration | Cost |
|---------|---------------|------|
| EKS Cluster | Control plane | $73 |
| EC2 Nodes | 4x t3.large | $240 |
| RDS PostgreSQL | db.r6g.large Multi-AZ | $350 |
| NAT Gateway | 3x (HA) | $96 |
| API Gateway | 10M requests | $35 |
| CloudFront | 100GB transfer | $8.50 |
| ALB | 1x | $22 |
| ECR | 50GB storage | $5 |
| Secrets Manager | 5 secrets | $2 |
| CloudWatch Logs | 50GB | $25 |
| **Total** | | **~$857/month** |

**Cost Optimization Tips**:
- Use Spot instances for non-critical workloads
- Right-size instances based on actual usage
- Use S3 lifecycle policies
- Enable EKS Fargate for sporadic workloads
- Use AWS Savings Plans for 40% discount

---

## Key Decisions & Trade-offs

### 1. Single Helm Chart vs. Multiple Charts
**Decision**: Single umbrella chart  
**Rationale**: Simpler management, atomic deployments, easier versioning  
**Trade-off**: Less flexibility for independent service releases

### 2. RDS vs. Self-Managed PostgreSQL
**Decision**: AWS RDS  
**Rationale**: Managed backups, HA, maintenance, security  
**Trade-off**: Higher cost, less control

### 3. API Gateway vs. Direct Ingress
**Decision**: API Gateway  
**Rationale**: Rate limiting, Lambda authorizer, AWS-native integration  
**Trade-off**: Additional cost, slight latency

### 4. Istio vs. AWS App Mesh
**Decision**: Istio  
**Rationale**: Industry standard, portability, rich features  
**Trade-off**: More complex, higher resource usage

### 5. ArgoCD vs. Flux
**Decision**: ArgoCD  
**Rationale**: Better UI, easier troubleshooting, wider adoption  
**Trade-off**: Slightly more resource intensive

### 6. Polling (CronJob) vs. Webhooks
**Decision**: CronJob every 5 minutes  
**Rationale**: Oura API doesn't support webhooks, simple implementation  
**Trade-off**: Not real-time, API quota usage

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Oura API rate limits | High | Implement exponential backoff, caching, alert on quota |
| Database connection exhaustion | High | Connection pooling, HPA, monitoring |
| EKS node failure | Medium | Multi-AZ, auto-scaling, PodDisruptionBudgets |
| Secret leakage | Critical | Secrets Manager, IRSA, no secrets in Git |
| Cost overrun | Medium | Budget alerts, resource quotas, auto-scaling limits |
| Service mesh complexity | Medium | Gradual rollout, training, comprehensive docs |

---

## Next Steps

1. **Review this plan** and confirm alignment with requirements
2. **Set up AWS account** and configure billing alerts
3. **Create GitHub repository** structure
4. **Obtain Oura API credentials** (when ready)
5. **Begin Phase 1** infrastructure setup

---

## Confirmed Project Details

Based on your clarifications, here's what we're building:

- **AWS Region**: us-east-1
- **Domain**: api.myhealth.eric-n.com
- **Budget**: Test/Dev phase (no constraints)
- **Frontend**: Grafana-only (simplified UI)
- **Users**: Single user (you)
- **Existing Infrastructure**: Refactor and extend `eks-three-tier-app`
- **Environments**: Dev initially (can add staging/prod later)

---

## Additional Steps & Tooling Changes

Based on your answers, here are the adjustments to the original plan:

### ✅ No Additional Tools Needed
Your existing setup already covers all required components. However, there are some **refinements and additions**:

### 1. **Simplifications (Faster Time-to-Value)**

| Aspect | Original | Simplified |
|--------|----------|-----------|
| Environments | Dev + Staging + Prod | Dev only (easily extensible) |
| User Management | Multi-tenant system | Single-user (no RBAC complexity) |
| Frontend | Custom web UI + Grafana | Grafana-only (Istio Ingress only) |
| Database | Multi-AZ RDS | Single-AZ db.t3.micro (cost-optimized) |
| Node Groups | Multiple specialized groups | Single general-purpose group (t3.medium Spot) |
| Monitoring | Full stack | Core stack (Prometheus + Grafana) |

**Impact**: ~40% faster implementation, ~60% cost reduction

---

### 2. **Refactoring Your Existing Terraform**

Your `eks-three-tier-app/terraform` structure is good but needs:

**Current State**:
- ✅ VPC with private/public subnets
- ✅ EKS cluster configuration (commented out)
- ✅ Managed node groups setup
- ✅ Security groups defined
- ⚠️ Missing: API Gateway, RDS, ECR, Secrets Manager, Istio modules

**Additions Needed**:
```
terraform/modules/
├── eks/                  # (exists as scattered code)
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── addons.tf         # (NEW)
│   └── irsa.tf           # (NEW)
├── rds/                  # (NEW)
│   ├── main.tf
│   ├── variables.tf
│   └── outputs.tf
├── ecr/                  # (NEW)
│   ├── main.tf
│   └── outputs.tf
├── api-gateway/          # (NEW)
│   ├── main.tf
│   └── outputs.tf
├── secrets-manager/      # (NEW)
│   └── main.tf
└── istio/                # (NEW)
    └── main.tf           # (Helm-based installation)

terraform/
├── dev/
│   ├── main.tf           # (NEW - environment-specific)
│   ├── terraform.tfvars  # (NEW - dev values)
│   └── backend.tf        # (NEW - S3 backend)
└── outputs.tf            # (refactored)
```

---

### 3. **Terraform Changes - Region Update**

Your current default is `us-west-2`. We need to:

**File: `variables.tf`**
```hcl
# Change default region
variable "region" {
  default = "us-east-1"  # ← Changed from us-west-2
}

# Add domain variable
variable "domain_name" {
  description = "Root domain name"
  default     = "eric-n.com"
}

# Add environment variable
variable "environment" {
  description = "Environment name"
  default     = "dev"
}
```

---

### 4. **Helm Chart Simplifications**

**Removed Components**:
- Ingress controller (using Istio only)
- Multiple dashboards (keep: Oura metrics + service health)
- ArgoCD separate installation (can add later if needed)

**Simplified Chart Structure**:
```
helm/myhealth/
├── Chart.yaml
├── values.yaml               # (single values file - dev)
├── templates/
│   ├── _helpers.tpl
│   ├── oura-collector/
│   │   ├── cronjob.yaml
│   │   └── servicemonitor.yaml
│   ├── data-processor/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── hpa.yaml
│   │   └── servicemonitor.yaml
│   ├── api-service/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── hpa.yaml
│   │   ├── virtualservice.yaml
│   │   └── servicemonitor.yaml
│   ├── prometheus/
│   │   ├── deployment.yaml
│   │   └── configmap-scrapeconfig.yaml
│   ├── grafana/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── configmap-datasources.yaml
│   │   └── configmap-dashboards.yaml
│   └── istio/
│       ├── gateway.yaml
│       └── virtualservice-istio-gateway.yaml
└── dashboards/               # (JSON files)
    ├── oura-metrics.json
    └── service-health.json
```

---

### 5. **CI/CD Pipeline - Simplified**

**GitHub Actions Workflows Kept**:
- `terraform-plan.yml` (PR validation)
- `terraform-apply.yml` (infrastructure deployment)
- `build-*.yml` (service builds)
- `helm-lint.yml` (Helm validation)

**GitHub Actions Workflows Removed** (add later):
- Multiple environment workflows (just use dev)
- Staging/prod approval workflows
- Complex promotion pipelines

---

### 6. **Microservices - No Changes**

Your 3 Go services remain the same:
- `oura-collector` (CronJob every 5 min)
- `data-processor` (HTTP service)
- `api-service` (REST API)

**Single-User Implication**:
- No multi-tenant database schema needed
- Simpler authentication (just JWT with fixed user_id=1)
- No user management API endpoints

---

### 7. **Database Schema - Simplified**

```sql
-- Single-user focused schema
CREATE TABLE sleep_metrics (
  id BIGSERIAL PRIMARY KEY,
  date DATE NOT NULL UNIQUE,
  score INT,
  duration_seconds INT,
  efficiency DECIMAL,
  latency_minutes INT,
  
  deep_sleep_seconds INT,
  light_sleep_seconds INT,
  rem_sleep_seconds INT,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE activity_metrics (
  id BIGSERIAL PRIMARY KEY,
  date DATE NOT NULL UNIQUE,
  score INT,
  steps INT,
  calories INT,
  active_minutes INT,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE readiness_metrics (
  id BIGSERIAL PRIMARY KEY,
  date DATE NOT NULL UNIQUE,
  score INT,
  hrv DECIMAL,
  resting_hr INT,
  temperature_deviation DECIMAL,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Remove users table entirely (single-user app)
```

---

### 8. **API Simplifications**

**Removed Endpoints**:
- `/auth/register` (no user management)
- `/auth/refresh` (simpler token handling)
- `/users/*` (single user)
- `/admin/*` (no admin panel)

**Kept Endpoints**:
```
POST   /auth/login              # Single-user login (API key based)
GET    /api/v1/dashboard        # Dashboard summary
GET    /api/v1/sleep            # Sleep metrics
GET    /api/v1/activity         # Activity metrics
GET    /api/v1/readiness        # Readiness metrics
GET    /api/v1/trends           # Trends query
GET    /health                  # Health check
GET    /metrics                 # Prometheus metrics
```

---

### 9. **Security - Simplifications**

| Component | Original | Simplified |
|-----------|----------|-----------|
| Auth | Multi-tenant JWT | Single API key in Secrets Manager |
| RBAC | Role-based (admin/user) | None (single user) |
| Encryption | Full envelope encryption | At-rest only (RDS KMS) |
| CORS | Complex origin handling | Allow eric-n.com only |
| WAF | AWS WAF enabled | API Gateway rate limiting only |

---

### 10. **Monitoring Simplifications**

**Kept**:
- Prometheus (metrics collection)
- Grafana (visualization)
- ServiceMonitors (metrics discovery)
- 2 pre-built dashboards (Oura + Service Health)

**Removed**:
- Alertmanager + complex alerting rules
- Jaeger distributed tracing
- CloudWatch custom metrics
- Advanced logging pipeline (use CloudWatch Logs only)

---

## Updated Implementation Timeline

With simplifications, new timeline:

| Phase | Original | Simplified | Notes |
|-------|----------|-----------|-------|
| 1: Infrastructure | 2 weeks | 1 week | Single env, fewer variables |
| 2: Microservices | 3 weeks | 2 weeks | No multi-tenant logic |
| 3: Helm Chart | 1 week | 3 days | Fewer templates |
| 4: CI/CD | 1 week | 3 days | Single deployment pipeline |
| 5: Monitoring | 1 week | 3 days | 2 dashboards only |
| 6: API Gateway | 2 weeks | 1 week | Simpler auth (API key) |
| 7: Security | 1 week | 2 days | Reduced RBAC |
| 8: Testing | 1 week | 3 days | Single environment |
| **Total** | **12 weeks** | **~4-5 weeks** | **~60% reduction** |

---

## Cost Impact

**Revised Cost Estimate (Dev - us-east-1)**:

| Service | Configuration | Cost |
|---------|---------------|------|
| EKS Cluster | Control plane | $73 |
| EC2 Nodes | 2x t3.medium Spot | $20 |
| RDS PostgreSQL | db.t3.micro Single-AZ | $15 |
| NAT Gateway | 1x | $32 |
| API Gateway | 1M requests/month | $3.50 |
| ECR | 10GB storage | $1 |
| Secrets Manager | 5 secrets | $2 |
| CloudWatch Logs | 5GB/month | $2.50 |
| **Total** | | **~$149/month** |

**Optimization**: Use Reserved Instances or Savings Plans for 30% additional discount (~$100/month)

---

## What Can Be Reused From Current Setup

✅ **VPC & Networking** - Already configured in your `networking.tf`  
✅ **EKS Cluster Skeleton** - Code exists but commented out; just uncomment and refactor  
✅ **Node Groups Config** - Already in `variables.tf`  
✅ **Security Groups** - Already defined  
✅ **Provider Config** - Already setup in `providers.tf`  

⚠️ **What Needs Adding**:
- EKS add-ons configuration (VPC CNI, CoreDNS, kube-proxy)
- IRSA setup
- RDS PostgreSQL module
- ECR module
- API Gateway module
- Secrets Manager module
- Istio Helm deployment

---

**Ready to Proceed?** Let me know if you'd like me to:
1. Start creating the refactored Terraform structure
2. Build the three microservices in Go
3. Create the simplified Helm chart
4. Set up GitHub Actions workflows
5. All of the above in sequence

