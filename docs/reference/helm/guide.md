# Helm Guide - Complete Tutorial

## What is Helm?

**Helm** is the package manager for Kubernetes - think of it like `apt` for Ubuntu, `yum` for RedHat, or `npm` for Node.js, but for Kubernetes applications.

Instead of manually creating and managing dozens of YAML files for Deployments, Services, ConfigMaps, Secrets, etc., Helm lets you:
- Package everything into a single **Chart**
- Template values so you can reuse the same chart for dev/staging/prod
- Version and rollback deployments easily
- Share charts with others

---

## Core Concepts

### 1. **Chart**
A Helm chart is a collection of files that describe Kubernetes resources. Think of it as a blueprint.

```
helm/myhealth/          ← Your chart
├── Chart.yaml          ← Metadata (name, version, description)
├── values.yaml         ← Default configuration values
├── values-prod.yaml    ← Production-specific overrides
└── templates/          ← Kubernetes YAML templates
    ├── deployment.yaml
    ├── service.yaml
    └── ...
```

### 2. **Release**
A release is a running instance of a chart. You can install the same chart multiple times with different release names.

```bash
# Same chart, different releases
helm install myhealth-dev ./helm/myhealth -f values-dev.yaml
helm install myhealth-prod ./helm/myhealth -f values-prod.yaml
```

### 3. **Values**
Values are configuration variables that customize your chart. They're defined in `values.yaml` and can be overridden.

```yaml
# values.yaml
replicaCount: 2
image:
  repository: myapp
  tag: "1.0.0"
```

### 4. **Templates**
Templates are Kubernetes YAML files with Go templating syntax. Helm renders these with your values.

```yaml
# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Values.appName }}
spec:
  replicas: {{ .Values.replicaCount }}
```

---

## Installing Helm

```bash
# Windows (via Chocolatey)
choco install kubernetes-helm

# Windows (via Scoop)
scoop install helm

# macOS
brew install helm

# Linux
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Verify installation
helm version
```

---

## Essential Helm Commands

### **Repository Management**

```bash
# Add a chart repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts

# Update repositories (like apt update)
helm repo update

# List repositories
helm repo list

# Search for charts
helm search repo prometheus
helm search repo grafana
```

### **Working with Charts**

```bash
# Create a new chart
helm create mychart

# Validate chart syntax
helm lint ./helm/myhealth

# Package a chart into a .tgz file
helm package ./helm/myhealth

# Show chart information
helm show chart ./helm/myhealth
helm show values ./helm/myhealth
helm show all ./helm/myhealth
```

### **Installing & Upgrading**

```bash
# Install a chart (creates a release)
helm install RELEASE_NAME CHART_PATH [flags]

# Example: Install with default values
helm install myhealth ./helm/myhealth

# Install with custom values file
helm install myhealth ./helm/myhealth -f ./helm/myhealth/values-prod.yaml

# Install with inline value overrides
helm install myhealth ./helm/myhealth --set replicaCount=3

# Install in a specific namespace
helm install myhealth ./helm/myhealth -n myhealth-prod --create-namespace

# Dry-run (see what would be created without actually creating)
helm install myhealth ./helm/myhealth --dry-run --debug

# Upgrade an existing release
helm upgrade myhealth ./helm/myhealth

# Upgrade or install (install if not exists, upgrade if exists)
helm upgrade --install myhealth ./helm/myhealth

# Wait for resources to be ready
helm upgrade --install myhealth ./helm/myhealth --wait --timeout 10m
```

### **Managing Releases**

```bash
# List all releases
helm list
helm list -n myhealth-prod
helm list --all-namespaces

# Get release status
helm status myhealth
helm status myhealth -n myhealth-prod

# Get release history (shows all revisions)
helm history myhealth
helm history myhealth -n myhealth-prod

# Get values used in a release
helm get values myhealth
helm get values myhealth -n myhealth-prod

# Get all resources created by a release
helm get manifest myhealth
helm get manifest myhealth -n myhealth-prod
```

### **Rollbacks & Deletions**

```bash
# Rollback to previous revision
helm rollback myhealth
helm rollback myhealth -n myhealth-prod

# Rollback to specific revision
helm rollback myhealth 2
helm rollback myhealth 2 -n myhealth-prod

# Uninstall a release (delete all resources)
helm uninstall myhealth
helm uninstall myhealth -n myhealth-prod

# Uninstall but keep history
helm uninstall myhealth --keep-history
```

### **Templating & Debugging**

```bash
# Render templates locally (see generated YAML)
helm template myhealth ./helm/myhealth

# Render with specific values
helm template myhealth ./helm/myhealth -f ./helm/myhealth/values-prod.yaml

# Debug template rendering
helm template myhealth ./helm/myhealth --debug

# Render and output to file
helm template myhealth ./helm/myhealth > output.yaml

# Show only specific template
helm template myhealth ./helm/myhealth -s templates/deployment.yaml
```

---

## Working with Your myHealth Chart

### **Chart Structure**

```
helm/myhealth/
├── Chart.yaml                      # Chart metadata
├── values.yaml                     # Default values (dev/local)
├── values-prod.yaml                # Production overrides
├── dashboards/                     # Grafana dashboards
├── templates/
│   ├── _helpers.tpl               # Template helpers/functions
│   ├── namespace.yaml             # Namespace definition
│   ├── serviceaccount.yaml        # Service account
│   ├── externalsecrets.yaml       # External Secrets integration
│   ├── networkpolicies.yaml       # Network policies
│   ├── api-service/               # API service resources
│   ├── data-processor/            # Data processor resources
│   ├── oura-collector/            # Oura collector resources
│   ├── grafana/                   # Grafana deployment
│   ├── prometheus/                # Prometheus deployment
│   └── istio/                     # Istio configuration
```

### **Common Operations**

#### **1. Validate Your Chart**

```bash
# Lint the chart
helm lint helm/myhealth

# Dry-run install
helm install myhealth helm/myhealth --dry-run --debug -n myhealth-prod
```

#### **2. Deploy to Development**

```bash
# Install with default values (development)
helm upgrade --install myhealth ./helm/myhealth \
  --namespace myhealth-dev \
  --create-namespace \
  --wait --timeout 10m
```

#### **3. Deploy to Production**

```bash
# Install with production values
helm upgrade --install myhealth ./helm/myhealth \
  --namespace myhealth-prod \
  --create-namespace \
  --values ./helm/myhealth/values.yaml \
  --values ./helm/myhealth/values-prod.yaml \
  --set global.environment=prod \
  --set imageRegistry=211125604618.dkr.ecr.us-east-1.amazonaws.com \
  --set apiService.image.tag=abc123 \
  --wait --timeout 10m
```

#### **4. Update Just One Service**

```bash
# Update only the api-service image
helm upgrade myhealth ./helm/myhealth \
  -n myhealth-prod \
  --reuse-values \
  --set apiService.image.tag=new-sha-here
```

#### **5. Check What Changed**

```bash
# See what would change with an upgrade
helm diff upgrade myhealth ./helm/myhealth -n myhealth-prod

# Or use template to compare
helm template myhealth ./helm/myhealth -f values-prod.yaml > new.yaml
kubectl diff -f new.yaml
```

#### **6. Troubleshoot a Failed Deployment**

```bash
# Check release status
helm status myhealth -n myhealth-prod

# View release history
helm history myhealth -n myhealth-prod

# Get deployed manifest
helm get manifest myhealth -n myhealth-prod

# Check values used
helm get values myhealth -n myhealth-prod

# Check all info
helm get all myhealth -n myhealth-prod
```

#### **7. Rollback a Bad Deployment**

```bash
# Rollback to previous version
helm rollback myhealth -n myhealth-prod

# Rollback to specific revision
helm rollback myhealth 5 -n myhealth-prod
```

---

## Understanding Values

### **Values Hierarchy**

Helm merges values from multiple sources (highest priority first):

1. `--set` flags (command line)
2. `-f` / `--values` files (in order specified)
3. `values.yaml` (chart default)

```bash
# This combines all three
helm install myhealth ./helm/myhealth \
  -f ./helm/myhealth/values.yaml \
  -f ./helm/myhealth/values-prod.yaml \
  --set apiService.replicaCount=5
```

### **Viewing Effective Values**

```bash
# See what values would be used
helm template myhealth ./helm/myhealth \
  -f values-prod.yaml \
  --set apiService.replicaCount=5 \
  --show-only values.yaml

# Or after deployment
helm get values myhealth -n myhealth-prod
```

### **Common Value Overrides**

```bash
# Scale replicas
--set apiService.replicaCount=3

# Change image tag
--set apiService.image.tag=v1.2.3

# Change multiple values
--set apiService.replicaCount=3,dataProcessor.replicaCount=5

# Set nested values
--set apiService.resources.limits.memory=512Mi

# Set array values
--set apiService.env[0].name=DEBUG,apiService.env[0].value=true
```

---

## Template Functions & Logic

### **Common Template Syntax**

```yaml
# Access values
{{ .Values.apiService.replicaCount }}

# Access chart metadata
{{ .Chart.Name }}
{{ .Chart.Version }}

# Access release info
{{ .Release.Name }}
{{ .Release.Namespace }}

# Default values
{{ .Values.apiService.tag | default "latest" }}

# Conditionals
{{- if .Values.apiService.enabled }}
# ... resources ...
{{- end }}

# Loops
{{- range .Values.services }}
- name: {{ .name }}
{{- end }}

# Include templates
{{- include "myhealth.labels" . | nindent 4 }}

# String functions
{{ .Values.name | upper }}
{{ .Values.name | lower }}
{{ .Values.name | quote }}
{{ .Values.name | trim }}
```

### **Helpers (_helpers.tpl)**

Your chart uses helpers for reusable snippets:

```yaml
# Define in _helpers.tpl
{{- define "myhealth.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/version: {{ .Chart.Version }}
{{- end }}

# Use in templates
labels:
  {{- include "myhealth.labels" . | nindent 2 }}
```

---

## Helm Dependencies

Your chart can depend on other charts (like Prometheus, Grafana):

```yaml
# Chart.yaml
dependencies:
  - name: prometheus
    version: "15.x.x"
    repository: https://prometheus-community.github.io/helm-charts
  - name: grafana
    version: "6.x.x"
    repository: https://grafana.github.io/helm-charts
```

```bash
# Update dependencies
helm dependency update ./helm/myhealth

# This downloads charts to charts/ directory
helm dependency build ./helm/myhealth

# List dependencies
helm dependency list ./helm/myhealth
```

---

## Best Practices

### ✅ **DO**

1. **Use `--dry-run` before real deployments**
   ```bash
   helm upgrade --install myhealth ./helm/myhealth --dry-run
   ```

2. **Version your charts properly** (semantic versioning)
   ```yaml
   version: 1.2.3  # MAJOR.MINOR.PATCH
   ```

3. **Use `--wait` for production deployments**
   ```bash
   helm upgrade --install myhealth ./helm/myhealth --wait --timeout 10m
   ```

4. **Separate values files per environment**
   ```
   values.yaml       # defaults
   values-dev.yaml
   values-prod.yaml
   ```

5. **Use `--atomic` for all-or-nothing deployments**
   ```bash
   helm upgrade --install myhealth ./helm/myhealth --atomic
   ```

6. **Always specify namespace**
   ```bash
   helm install myhealth ./helm/myhealth -n myhealth-prod
   ```

7. **Test templates locally**
   ```bash
   helm template myhealth ./helm/myhealth | kubectl apply --dry-run=client -f -
   ```

### ❌ **DON'T**

1. **Don't skip validation**
   - Always run `helm lint` before deploying

2. **Don't hardcode values in templates**
   - Use values.yaml instead

3. **Don't forget to update dependencies**
   - Run `helm dependency update` when Chart.yaml changes

4. **Don't use `latest` tags in production**
   - Pin specific versions

5. **Don't deploy without testing**
   - Use `--dry-run` first

---

## Troubleshooting

### **Helm Install/Upgrade Hangs**

```bash
# Check pod status
kubectl get pods -n myhealth-prod

# Check events
kubectl get events -n myhealth-prod --sort-by='.lastTimestamp'

# Describe problematic pod
kubectl describe pod POD_NAME -n myhealth-prod

# Check logs
kubectl logs POD_NAME -n myhealth-prod
```

### **Template Errors**

```bash
# Render templates to see errors
helm template myhealth ./helm/myhealth --debug

# Validate with kubectl
helm template myhealth ./helm/myhealth | kubectl apply --dry-run=client -f -
```

### **Values Not Applied**

```bash
# Check effective values
helm get values myhealth -n myhealth-prod

# Check all values (including defaults)
helm get values myhealth -n myhealth-prod --all
```

### **Release in Bad State**

```bash
# Check status
helm status myhealth -n myhealth-prod

# View history
helm history myhealth -n myhealth-prod

# Rollback
helm rollback myhealth -n myhealth-prod

# Force delete if stuck
helm uninstall myhealth -n myhealth-prod --no-hooks
```

---

## Quick Reference Cheat Sheet

```bash
# INSTALL
helm install RELEASE CHART -n NAMESPACE --create-namespace

# UPGRADE
helm upgrade RELEASE CHART -n NAMESPACE

# UPGRADE OR INSTALL
helm upgrade --install RELEASE CHART -n NAMESPACE

# LIST
helm list -n NAMESPACE

# STATUS
helm status RELEASE -n NAMESPACE

# ROLLBACK
helm rollback RELEASE -n NAMESPACE

# UNINSTALL
helm uninstall RELEASE -n NAMESPACE

# TEMPLATE (dry-run)
helm template RELEASE CHART -f values.yaml

# GET VALUES
helm get values RELEASE -n NAMESPACE

# GET MANIFEST
helm get manifest RELEASE -n NAMESPACE

# HISTORY
helm history RELEASE -n NAMESPACE
```

---

## Real-World Examples for Your Project

### **Deploy Everything Fresh**

```bash
# Deploy to production
helm upgrade --install myhealth ./helm/myhealth \
  --namespace myhealth-prod \
  --create-namespace \
  --values ./helm/myhealth/values.yaml \
  --values ./helm/myhealth/values-prod.yaml \
  --set global.environment=prod \
  --set imageRegistry=$AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com \
  --set apiService.image.tag=$GIT_SHA \
  --set dataProcessor.image.tag=$GIT_SHA \
  --set ouraCollector.image.tag=$GIT_SHA \
  --wait --timeout 10m
```

### **Update Just One Service After Build**

```bash
# After building new api-service image
helm upgrade myhealth ./helm/myhealth \
  -n myhealth-prod \
  --reuse-values \
  --set apiService.image.tag=abc123def
```

### **Scale Services**

```bash
# Scale up for high traffic
helm upgrade myhealth ./helm/myhealth \
  -n myhealth-prod \
  --reuse-values \
  --set apiService.replicaCount=5 \
  --set dataProcessor.replicaCount=3
```

### **Debug Before Deploying**

```bash
# See what would be created
helm template myhealth ./helm/myhealth \
  -f ./helm/myhealth/values-prod.yaml \
  --set apiService.image.tag=test123 \
  | less

# Apply to cluster without Helm (for testing)
helm template myhealth ./helm/myhealth \
  -f ./helm/myhealth/values-prod.yaml \
  | kubectl apply --dry-run=client -f -
```

---

## Next Steps

1. **Practice with your chart:**
   ```bash
   helm template myhealth ./helm/myhealth | less
   ```

2. **Try a local install:**
   ```bash
   minikube start
   helm install myhealth ./helm/myhealth
   ```

3. **Explore Helm Hub:**
   - Visit https://artifacthub.io/
   - Browse popular charts
   - Learn from their structure

4. **Read official docs:**
   - https://helm.sh/docs/

5. **Experiment with values:**
   - Modify `values.yaml`
   - Run `helm template` to see changes
   - Deploy to test namespace

---

## Additional Resources

- **Helm Documentation**: https://helm.sh/docs/
- **Helm Best Practices**: https://helm.sh/docs/chart_best_practices/
- **Artifact Hub**: https://artifacthub.io/ (search for charts)
- **Helm GitHub**: https://github.com/helm/helm
- **Chart Template Guide**: https://helm.sh/docs/chart_template_guide/

---

**Pro Tip**: Start with `helm template` to understand what your chart generates before deploying. It's the safest way to learn! 🚀
