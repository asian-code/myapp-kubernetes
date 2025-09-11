# Three-Tier Application Portfolio - AKS Implementation

![Project Status](https://img.shields.io/badge/status-production--ready-green)
![Kubernetes](https://img.shields.io/badge/kubernetes-1.28+-blue)
![Azure AKS](https://img.shields.io/badge/Azure-AKS-blue)
![Terraform](https://img.shields.io/badge/terraform-1.5+-purple)

A comprehensive, production-ready three-tier application deployed on Azure AKS with full observability, security, and automation.

## 🏗️ Architecture Overview

```text
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│     Frontend        │    │      Backend        │    │     Database        │
│   (React/Nginx)     │    │   (Node.js API)     │    │   (PostgreSQL)      │
│                     │    │                     │    │                     │
│ • Nginx 1.25        │◄───┤ • Express.js        │◄───┤ • PostgreSQL 15     │
│ • React SPA         │    │ • Health Checks     │    │ • Persistent Vol    │
│ • Static Assets     │    │ • Metrics Export    │    │ • Backup Strategy   │
│ • SSL/TLS          │    │ • Auto-scaling      │    │ • HA Configuration  │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
         │                           │                           │
         └───────────────────────────┼───────────────────────────┘
                                     │
                    ┌─────────────────────┐
                    │   Observability     │
                    │                     │
                    │ • Prometheus        │
                    │ • Grafana          │
                    │ • Azure Monitor     │
                    │ • Application Insights │
                    └─────────────────────┘
```

## 🚀 Project Portfolio Features

This Azure implementation demonstrates enterprise-grade Kubernetes deployment capabilities across **7 comprehensive phases**:

### Phase 1: ✅ Project Setup & Repository Structure

- Multi-cloud architecture (EKS & AKS implementations)
- Modular Terraform infrastructure
- Comprehensive documentation and README files
- Git repository organization following best practices

### Phase 2: ✅ Cluster Provisioning

- **AKS Implementation**: Production-ready cluster with Virtual Network, NSGs
- **Multi-Zone Deployment**: High availability across availability zones
- **Auto Scaling**: Both cluster and application-level scaling
- **Cost Optimization**: Spot node pools and Azure Reserved Instances

### Phase 3: ✅ 3-Tier Application Deployment

- **Frontend**: React.js with Nginx, optimized Docker containers
- **Backend**: Node.js Express API with health checks and metrics
- **Database**: PostgreSQL with persistent storage and backup strategies
- **Service Mesh Ready**: Prepared for Istio or Linkerd integration

### Phase 4: ✅ Observability Stack

- **Prometheus**: Comprehensive metrics collection with custom rules
- **Grafana**: Rich dashboards for application and infrastructure monitoring
- **Azure Monitor**: Native Azure monitoring integration
- **Application Insights**: APM with distributed tracing
- **Log Analytics**: Centralized logging with KQL queries

### Phase 5: ✅ Security & Secrets Management

- **RBAC**: Fine-grained role-based access control
- **Network Policies**: Calico-based micro-segmentation
- **Pod Security Standards**: Restricted security contexts
- **Azure Key Vault**: Secrets management with CSI driver
- **Workload Identity**: AAD pod identity integration
- **Azure Defender**: Container security scanning
- **Compliance**: Azure Policy and CIS benchmarks

### Phase 6: ✅ CI/CD Automation

- **GitHub Actions**: Azure-specific deployment pipeline
- **Multi-Environment**: Dev, staging, production deployments
- **Security Testing**: Automated scanning with Azure Security Center
- **Load Testing**: Azure Load Testing integration
- **Blue-Green Deployments**: Zero-downtime deployment strategies

### Phase 7: ✅ Bonus Enhancements

- **Azure Monitor Integration**: Native cloud monitoring
- **Key Vault CSI Driver**: Seamless secret injection
- **Azure Policy**: Governance and compliance automation
- **Multi-Region**: Disaster recovery across regions
- **Cost Management**: Azure Cost Analysis integration

## 📋 Prerequisites

### Required Tools

```bash
# Core tools
azure-cli >= 2.50.0
kubectl >= 1.28.0
terraform >= 1.5.0
docker >= 24.0.0
helm >= 3.12.0

# Optional but recommended
k9s >= 0.27.0
kubectx >= 0.9.4
kubelogin >= 0.0.29
```

### Azure Permissions

- AKS cluster management
- Virtual Network configuration
- Azure Active Directory integration
- Azure Container Registry access
- Key Vault management
- Monitor and Log Analytics workspace

## 🛠️ Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd aks-three-tier-app

# Configure Azure CLI
az login
az account set --subscription "your-subscription-id"

# Set environment variables
export AZURE_REGION=eastus
export AKS_CLUSTER_NAME=three-tier-aks-cluster
export AKS_RESOURCE_GROUP=three-tier-aks-rg
```

### 2. Deploy Infrastructure (Phase 2)

```bash
# Option 1: Use helper script (recommended)
chmod +x scripts/deploy.sh
./scripts/deploy.sh deploy

# Option 2: Manual deployment
cd terraform
terraform init
terraform plan -var-file="terraform.tfvars"
terraform apply -var-file="terraform.tfvars"
```

### 3. Deploy Applications (Phase 3)

```bash
# Get AKS credentials
az aks get-credentials --resource-group $AKS_RESOURCE_GROUP --name $AKS_CLUSTER_NAME

# Deploy all tiers
kubectl apply -f k8s-manifests/database/
kubectl apply -f k8s-manifests/backend/
kubectl apply -f k8s-manifests/frontend/

# Check deployment status
kubectl get pods -A
```

### 4. Setup Observability (Phase 4)

```bash
# Deploy monitoring stack
kubectl apply -f k8s-manifests/monitoring/

# Enable Container Insights
az aks enable-addons --resource-group $AKS_RESOURCE_GROUP --name $AKS_CLUSTER_NAME --addons monitoring

# Access Grafana
kubectl port-forward svc/grafana -n monitoring 3000:3000
```

### 5. Configure Security (Phase 5)

```bash
# Deploy security policies
kubectl apply -f k8s-manifests/security/

# Enable Azure Key Vault CSI Driver
./scripts/deploy.sh keyvault

# Verify Workload Identity
kubectl get serviceaccounts -A
```

## 📁 Project Structure

```text
aks-three-tier-app/
├── terraform/                     # Phase 2: Infrastructure as Code
│   ├── main.tf                   # Root Terraform configuration
│   ├── variables.tf              # Variable definitions
│   ├── outputs.tf                # Output values
│   ├── terraform.tfvars          # Variable values
│   └── modules/                  # Reusable Terraform modules
│       ├── aks/                  # AKS cluster module
│       ├── networking/           # Virtual Network module
│       └── security/             # NSGs & Azure Policy
├── k8s-manifests/                # Phase 3: Application Deployment
│   ├── database/                 # PostgreSQL configuration
│   │   ├── postgresql.yaml       # StatefulSet with Azure Disk
│   │   └── configmap.yaml        # Database configuration
│   ├── backend/                  # Node.js API configuration
│   │   ├── backend-api.yaml      # Deployment with HPA
│   │   ├── service.yaml          # ClusterIP service
│   │   └── configmap.yaml        # Application configuration
│   ├── frontend/                 # React app configuration
│   │   ├── frontend-web.yaml     # Nginx deployment
│   │   ├── service.yaml          # LoadBalancer service
│   │   └── ingress.yaml          # NGINX ingress configuration
│   ├── monitoring/               # Phase 4: Observability Stack
│   │   ├── prometheus.yaml       # Prometheus with Azure integration
│   │   └── grafana.yaml          # Grafana with Azure AD auth
│   └── security/                 # Phase 5: Security & Policies
│       ├── rbac.yaml             # Role-based access control
│       ├── network-policies.yaml # Calico network segmentation
│       ├── external-secrets.yaml # Key Vault integration
│       └── pod-security-standards.yaml # Security contexts
├── app-src/                      # Application source code
│   ├── frontend/                 # React application
│   │   ├── Dockerfile           # Multi-stage build
│   │   ├── package.json         # Dependencies
│   │   └── src/                 # React components
│   ├── backend/                  # Node.js API
│   │   ├── Dockerfile           # Optimized Node.js image
│   │   ├── package.json         # API dependencies
│   │   └── src/                 # Express.js application
│   └── database/                 # Database schemas
│       └── init.sql             # Database initialization
├── ci-cd/                        # Phase 6: CI/CD Automation
│   └── github-actions/           # GitHub Actions workflows
│       ├── aks-deploy.yml        # AKS deployment pipeline
│       └── security-scan.yml     # Security scanning workflow
├── scripts/                      # Phase 7: Helper utilities
│   ├── deploy.sh                 # Deployment automation
│   ├── monitoring.sh             # Azure Monitor setup
│   └── troubleshoot.sh           # Diagnostic tools
├── tests/                        # Phase 7: Testing
│   ├── load-test.js             # K6 performance testing
│   ├── security-test.sh         # Security validation
│   └── integration-test.js      # API integration tests
└── docs/                         # Comprehensive documentation
    ├── ARCHITECTURE.md           # System architecture
    ├── DEPLOYMENT.md             # Deployment guide
    ├── MONITORING.md             # Azure Monitor setup
    ├── SECURITY.md               # Security implementation
    └── TROUBLESHOOTING.md        # Common issues & solutions
```

## 🔧 Configuration Examples

### Terraform Configuration

```hcl
# terraform/terraform.tfvars
resource_group_name = "three-tier-aks-rg"
location = "East US"
cluster_name = "three-tier-aks-cluster"
kubernetes_version = "1.28.3"

# Virtual Network configuration
vnet_address_space = ["10.0.0.0/16"]
subnet_address_prefixes = {
  aks_subnet     = ["10.0.1.0/24"]
  gateway_subnet = ["10.0.2.0/24"]
  db_subnet      = ["10.0.3.0/24"]
}

# Node pool configuration
default_node_pool = {
  name       = "default"
  node_count = 3
  min_count  = 1
  max_count  = 10
  vm_size    = "Standard_D2s_v3"
  disk_size  = 50
}

additional_node_pools = {
  spot = {
    name               = "spot"
    vm_size           = "Standard_D2s_v3"
    node_count        = 2
    min_count         = 0
    max_count         = 20
    priority          = "Spot"
    eviction_policy   = "Delete"
    spot_max_price    = 0.5
  }
}

# Azure Monitor integration
enable_log_analytics = true
enable_container_insights = true
enable_azure_defender = true

# Azure Key Vault
key_vault_name = "three-tier-app-kv"
enable_key_vault_csi = true

# Azure Container Registry
acr_name = "threetierappacr"
acr_sku = "Premium"

tags = {
  Environment = "production"
  Project     = "three-tier-portfolio"
  Owner       = "platform-engineering"
  CostCenter  = "engineering"
}
```

## 📊 Azure Monitoring Integration (Phase 4)

### Azure Monitor Features

**Container Insights:**

- **Cluster Performance**: Node and pod resource utilization
- **Live Metrics**: Real-time performance data
- **Log Analytics**: KQL queries for custom insights
- **Workbooks**: Interactive dashboards and reports

**Application Insights Integration:**

```javascript
// Application Insights configuration
const appInsights = require('applicationinsights');
appInsights.setup(process.env.APPINSIGHTS_CONNECTION_STRING)
    .setAutoDependencyCorrelation(true)
    .setAutoCollectRequests(true)
    .setAutoCollectPerformance(true)
    .setAutoCollectExceptions(true)
    .setAutoCollectDependencies(true)
    .setAutoCollectConsole(true)
    .setUseDiskRetryCaching(true)
    .setSendLiveMetrics(true);

appInsights.start();
```

**Custom KQL Queries:**

```kusto
// High CPU usage pods
let threshold = 80;
Perf
| where ObjectName == "K8SContainer"
| where CounterName == "cpuUsageNanoCores"
| extend CPUPercent = CounterValue / 10000000
| where CPUPercent > threshold
| summarize avg(CPUPercent) by Computer, InstanceName
| order by avg_CPUPercent desc
```

## 🔐 Azure Security Implementation (Phase 5)

### Azure Active Directory Integration

**Workload Identity Configuration:**

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: backend-workload-identity
  namespace: backend
  annotations:
    azure.workload.identity/client-id: "12345678-1234-1234-1234-123456789012"
    azure.workload.identity/tenant-id: "87654321-4321-4321-4321-210987654321"
  labels:
    azure.workload.identity/use: "true"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-api
spec:
  template:
    metadata:
      labels:
        azure.workload.identity/use: "true"
    spec:
      serviceAccountName: backend-workload-identity
      containers:
      - name: backend
        image: threetierappacr.azurecr.io/backend:latest
        env:
        - name: AZURE_CLIENT_ID
          value: "12345678-1234-1234-1234-123456789012"
```

### Azure Key Vault CSI Driver

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: azure-keyvault-csi
  namespace: backend
spec:
  provider: azure
  parameters:
    usePodIdentity: "false"
    useVMManagedIdentity: "false"
    userAssignedIdentityID: "12345678-1234-1234-1234-123456789012"
    keyvaultName: "three-tier-app-kv"
    objects: |
      array:
        - |
          objectName: database-password
          objectType: secret
          objectVersion: ""
        - |
          objectName: jwt-secret
          objectType: secret
          objectVersion: ""
  secretObjects:
  - secretName: app-secrets
    type: Opaque
    data:
    - objectName: database-password
      key: POSTGRES_PASSWORD
    - objectName: jwt-secret
      key: JWT_SECRET
```

### Azure Policy Integration

```json
{
  "if": {
    "allOf": [
      {
        "field": "type",
        "equals": "Microsoft.ContainerService/managedClusters/pods"
      },
      {
        "field": "Microsoft.ContainerService/managedClusters/pods/containers[*].securityContext.runAsRoot",
        "equals": true
      }
    ]
  },
  "then": {
    "effect": "deny"
  }
}
```

## 🔄 Azure CI/CD Pipeline (Phase 6)

### GitHub Actions with Azure Integration

**Azure Login and Deployment:**

```yaml
- name: Azure CLI login
  uses: azure/login@v1
  with:
    creds: ${{ secrets.AZURE_CREDENTIALS }}

- name: Login to Azure Container Registry
  run: |
    az acr login --name ${{ env.ACR_NAME }}

- name: Build and push Docker image
  run: |
    docker build -t ${{ env.ACR_NAME }}.azurecr.io/backend:${{ github.sha }} ./app-src/backend/
    docker push ${{ env.ACR_NAME }}.azurecr.io/backend:${{ github.sha }}

- name: Deploy to AKS
  uses: azure/k8s-deploy@v1
  with:
    manifests: |
      k8s-manifests/backend/
    images: |
      ${{ env.ACR_NAME }}.azurecr.io/backend:${{ github.sha }}
```

**Azure Load Testing Integration:**

```yaml
- name: Run Azure Load Test
  uses: azure/load-testing@v1
  with:
    loadTestConfigFile: 'tests/azure-load-test.yaml'
    loadTestResource: 'three-tier-load-test'
    resourceGroup: ${{ env.AKS_RESOURCE_GROUP }}
```

## 🧪 Azure-Specific Testing (Phase 7)

### Azure Load Testing Configuration

```yaml
# azure-load-test.yaml
version: v0.1
testName: three-tier-app-load-test
testPlan: tests/load-test.jmx
description: Load test for three-tier application
engineInstances: 5
configurationFiles:
- name: config.csv
  value: tests/test-data.csv
secrets:
- name: app_url
  value: https://your-app-domain.com
failureCriteria:
- avg(response_time_ms) > 500
- percentage(error) > 5
- avg(latency) > 1000
```

## 🚨 Azure-Specific Troubleshooting

### Common AKS Issues

**1. Node Pool Scaling Issues:**

```bash
# Check node pool status
az aks nodepool show --resource-group $AKS_RESOURCE_GROUP --cluster-name $AKS_CLUSTER_NAME --name default

# Check cluster autoscaler logs
kubectl logs -n kube-system -l app=cluster-autoscaler

# Check node conditions
kubectl describe nodes
```

**2. Azure CNI Networking:**

```bash
# Check Azure CNI configuration
kubectl get pods -n kube-system | grep azure-cni

# Check IP allocation
az network vnet subnet show --resource-group MC_${AKS_RESOURCE_GROUP}_${AKS_CLUSTER_NAME}_${LOCATION} --vnet-name aks-vnet --name aks-subnet

# Check network policies
kubectl get netpol -A
```

**3. Azure Key Vault Integration:**

```bash
# Check CSI driver status
kubectl get pods -n kube-system | grep secrets-store

# Check workload identity
kubectl get azureidentity
kubectl get azureidentitybinding

# Check secret provider class
kubectl describe secretproviderclass azure-keyvault-csi -n backend
```

## 📈 Azure Cost Optimization

### Cost Management Features

**Azure Cost Analysis Integration:**

```bash
# Check cluster costs
az consumption usage list --start-date 2024-01-01 --end-date 2024-01-31 --resource-group $AKS_RESOURCE_GROUP

# Set up cost alerts
az consumption budget create --resource-group $AKS_RESOURCE_GROUP --budget-name aks-monthly-budget --amount 1000 --time-grain Monthly
```

**Spot Node Pools:**

```hcl
additional_node_pools = {
  spot = {
    name               = "spot"
    vm_size           = "Standard_D2s_v3"
    priority          = "Spot"
    eviction_policy   = "Delete"
    spot_max_price    = 0.5  # 50% of on-demand price
    node_taints       = ["kubernetes.azure.com/scalesetpriority=spot:NoSchedule"]
  }
}
```

## 🔄 Azure Backup & Disaster Recovery

### Azure-Native Backup Solutions

**AKS Backup Configuration:**

```yaml
apiVersion: dataprotection.azure.com/v1beta1
kind: BackupPolicy
metadata:
  name: aks-backup-policy
spec:
  datasourceType: "Microsoft.ContainerService/managedClusters"
  policyRules:
  - backupParameters:
      backupType: "Incremental"
    trigger:
      schedule:
        repeatingTimeIntervals:
        - "R/2024-01-01T02:00:00+00:00/P1D"  # Daily at 2 AM UTC
  retentionPolicy:
    defaultPolicy:
      lifecycles:
      - deleteAfter:
          duration: "P30D"
        sourceDataStore:
          dataStoreType: "VaultStore"
```

**Cross-Region Replication:**

```bash
# Enable geo-redundant backup for ACR
az acr replication create --resource-group $AKS_RESOURCE_GROUP --registry $ACR_NAME --location westus2

# Configure cross-region cluster for DR
az aks create --resource-group three-tier-aks-dr-rg --name three-tier-aks-dr-cluster --location westus2
```

## 🌟 Azure-Specific Learning Outcomes

This Azure implementation demonstrates expertise in:

### Azure Platform Engineering

- **Azure Kubernetes Service (AKS)** cluster management
- **Azure Container Registry (ACR)** integration
- **Azure Monitor** and Application Insights
- **Azure Key Vault** secrets management

### Azure Security & Compliance

- **Azure Active Directory** workload identity
- **Azure Policy** governance automation
- **Azure Defender** container security
- **Network Security Groups** and Azure Firewall

### Azure DevOps Integration

- **Azure Pipelines** CI/CD automation
- **Azure Load Testing** performance validation
- **Azure Cost Management** optimization
- **Azure Resource Manager** templates

## 📚 Azure Resources & Documentation

**Official Azure Documentation:**

- [Azure Kubernetes Service (AKS)](https://docs.microsoft.com/en-us/azure/aks/)
- [Azure Container Registry](https://docs.microsoft.com/en-us/azure/container-registry/)
- [Azure Key Vault](https://docs.microsoft.com/en-us/azure/key-vault/)
- [Azure Monitor](https://docs.microsoft.com/en-us/azure/azure-monitor/)

**Best Practices:**

- [AKS Best Practices](https://docs.microsoft.com/en-us/azure/aks/best-practices)
- [Azure Security Baseline](https://docs.microsoft.com/en-us/security/benchmark/azure/)
- [Well-Architected Framework](https://docs.microsoft.com/en-us/azure/architecture/framework/)

## 🎯 Portfolio Highlights

**Azure-Specific Achievements:**

✅ **Native Integration**: Deep Azure services integration
✅ **Enterprise Security**: AAD, Key Vault, and Policy integration
✅ **Cost Optimized**: Spot instances and reserved capacity
✅ **Highly Available**: Multi-zone deployment with auto-scaling
✅ **Monitoring Native**: Azure Monitor and Application Insights
✅ **Compliance Ready**: Azure Policy and security baselines
✅ **DevOps Integrated**: GitHub Actions with Azure services

## 🏆 Conclusion

This AKS Three-Tier Application showcases enterprise-grade Azure Kubernetes deployments with native cloud service integrations. It demonstrates mastery of Azure platform services, security best practices, and production-ready infrastructure automation.

**Key Differentiators:**

- 🔵 **Azure-Native**: Leverages native Azure services extensively
- 🔐 **Enterprise Security**: AAD integration and advanced security
- 📊 **Rich Monitoring**: Application Insights and Azure Monitor
- 💰 **Cost Optimized**: Spot instances and cost management
- 🚀 **Production Ready**: High availability and disaster recovery
- 📈 **Scalable**: Auto-scaling at multiple layers

This implementation serves as a comprehensive demonstration of Azure Kubernetes expertise suitable for enterprise environments and technical portfolio showcases.

---

**🔵 Azure Native | 🔒 Enterprise Secure | 📊 Fully Observable | 🤖 CI/CD Automated**
