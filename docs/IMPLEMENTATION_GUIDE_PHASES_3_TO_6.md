# myHealth - Implementation Guide (Continued)
## Phases 3-6: Helm Charts, CI/CD, Monitoring, API Gateway

---

# PHASE 3: HELM CHART CREATION
**Duration**: 3 days  
**Objective**: Create single Helm chart to deploy entire stack

## Sprint 3.1: Helm Chart Foundation

### Task 3.1.1: Create Helm Chart Structure
**What**: Setup chart directory structure  
**Where**: `helm/myhealth/`

```bash
mkdir -p helm/myhealth/{templates,dashboards}

# Create template subdirectories
mkdir -p helm/myhealth/templates/{oura-collector,data-processor,api-service,prometheus,grafana,istio}
```

---

### Task 3.1.2: Create Chart.yaml
**File**: `helm/myhealth/Chart.yaml`

```yaml
apiVersion: v2
name: myhealth
description: Oura Ring metrics dashboard on Kubernetes
type: application
version: 0.1.0
appVersion: "1.0"

keywords:
  - oura-ring
  - metrics
  - kubernetes
  - prometheus
  - grafana

home: https://eric-n.com
sources:
  - https://github.com/asian-code/myapp-kubernetes

dependencies:
  - name: prometheus
    version: "25.8.0"
    repository: "https://prometheus-community.github.io/helm-charts"
    condition: prometheus.enabled
    alias: prometheus
  
  - name: grafana
    version: "7.0.0"
    repository: "https://grafana.github.io/helm-charts"
    condition: grafana.enabled
    alias: grafana
```

**Instructions**:
1. Create file with above content
2. Note: This references external Helm charts as dependencies

---

### Task 3.1.3: Create values.yaml
**File**: `helm/myhealth/values.yaml`

```yaml
# Global settings
global:
  environment: dev
  region: us-east-1
  domain: eric-n.com
  namespace: myhealth

# Image registry (ECR)
imageRegistry: "YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com"
imagePullPolicy: IfNotPresent

# Oura Collector (CronJob)
ouraCollector:
  enabled: true
  image:
    repository: "myhealth/oura-collector"
    tag: "latest"
  
  schedule: "*/5 * * * *"  # Every 5 minutes
  
  env:
    processorUrl: "http://data-processor:8080"
  
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi
  
  # Secret reference
  secret:
    name: myhealth-secrets
    ouraApiKeyField: oura_api_key

# Data Processor (Deployment)
dataProcessor:
  enabled: true
  image:
    repository: "myhealth/data-processor"
    tag: "latest"
  
  replicaCount: 2
  
  service:
    type: ClusterIP
    port: 8080
    targetPort: 8080
  
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
  
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
    targetCPUUtilizationPercentage: 70
  
  env:
    logLevel: "info"
  
  secret:
    name: myhealth-secrets
    dbHostField: db_host
    dbUserField: db_user
    dbPassField: db_password
    dbNameField: db_name

# API Service (Deployment)
apiService:
  enabled: true
  image:
    repository: "myhealth/api-service"
    tag: "latest"
  
  replicaCount: 3
  
  service:
    type: ClusterIP
    port: 8080
    targetPort: 8080
  
  resources:
    requests:
      cpu: 300m
      memory: 384Mi
    limits:
      cpu: 1000m
      memory: 768Mi
  
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
  
  env:
    logLevel: "info"
    processorUrl: "http://data-processor:8080"
  
  secret:
    name: myhealth-secrets
    jwtSecretField: jwt_secret
    dbHostField: db_host
    dbUserField: db_user
    dbPassField: db_password

# Istio Configuration
istio:
  enabled: true
  gateway:
    name: myhealth-gateway
    namespace: myhealth
    hosts:
      - api.eric-n.com
  
  virtualServices:
    apiService:
      name: api-service-vs
      hosts:
        - api.eric-n.com
      gateways:
        - myhealth-gateway
      destinations:
        - host: api-service
          port: 8080

# Prometheus
prometheus:
  enabled: true
  
  server:
    retention: "15d"
    storageSize: "10Gi"
    
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: 500m
        memory: 1Gi
    
    persistence:
      enabled: true
      storageClassName: "ebs-sc"
      size: 10Gi
  
  alertmanager:
    enabled: false
  
  pushgateway:
    enabled: false

# Grafana
grafana:
  enabled: true
  
  replicas: 1
  
  service:
    type: ClusterIP
    port: 3000
  
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  
  persistence:
    enabled: true
    storageClassName: "ebs-sc"
    size: 5Gi
  
  datasources:
    prometheus:
      url: "http://prometheus-server"
      isDefault: true
  
  adminUser: admin
  # Password from secret

# ConfigMaps for dashboards
dashboards:
  enabled: true
  
  # Dashboard files will be mounted as ConfigMaps
  ouraMetrics: true
  serviceHealth: true

# Service Account
serviceAccount:
  create: true
  annotations: {}
  name: myhealth
```

**Instructions**:
1. Create file with above content
2. Replace `YOUR_ACCOUNT_ID` with actual AWS account ID
3. This is the default values file

---

### Task 3.1.4: Create Template Helpers
**File**: `helm/myhealth/templates/_helpers.tpl`

```yaml
{{/*
Expand the name of the chart.
*/}}
{{- define "myhealth.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "myhealth.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version
*/}}
{{- define "myhealth.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "myhealth.labels" -}}
helm.sh/chart: {{ include "myhealth.chart" . }}
{{ include "myhealth.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "myhealth.selectorLabels" -}}
app.kubernetes.io/name: {{ include "myhealth.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
```

---

## Sprint 3.2: Service Templates

### Task 3.2.1: Create Namespace
**File**: `helm/myhealth/templates/namespace.yaml`

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
```

---

### Task 3.2.2: Create ServiceAccount
**File**: `helm/myhealth/templates/serviceaccount.yaml`

```yaml
{{- if .Values.serviceAccount.create -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "myhealth.fullname" . }}
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
  {{- with .Values.serviceAccount.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
```

---

### Task 3.2.3: Create Secret Template
**File**: `helm/myhealth/templates/secrets.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Values.dataProcessor.secret.name }}
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
type: Opaque
data:
  # Base64 encoded values (will be provided at deployment time)
  db_host: {{ .Values.database.host | b64enc }}
  db_user: {{ .Values.database.username | b64enc }}
  db_password: {{ .Values.database.password | b64enc }}
  db_name: myhealth
  oura_api_key: {{ .Values.ouraApiKey | b64enc }}
  jwt_secret: {{ .Values.jwtSecret | b64enc }}
```

---

### Task 3.2.4: Create oura-collector CronJob
**File**: `helm/myhealth/templates/oura-collector/cronjob.yaml`

```yaml
{{- if .Values.ouraCollector.enabled }}
apiVersion: batch/v1
kind: CronJob
metadata:
  name: oura-collector
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
    app.kubernetes.io/component: collector
spec:
  schedule: "{{ .Values.ouraCollector.schedule }}"
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            {{- include "myhealth.selectorLabels" . | nindent 12 }}
            app.kubernetes.io/component: collector
        spec:
          serviceAccountName: {{ include "myhealth.fullname" . }}
          containers:
          - name: oura-collector
            image: "{{ .Values.imageRegistry }}/{{ .Values.ouraCollector.image.repository }}:{{ .Values.ouraCollector.image.tag }}"
            imagePullPolicy: {{ .Values.imagePullPolicy }}
            
            env:
            - name: PROCESSOR_URL
              value: "{{ .Values.ouraCollector.env.processorUrl }}"
            - name: OURA_API_KEY
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.ouraCollector.secret.name }}
                  key: {{ .Values.ouraCollector.secret.ouraApiKeyField }}
            - name: LOG_LEVEL
              value: "{{ .Values.ouraCollector.env.logLevel | default "info" }}"
            
            resources:
              {{- toYaml .Values.ouraCollector.resources | nindent 14 }}
          
          restartPolicy: OnFailure
{{- end }}
```

---

### Task 3.2.5: Create ServiceMonitor for oura-collector
**File**: `helm/myhealth/templates/oura-collector/servicemonitor.yaml`

```yaml
{{- if .Values.ouraCollector.enabled }}
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: oura-collector
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "myhealth.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: collector
  endpoints:
  - port: metrics
    interval: 30s
{{- end }}
```

---

### Task 3.2.6: Create data-processor Deployment
**File**: `helm/myhealth/templates/data-processor/deployment.yaml`

```yaml
{{- if .Values.dataProcessor.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: data-processor
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
    app.kubernetes.io/component: processor
spec:
  replicas: {{ .Values.dataProcessor.replicaCount }}
  
  selector:
    matchLabels:
      {{- include "myhealth.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: processor
  
  template:
    metadata:
      labels:
        {{- include "myhealth.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: processor
    
    spec:
      serviceAccountName: {{ include "myhealth.fullname" . }}
      
      containers:
      - name: data-processor
        image: "{{ .Values.imageRegistry }}/{{ .Values.dataProcessor.image.repository }}:{{ .Values.dataProcessor.image.tag }}"
        imagePullPolicy: {{ .Values.imagePullPolicy }}
        
        ports:
        - containerPort: {{ .Values.dataProcessor.service.targetPort }}
          name: http
        - containerPort: 9090
          name: metrics
        
        env:
        - name: DATABASE_HOST
          valueFrom:
            secretKeyRef:
              name: {{ .Values.dataProcessor.secret.name }}
              key: {{ .Values.dataProcessor.secret.dbHostField }}
        - name: DATABASE_USER
          valueFrom:
            secretKeyRef:
              name: {{ .Values.dataProcessor.secret.name }}
              key: {{ .Values.dataProcessor.secret.dbUserField }}
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: {{ .Values.dataProcessor.secret.name }}
              key: {{ .Values.dataProcessor.secret.dbPassField }}
        - name: DATABASE_NAME
          value: "myhealth"
        - name: LOG_LEVEL
          value: "{{ .Values.dataProcessor.env.logLevel }}"
        
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 10
          periodSeconds: 10
        
        readinessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
        
        resources:
          {{- toYaml .Values.dataProcessor.resources | nindent 12 }}
{{- end }}
```

---

### Task 3.2.7: Create data-processor Service
**File**: `helm/myhealth/templates/data-processor/service.yaml`

```yaml
{{- if .Values.dataProcessor.enabled }}
apiVersion: v1
kind: Service
metadata:
  name: data-processor
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
    app.kubernetes.io/component: processor
spec:
  type: {{ .Values.dataProcessor.service.type }}
  ports:
  - port: {{ .Values.dataProcessor.service.port }}
    targetPort: {{ .Values.dataProcessor.service.targetPort }}
    protocol: TCP
    name: http
  selector:
    {{- include "myhealth.selectorLabels" . | nindent 4 }}
    app.kubernetes.io/component: processor
{{- end }}
```

---

### Task 3.2.8: Create data-processor HPA
**File**: `helm/myhealth/templates/data-processor/hpa.yaml`

```yaml
{{- if .Values.dataProcessor.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: data-processor
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: data-processor
  
  minReplicas: {{ .Values.dataProcessor.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.dataProcessor.autoscaling.maxReplicas }}
  
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: {{ .Values.dataProcessor.autoscaling.targetCPUUtilizationPercentage }}
{{- end }}
```

---

### Task 3.2.9: Create api-service Deployment, Service, HPA
**Files**: Similar structure to data-processor

`helm/myhealth/templates/api-service/deployment.yaml` (similar to data-processor)
`helm/myhealth/templates/api-service/service.yaml` (similar to data-processor)
`helm/myhealth/templates/api-service/hpa.yaml` (similar to data-processor)

---

## Sprint 3.3: Istio Configuration

### Task 3.3.1: Create Istio Gateway
**File**: `helm/myhealth/templates/istio/gateway.yaml`

```yaml
{{- if .Values.istio.enabled }}
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: {{ .Values.istio.gateway.name }}
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: myhealth-tls
    hosts:
    {{- range .Values.istio.gateway.hosts }}
    - "{{ . }}"
    {{- end }}
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    {{- range .Values.istio.gateway.hosts }}
    - "{{ . }}"
    {{- end }}
{{- end }}
```

---

### Task 3.3.2: Create VirtualService for api-service
**File**: `helm/myhealth/templates/istio/virtualservice.yaml`

```yaml
{{- if .Values.istio.enabled }}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: api-service-vs
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
spec:
  hosts:
  {{- range .Values.istio.virtualServices.apiService.hosts }}
  - "{{ . }}"
  {{- end }}
  
  gateways:
  {{- range .Values.istio.virtualServices.apiService.gateways }}
  - "{{ . }}"
  {{- end }}
  
  http:
  - match:
    - uri:
        prefix: "/"
    route:
    - destination:
        host: api-service
        port:
          number: 8080
      weight: 100
    timeout: 30s
    retries:
      attempts: 3
      perTryTimeout: 10s
{{- end }}
```

---

### Task 3.3.3: Create DestinationRule
**File**: `helm/myhealth/templates/istio/destinationrule.yaml`

```yaml
{{- if .Values.istio.enabled }}
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: api-service
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
spec:
  host: api-service
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 100
        maxRequestsPerConnection: 2
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 30s
      baseEjectionTime: 30s
{{- end }}
```

---

## Sprint 3.4: Prometheus & Grafana Configuration

### Task 3.4.1: Create Prometheus ConfigMap
**File**: `helm/myhealth/templates/prometheus/configmap.yaml`

```yaml
{{- if .Values.prometheus.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
data:
  prometheus.yml: |
    global:
      scrape_interval: 30s
      evaluation_interval: 30s
    
    scrape_configs:
    - job_name: 'kubernetes-pods'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - {{ .Values.global.namespace }}
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
{{- end }}
```

---

### Task 3.4.2: Create Grafana Dashboards ConfigMap
**File**: `helm/myhealth/templates/grafana/configmap-dashboards.yaml`

```yaml
{{- if and .Values.grafana.enabled .Values.dashboards.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
    grafana_dashboard: "1"
data:
  oura-metrics.json: |
    {{- if .Values.dashboards.ouraMetrics }}
    {{- $.Files.Get "dashboards/oura-metrics.json" | nindent 4 }}
    {{- end }}
  
  service-health.json: |
    {{- if .Values.dashboards.serviceHealth }}
    {{- $.Files.Get "dashboards/service-health.json" | nindent 4 }}
    {{- end }}
{{- end }}
```

---

### Task 3.4.3: Create Grafana Datasources ConfigMap
**File**: `helm/myhealth/templates/grafana/configmap-datasources.yaml`

```yaml
{{- if .Values.grafana.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "myhealth.labels" . | nindent 4 }}
data:
  prometheus.yaml: |
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://prometheus-server:80
      isDefault: true
      access: proxy
{{- end }}
```

---

## Sprint 3.5: Helm Chart Validation

### Task 3.5.1: Lint Helm Chart
**Commands**:
```bash
cd helm/myhealth

# Lint chart
helm lint . --strict

# Template and check
helm template myhealth . --values values.yaml --debug
```

**Expected Output**:
```
No warnings or errors
```

---

### Task 3.5.2: Package Helm Chart
**Commands**:
```bash
cd helm/

# Package chart
helm package myhealth/

# Verify package
ls -lh myhealth-*.tgz
```

**Expected Output**: `myhealth-0.1.0.tgz` file created

---

## Phase 3 Completion Criteria

✅ All tasks completed when:
1. All Helm templates created
2. `helm lint` passes with no warnings
3. `helm template` output is valid Kubernetes manifests
4. Chart is packaged successfully
5. Values file properly documented
6. All placeholder values identified and documented

---

# PHASE 4: CI/CD PIPELINE SETUP
**Duration**: 3 days  
**Objective**: Automate builds, tests, and deployments

## Sprint 4.1: GitHub Actions - Terraform Workflows

### Task 4.1.1: Create Terraform Plan Workflow
**File**: `.github/workflows/terraform-plan.yml`

```yaml
name: 'Terraform Plan'

on:
  pull_request:
    paths:
      - 'terraform/**'
      - '.github/workflows/terraform-plan.yml'

jobs:
  terraform:
    name: 'Terraform Plan'
    runs-on: ubuntu-latest
    
    defaults:
      run:
        working-directory: ./terraform
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v2
      with:
        terraform_version: 1.6.0
    
    - name: Terraform Format
      run: terraform fmt -check -recursive
      continue-on-error: true
    
    - name: Terraform Init
      run: terraform init -backend=false
    
    - name: Terraform Validate
      run: terraform validate
    
    - name: Terraform Plan
      run: terraform plan -no-color
      env:
        TF_VAR_db_username: placeholder
        TF_VAR_db_password: placeholder
        TF_VAR_jwt_secret_key: placeholder
```

---

### Task 4.1.2: Create Terraform Apply Workflow
**File**: `.github/workflows/terraform-apply.yml`

```yaml
name: 'Terraform Apply'

on:
  push:
    branches:
      - main
    paths:
      - 'terraform/**'

jobs:
  terraform:
    name: 'Terraform Apply'
    runs-on: ubuntu-latest
    
    defaults:
      run:
        working-directory: ./terraform
    
    permissions:
      id-token: write
      contents: read
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v2
      with:
        terraform_version: 1.6.0
    
    - name: Configure AWS Credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
        aws-region: us-east-1
    
    - name: Terraform Init
      run: terraform init
    
    - name: Terraform Plan
      run: terraform plan -out=tfplan
      env:
        TF_VAR_db_username: ${{ secrets.DB_USERNAME }}
        TF_VAR_db_password: ${{ secrets.DB_PASSWORD }}
        TF_VAR_jwt_secret_key: ${{ secrets.JWT_SECRET_KEY }}
    
    - name: Terraform Apply
      run: terraform apply tfplan
      if: github.event_name == 'push'
```

---

## Sprint 4.2: GitHub Actions - Microservice Build Workflows

### Task 4.2.1: Create Go Build & Test Workflow
**File**: `.github/workflows/build-oura-collector.yml`

```yaml
name: 'Build oura-collector'

on:
  pull_request:
    paths:
      - 'services/oura-collector/**'
      - 'services/shared/**'
  push:
    branches:
      - main
    paths:
      - 'services/oura-collector/**'
      - 'services/shared/**'

jobs:
  test:
    name: 'Test'
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Go fmt
      run: |
        cd services
        go fmt ./...
      continue-on-error: true
    
    - name: Go vet
      run: |
        cd services
        go vet ./...
    
    - name: Go test
      run: |
        cd services
        go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./services/coverage.out
  
  build:
    name: 'Build & Push'
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    permissions:
      id-token: write
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Configure AWS Credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
        aws-region: us-east-1
    
    - name: Login to ECR
      run: |
        aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ${{ secrets.AWS_ACCOUNT_ID }}.dkr.ecr.us-east-1.amazonaws.com
    
    - name: Build Docker image
      run: |
        cd services/oura-collector
        docker build -t ${{ secrets.AWS_ACCOUNT_ID }}.dkr.ecr.us-east-1.amazonaws.com/myhealth/oura-collector:${{ github.sha }} .
        docker tag ${{ secrets.AWS_ACCOUNT_ID }}.dkr.ecr.us-east-1.amazonaws.com/myhealth/oura-collector:${{ github.sha }} ${{ secrets.AWS_ACCOUNT_ID }}.dkr.ecr.us-east-1.amazonaws.com/myhealth/oura-collector:latest
    
    - name: Push to ECR
      run: |
        docker push ${{ secrets.AWS_ACCOUNT_ID }}.dkr.ecr.us-east-1.amazonaws.com/myhealth/oura-collector:${{ github.sha }}
        docker push ${{ secrets.AWS_ACCOUNT_ID }}.dkr.ecr.us-east-1.amazonaws.com/myhealth/oura-collector:latest
    
    - name: Update Helm values
      run: |
        sed -i 's|tag: latest|tag: ${{ github.sha }}|g' helm/myhealth/values.yaml
    
    - name: Commit and push
      run: |
        git config user.name "GitHub Actions"
        git config user.email "actions@github.com"
        git add helm/myhealth/values.yaml
        git commit -m "Update oura-collector image tag" || true
        git push
```

**Instructions**:
1. Create similar workflows for `data-processor` and `api-service`
2. Change `oura-collector` references to respective service names
3. Update paths accordingly

---

## Sprint 4.3: Helm Validation

### Task 4.3.1: Create Helm Lint Workflow
**File**: `.github/workflows/helm-lint.yml`

```yaml
name: 'Helm Lint'

on:
  pull_request:
    paths:
      - 'helm/**'

jobs:
  helm-lint:
    name: 'Lint Helm Chart'
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Setup Helm
      uses: azure/setup-helm@v3
      with:
        version: 'v3.13.0'
    
    - name: Helm lint
      run: helm lint helm/myhealth --strict
    
    - name: Helm template
      run: |
        helm template myhealth helm/myhealth --values helm/myhealth/values.yaml > /tmp/manifest.yaml
        cat /tmp/manifest.yaml
    
    - name: Kubeval
      uses: instrumenta/kubeval-action@master
      with:
        files: /tmp/manifest.yaml
        strict: true
```

---

## Sprint 4.4: GitHub Secrets Setup

### Task 4.4.1: Add Secrets to GitHub
**Instructions**:
1. Go to repository Settings → Secrets and variables → Actions
2. Add these secrets:

| Secret Name | Value |
|-------------|-------|
| AWS_ROLE_ARN | arn:aws:iam::ACCOUNT_ID:role/github-actions |
| AWS_ACCOUNT_ID | Your AWS Account ID |
| DB_USERNAME | myhealth_user |
| DB_PASSWORD | (from terraform.tfvars) |
| JWT_SECRET_KEY | (from terraform.tfvars) |

---

## Phase 4 Completion Criteria

✅ All tasks completed when:
1. All GitHub Actions workflows created
2. Workflows validate without errors
3. Manual test runs pass successfully
4. GitHub secrets configured
5. PR validation triggers correctly
6. Container images build and push to ECR

---

# PHASE 5: MONITORING & OBSERVABILITY
**Duration**: 3 days  
**Objective**: Setup Prometheus + Grafana dashboards

## Sprint 5.1: Prometheus Configuration

### Task 5.1.1: Deploy Prometheus via Helm
**Commands**:
```bash
# Add Prometheus Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus as dependency (already in Helm chart)
cd helm/myhealth
helm dependency update

# Test deployment
helm template myhealth . | kubectl apply --dry-run=client -f -
```

---

## Sprint 5.2: Grafana Dashboards

### Task 5.2.1: Create Oura Metrics Dashboard
**File**: `helm/myhealth/dashboards/oura-metrics.json`

```json
{
  "dashboard": {
    "title": "Oura Metrics",
    "panels": [
      {
        "title": "Sleep Score (Last 7 days)",
        "targets": [
          {
            "expr": "sleep_metrics_score"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Activity Steps",
        "targets": [
          {
            "expr": "activity_metrics_steps"
          }
        ],
        "type": "stat"
      },
      {
        "title": "Readiness Score",
        "targets": [
          {
            "expr": "readiness_metrics_score"
          }
        ],
        "type": "gauge"
      }
    ]
  }
}
```

**Instructions**:
1. Create JSON dashboard file
2. This is a simplified example - expand with more panels as needed
3. Panels should match Prometheus metrics exposed by services

---

### Task 5.2.2: Create Service Health Dashboard
**File**: `helm/myhealth/dashboards/service-health.json`

```json
{
  "dashboard": {
    "title": "Service Health",
    "panels": [
      {
        "title": "HTTP Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~'5..'}[5m])"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Database Query Latency",
        "targets": [
          {
            "expr": "db_query_duration_seconds"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

---

## Sprint 5.3: Prometheus Alerting

### Task 5.3.1: Create Alert Rules
**File**: `helm/myhealth/templates/prometheus/alerts.yaml`

```yaml
{{- if .Values.prometheus.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-alerts
  namespace: {{ .Values.global.namespace }}
data:
  alerts.yml: |
    groups:
    - name: myhealth
      rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
      
      - alert: ServiceDown
        expr: up{job="myhealth"} == 0
        for: 2m
        annotations:
          summary: "Service is down"
      
      - alert: DatabaseHighLatency
        expr: db_query_duration_seconds > 1
        for: 5m
        annotations:
          summary: "Database queries are slow"
{{- end }}
```

---

## Phase 5 Completion Criteria

✅ All tasks completed when:
1. Prometheus deployed and collecting metrics
2. Grafana accessible with dashboards
3. Datasource configured correctly
4. Dashboards display live metrics
5. No missing metrics or errors

---

# PHASE 6: API GATEWAY & TESTING
**Duration**: 1 week  
**Objective**: Complete API Gateway setup and perform comprehensive testing

## Sprint 6.1: API Gateway Configuration

### Task 6.1.1: Create Custom Domain Mapping
**Commands**:
```bash
# Get ACM certificate for eric-n.com
aws acm request-certificate \
  --domain-name "*.eric-n.com" \
  --validation-method DNS \
  --region us-east-1

# Note the Certificate ARN
CERT_ARN="arn:aws:acm:us-east-1:ACCOUNT_ID:certificate/..."

# Add custom domain to API Gateway
aws apigatewayv2 create-domain-name \
  --domain-name api.eric-n.com \
  --domain-name-configurations DomainName=api.eric-n.com,CertificateArn=$CERT_ARN,EndpointType=REGIONAL \
  --region us-east-1
```

---

### Task 6.1.2: Create Route53 DNS Record
**Commands**:
```bash
# Get API Gateway distribution domain name
API_GATEWAY_DOMAIN=$(aws apigatewayv2 get-domain-name \
  --domain-name api.eric-n.com \
  --query 'DomainNameAttributes.DomainNameStatus' \
  --region us-east-1)

# Create Route53 record (assuming hosted zone exists)
aws route53 change-resource-record-sets \
  --hosted-zone-id HOSTED_ZONE_ID \
  --change-batch '{
    "Changes": [{
      "Action": "CREATE",
      "ResourceRecordSet": {
        "Name": "api.eric-n.com",
        "Type": "CNAME",
        "TTL": 300,
        "ResourceRecords": [{"Value": "'$API_GATEWAY_DOMAIN'"}]
      }
    }]
  }' \
  --region us-east-1
```

---

## Sprint 6.2: Deployment Testing

### Task 6.2.1: Deploy to Dev Cluster
**Commands**:
```bash
# Deploy Helm chart
helm install myhealth helm/myhealth \
  --namespace myhealth \
  --context myhealth-dev \
  --values helm/myhealth/values.yaml \
  --wait

# Verify deployment
kubectl --context myhealth-dev -n myhealth get all

# Check pod status
kubectl --context myhealth-dev -n myhealth get pods -o wide
```

---

### Task 6.2.2: Verify Services
**Commands**:
```bash
# Port forward to api-service
kubectl --context myhealth-dev -n myhealth port-forward svc/api-service 8080:8080 &

# Test API health
curl http://localhost:8080/health

# Test metrics
curl http://localhost:8080/metrics

# Test API endpoint
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/dashboard
```

---

### Task 6.2.3: Load Testing
**File**: `tests/load-test.js` (k6 format)

```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m30s', target: 100 },
    { duration: '20s', target: 0 },
  ],
};

export default function () {
  let response = http.get('http://localhost:8080/api/v1/dashboard');
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });
}
```

**Commands**:
```bash
# Install k6
brew install k6

# Run load test
k6 run tests/load-test.js
```

---

## Sprint 6.3: End-to-End Testing

### Task 6.3.1: Create Test Suite
**File**: `tests/e2e-test.sh`

```bash
#!/bin/bash

set -e

echo "=== End-to-End Test Suite ==="

# Test 1: Health checks
echo "Test 1: Health checks"
curl -f http://api.eric-n.com/health || exit 1
echo "✓ API service healthy"

# Test 2: Prometheus metrics
echo "Test 2: Prometheus metrics"
curl -f http://api.eric-n.com/metrics | grep -q "http_requests_total" || exit 1
echo "✓ Metrics available"

# Test 3: Data flow (collector -> processor -> API)
echo "Test 3: Data collection flow"
kubectl -n myhealth create job test-collector --from=cronjob/oura-collector
sleep 30
kubectl -n myhealth logs job/test-collector || true
echo "✓ Collection job completed"

# Test 4: Database connectivity
echo "Test 4: Database connectivity"
kubectl -n myhealth exec deploy/data-processor -- psql -h $DB_HOST -U $DB_USER -d myhealth -c "\dt" || exit 1
echo "✓ Database connected"

# Test 5: API endpoints
echo "Test 5: API endpoints"
TOKEN="test-token"
curl -f -H "Authorization: Bearer $TOKEN" http://api.eric-n.com/api/v1/dashboard || exit 1
echo "✓ API endpoints responding"

echo ""
echo "=== All Tests Passed ==="
```

---

## Sprint 6.4: Documentation & Runbook

### Task 6.4.1: Create Operations Runbook
**File**: `docs/operations-runbook.md`

```markdown
# Operations Runbook

## Accessing the Cluster

### kubectl access
\`\`\`bash
aws eks update-kubeconfig \
  --name myhealth \
  --region us-east-1 \
  --alias myhealth-dev

kubectl --context myhealth-dev get nodes
\`\`\`

## Monitoring

### Grafana access
\`\`\`bash
kubectl -n monitoring port-forward svc/grafana 3000:3000
# Open http://localhost:3000
# User: admin, Password: (see secrets)
\`\`\`

## Common Troubleshooting

### Pod stuck in pending
\`\`\`bash
kubectl describe pod <pod-name> -n myhealth
# Check resource requests vs node capacity
\`\`\`

### Database connection issues
\`\`\`bash
# Test connectivity
kubectl -n myhealth run -it --rm debug \
  --image=postgres:14 \
  --restart=Never -- \
  psql -h <RDS_ENDPOINT> -U myhealth_user -d myhealth
\`\`\`

## Backup & Restore

### Backup database
\`\`\`bash
aws rds create-db-snapshot \
  --db-instance-identifier myhealth-db \
  --db-snapshot-identifier myhealth-db-backup-$(date +%Y%m%d)
\`\`\`
```

---

## Phase 6 Completion Criteria

✅ All tasks completed when:
1. API Gateway deployed with custom domain
2. DNS records configured
3. Helm chart deployed successfully
4. All pods running and healthy
5. Load testing shows acceptable performance
6. End-to-end tests all passing
7. Operations documentation complete

---

# FINAL CHECKLIST: Project Completion

## Infrastructure ✓
- [ ] EKS cluster running in us-east-1
- [ ] RDS PostgreSQL accessible
- [ ] ECR repositories populated
- [ ] Secrets Manager configured
- [ ] Istio service mesh running
- [ ] External Secrets Operator running

## Microservices ✓
- [ ] oura-collector building and running
- [ ] data-processor building and running
- [ ] api-service building and running
- [ ] All services have >80% test coverage
- [ ] All services expose Prometheus metrics
- [ ] Docker images in ECR

## Deployment ✓
- [ ] Helm chart created and validated
- [ ] Chart deploys without errors
- [ ] All pods healthy and running
- [ ] Services accessible

## CI/CD ✓
- [ ] Terraform workflows working
- [ ] Go build/test workflows working
- [ ] Helm lint workflow working
- [ ] GitHub secrets configured
- [ ] Images pushing to ECR

## Monitoring ✓
- [ ] Prometheus collecting metrics
- [ ] Grafana dashboards visible
- [ ] Alerts configured
- [ ] Logs aggregated

## Testing ✓
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Load tests acceptable
- [ ] E2E tests all passing

## Documentation ✓
- [ ] Architecture documented
- [ ] Deployment guide written
- [ ] Operations runbook created
- [ ] API documentation complete

---

**Estimated Total Time**: 4-5 weeks  
**Next Phase**: Production deployment preparation

