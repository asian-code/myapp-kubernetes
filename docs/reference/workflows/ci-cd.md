# GitHub Actions Workflows Reference

## Overview

This document provides a quick reference for all GitHub Actions workflows in the Phase 4 CI/CD pipeline.

---

## Workflow Index

| Workflow | Trigger | Purpose | Environment |
|----------|---------|---------|-------------|
| `pr-validation.yml` | Every PR | Validate PR title & size | N/A |
| `terraform-plan.yml` | PR (terraform/) | Plan infrastructure changes | N/A |
| `terraform-apply.yml` | Main (terraform/) | Apply infrastructure changes | Production |
| `build-api-service.yml` | PR + Main | Test & build API service | Dev/Staging/Prod |
| `build-data-processor.yml` | PR + Main | Test & build data processor | Dev/Staging/Prod |
| `build-oura-collector.yml` | PR + Main | Test & build collector | Dev/Staging/Prod |
| `helm-validate.yml` | PR (helm/) | Validate Helm chart | N/A |
| `helm-deploy.yml` | Main (helm/svc/) | Deploy to all environments | Dev/Staging/Prod |
| `notify-ntfy.yml` | Called by others | Send notifications | External (ntfy) |

---

## Detailed Workflows

### 1. pr-validation.yml

**Trigger:** Every PR to main  
**Purpose:** Validate PR format and structure  

**Jobs:**
- Check conventional commit format in PR title
- Warn on large PRs (>50 files)
- Post summary comment

**Output:**
- ✅ or ⚠️ comment on PR
- Summary of running checks

---

### 2. terraform-plan.yml

**Trigger:** PR with changes to `terraform/**`  
**Purpose:** Validate Terraform changes without applying  

**Jobs:**
- `terraform` job:
  - Format check: `terraform fmt -check`
  - Validation: `terraform validate`
  - Plan: `terraform plan` with dummy vars
  - Comment summary on PR

**Inputs:** None (uses placeholder env vars)  
**Outputs:**
- Comment on PR with plan summary
- Artifact: tfplan (for later use)

**Permissions:**
- PR: read
- AWS: read-only

---

### 3. terraform-apply.yml

**Trigger:** Push to main with changes to `terraform/**`  
**Purpose:** Apply Terraform changes to AWS  

**Jobs:**
- `terraform-apply` job:
  - OIDC auth to AWS
  - Initialize backend (S3)
  - Plan: `terraform plan`
  - Apply: `terraform apply -auto-approve`
  - Send success/failure notification

**Inputs:** GitHub secrets (AWS_ROLE_ARN, DB_* vars)  
**Outputs:**
- Infrastructure provisioned in AWS
- ntfy notification
- GitHub Action run logs

**Environment:** `production` (separate environment for approval if added)

**Permissions:**
- AWS: AdministratorAccess equivalent
- Secrets: read

---

### 4. build-api-service.yml

**Trigger:**
- PR with changes to `services/api-service/**` or `services/shared/**`
- Push to main with same path changes

**Purpose:** Test and build API service Docker image  

**Jobs:**
- `test` job (runs on PR and push):
  - Go fmt check
  - Go vet analysis
  - Run tests: `go test -race -cover`
  - Upload coverage to codecov

- `build-and-push` job (runs only on main after test passes):
  - OIDC auth to AWS
  - Login to ECR
  - Build Docker image
  - Tag with: `git-sha` and `latest`
  - Push to ECR: `myhealth/api-service`
  - Send success/failure notification

**Inputs:**
- GitHub secrets (AWS_*)
- Services code

**Outputs:**
- Docker image in ECR
- Code coverage report
- ntfy notification

**Docker Registry:** `ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/myhealth/api-service:SHA`

---

### 5. build-data-processor.yml

**Trigger:**
- PR with changes to `services/data-processor/**` or `services/shared/**`
- Push to main with same path changes

**Purpose:** Test and build data processor Docker image  

**Jobs:** Same as `build-api-service.yml` but for data-processor

**Docker Registry:** `ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/myhealth/data-processor:SHA`

---

### 6. build-oura-collector.yml

**Trigger:**
- PR with changes to `services/oura-collector/**` or `services/shared/**`
- Push to main with same path changes

**Purpose:** Test and build Oura collector Docker image  

**Jobs:** Same as `build-api-service.yml` but for oura-collector

**Docker Registry:** `ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/myhealth/oura-collector:SHA`

---

### 7. helm-validate.yml

**Trigger:** PR with changes to `helm/**`  
**Purpose:** Validate Helm chart without deploying  

**Jobs:**
- `helm-lint` job:
  - Add Helm repositories (prometheus, grafana)
  - Run: `helm lint helm/myhealth --strict`
  - Template chart: `helm template myhealth helm/myhealth`
  - Validate manifests with kubeval
  - Comment results on PR

**Inputs:** Helm chart in git  
**Outputs:**
- ✅ or ❌ comment on PR
- No infrastructure changes

**Permissions:**
- PR: read
- No AWS access needed

---

### 8. helm-deploy.yml

**Trigger:** Push to main with changes to `helm/**` or `services/**`  
**Purpose:** Deploy Helm chart to dev, staging, and production environments  

**Jobs (Sequential):**

1. `prepare` job:
   - Sets image tag to git SHA
   - Outputs for downstream jobs

2. `deploy-dev` job:
   - Environment: `dev` (no approval)
   - Namespace: `myhealth-dev`
   - Log level: `debug`
   - Replicas: minimal
   - Auto-deploys
   - Waits for rollout (5m timeout)
   - Sends notification

3. `deploy-staging` job:
   - Environment: `staging` (no approval)
   - Runs after dev succeeds
   - Namespace: `myhealth-staging`
   - Log level: `info`
   - Replicas: medium
   - Auto-deploys
   - Waits for rollout (5m timeout)
   - Sends notification

4. `deploy-prod` job:
   - Environment: `production` (REQUIRES APPROVAL)
   - Runs after staging succeeds
   - Pauses for reviewer approval
   - Namespace: `myhealth-prod`
   - Log level: `warning`
   - Replicas: maximum
   - Deploys only after approval
   - Waits for rollout (10m timeout)
   - Health checks on API
   - Sends notification (high priority on success/failure)

**Inputs:**
- GitHub secrets (AWS_*)
- Helm chart from git
- Image tags from build workflows

**Outputs:**
- Kubernetes deployments in three environments
- Pod running in corresponding namespaces
- ntfy notifications (success/failure)

**Permissions:**
- AWS: EKS cluster access
- Kubernetes: deploy to namespaces

---

### 9. notify-ntfy.yml

**Type:** Reusable action  
**Trigger:** Called by other workflows  
**Purpose:** Send notifications to ntfy server  

**Inputs:**
- `ntfy_url`: Base URL (e.g., https://alert.eric-n.com)
- `topic`: Topic name (e.g., cicd)
- `title`: Notification title
- `message`: Notification body
- `priority`: 1-5 (optional, default 3)
- `tags`: CSV tags (optional)
- `actions`: JSON actions (optional)

**HTTP Request:**
```bash
curl -X POST \
  -H "Title: $title" \
  -H "Priority: $priority" \
  -H "Tags: $tags" \
  -d "$message" \
  "$ntfy_url/$topic"
```

**Output:** HTTP response from ntfy  
**Failure Handling:** Logs warning but doesn't fail workflow

---

## Workflow Dependencies

```
PR Created
├─→ pr-validation.yml (immediate)
├─→ terraform-plan.yml (if terraform/ changed)
├─→ build-*.yml (if services/ changed)
│   ├─→ test (immediate)
│   └─→ build-and-push (depends on test)
└─→ helm-validate.yml (if helm/ changed)

Merge to Main
├─→ terraform-apply.yml (if terraform/ changed)
├─→ build-*.yml (if services/ changed)
│   └─→ build-and-push (after test passes)
└─→ helm-deploy.yml (if helm/ or services/ changed)
    ├─→ prepare (immediate)
    ├─→ deploy-dev (after prepare)
    ├─→ deploy-staging (after deploy-dev succeeds)
    └─→ deploy-prod (after deploy-staging succeeds + APPROVAL)
```

---

## Environment Variables & Secrets

### Global Secrets (used by multiple workflows)
```
AWS_ROLE_ARN      - For OIDC auth to AWS
AWS_ACCOUNT_ID    - For ECR repository URI
NTFY_URL          - For notifications to ntfy server
```

### Service-Specific Secrets (for Terraform & Helm)
```
DB_USERNAME       - Database user
DB_PASSWORD       - Database password
JWT_SECRET_KEY    - JWT signing secret
```

### Workflow-Specific Variables
```
SERVICE_NAME      - Set in build-*.yml (api-service, etc.)
DOCKER_REGISTRY   - Constructed from AWS_ACCOUNT_ID
IMAGE_TAG         - Set to github.sha (git commit)
```

---

## Notifications Sent

### Build Workflows
- **Success**: "Build {service} - SUCCESS"
- **Failure**: "Build {service} - FAILED"

### Terraform Workflows
- **Success**: "Terraform Apply - SUCCESS"
- **Failure**: "Terraform Apply - FAILED"

### Helm Deploy Workflows
- **Dev Success**: "Deploy to Dev - SUCCESS"
- **Dev Failure**: "Deploy to Dev - FAILED"
- **Staging Success**: "Deploy to Staging - SUCCESS"
- **Staging Failure**: "Deploy to Staging - FAILED"
- **Prod Started**: "Production Deployment STARTED"
- **Prod Success**: "Production Deployment - SUCCESS ✅"
- **Prod Failure**: "Production Deployment - FAILED ❌"

All sent to: `https://alert.eric-n.com/cicd`

---

## Performance Characteristics

| Workflow | Duration | Parallelization |
|----------|----------|-----------------|
| pr-validation | <1 min | N/A (comments only) |
| terraform-plan | ~2 min | Single run |
| terraform-apply | ~2-3 min | Single run (locks state) |
| build-*-service (test) | ~3 min | 3 services in parallel |
| build-*-service (build) | ~5-10 min | Sequential (waits for test) |
| helm-validate | ~1 min | Single run |
| helm-deploy (dev) | ~5 min | Waits for rollout |
| helm-deploy (staging) | ~5 min | Sequential after dev |
| helm-deploy (prod) | ~10 min | Sequential after staging |

**Total PR cycle:** ~5-10 minutes  
**Total deployment cycle:** ~30-45 minutes

---

## Status Checks

All workflows create status checks visible on PR:

```
✅ pr-validation
✅ terraform-plan (if triggered)
✅ build-api-service (test + build)
✅ build-data-processor (test + build)
✅ build-oura-collector (test + build)
✅ helm-validate (if triggered)
```

All must pass before merge (if branch protection enabled).

---

## Troubleshooting by Workflow

### pr-validation issues
- **Problem**: Comment not appearing
- **Solution**: Check GitHub repo settings, verify workflows enabled

### terraform-plan issues
- **Problem**: "Terraform validate failed"
- **Solution**: Run `terraform validate` locally, fix syntax

### terraform-apply issues
- **Problem**: "AWS auth failed"
- **Solution**: Verify AWS_ROLE_ARN secret and OIDC provider

### build-* issues
- **Problem**: "Go tests failed"
- **Solution**: Run `go test ./...` locally, fix code

### helm-validate issues
- **Problem**: "helm lint failed"
- **Solution**: Run `helm lint helm/myhealth` locally

### helm-deploy issues
- **Problem**: "Pods not starting"
- **Solution**: Check pod logs, verify image exists in ECR

### notify issues
- **Problem**: "No notifications arriving"
- **Solution**: Verify NTFY_URL secret, test with curl

---

## Extending Workflows

To add new services/capabilities:

1. **New microservice**: Copy `build-api-service.yml`, update paths
2. **New Terraform module**: Terraform workflows auto-run
3. **New environment**: Add job in `helm-deploy.yml`
4. **New notification**: Update `notify-ntfy.yml` calls

See workflow files for examples.

---

## Security Considerations

- OIDC prevents long-lived AWS credentials
- Production approval required
- PR validation catches bad commits
- Image tags immutable (git SHA)
- Secrets encrypted in transit
- Audit trail in GitHub Actions logs

---

## Related Documentation

- `PHASE_4_QUICK_START.md` - Setup & overview
- `PHASE_4_CICD_IMPLEMENTATION.md` - Detailed configuration
- `PHASE_4_DEPLOYMENT_CHECKLIST.md` - Pre-deployment
- `ARGOCD_COMPATIBILITY.md` - Future GitOps integration

---

Last Updated: December 2025  
Workflows Version: 1.0  
GitHub Actions: ubuntu-latest
