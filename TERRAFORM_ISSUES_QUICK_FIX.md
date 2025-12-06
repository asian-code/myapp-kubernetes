# Terraform Issues Summary & Fix Guide

## Quick Reference: Issues Found

```
┌─────────────────────────────────────────────────────────────────┐
│  CRITICAL ISSUES (Must fix before terraform apply)              │
├─────────────────────────────────────────────────────────────────┤
│ 1. Password generation conflict (main.tf vs modules)            │
│ 2. API Gateway domain hardcoded                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  HIGH PRIORITY ISSUES (Fix before production)                   │
├─────────────────────────────────────────────────────────────────┤
│ 3. Redundant password generation in secrets-manager             │
│ 4. Missing variable validation                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  MEDIUM PRIORITY ISSUES (Nice to have)                          │
├─────────────────────────────────────────────────────────────────┤
│ 5. Single NAT gateway (HA concern for prod)                     │
│ 6. Missing output descriptions                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Issue #1: Password Generation Conflict

### The Problem

```
THREE password generation paths:

1. main.tf generates:
   random_password.db_password → 32 chars
   
2. RDS module receives:
   var.db_password (from main.tf) → 32 chars
   
3. Secrets-manager module generates:
   random_password.db_master_password → 40 chars
   AND receives var.db_password (from main.tf) → 32 chars

RESULT: Confusion about which password is actually used
```

### Current Flow (WRONG)

```
main.tf
├─ Creates: random_password.db_password (32 chars)
├─ Passes to RDS: random_password.db_password.result
├─ Passes to Secrets: random_password.db_password.result
│
RDS Module
├─ Receives: var.db_password (32 chars) ✓
└─ Uses this for database
│
Secrets Manager Module  ❌ PROBLEM HERE
├─ Receives: var.db_password (32 chars)
├─ Creates: random_password.db_master_password (40 chars) [UNUSED]
└─ Line 62: Uses var.db_password if not null, else random_password
   (It uses the 32-char password, so the 40-char one is wasted)
```

### Solution (RECOMMENDED)

```
Consolidate ALL secret generation in main.tf:

main.tf
├─ Creates: random_password.db_password (32 chars)
├─ Creates: random_password.jwt_secret (64 chars)
├─ Creates: random_password.oura_api_key (40 chars)
├─ Passes all to RDS module
└─ Passes all to Secrets Manager module

RDS Module
├─ Receives: var.db_password (32 chars)
└─ Uses this for database

Secrets Manager Module
├─ Receives: var.db_password (32 chars)
├─ Receives: var.jwt_secret_key (64 chars)
├─ Receives: var.oura_api_key (40 chars)
└─ Stores all three securely
```

### Code Changes Required

**In `terraform/main.tf`:**
- Keep password generation (already there ✓)
- No changes needed

**In `terraform/modules/secrets-manager/main.tf`:**
- DELETE lines 6-8 (random_password.db_master_password generation)
- DELETE lines 13-17 (random_password.jwt_secret generation)
- Keep only random_password.db_master_password if truly needed for backwards compatibility

**In `terraform/modules/secrets-manager/variables.tf`:**
- Ensure jwt_secret_key is marked as required
- Ensure db_password is marked as required

---

## Issue #2: API Gateway Domain Hardcoded

### The Problem

```hcl
# Current (WRONG):
domain_name = "myhealth.eric-n.com"  # Hardcoded!

# Should be (CORRECT):
domain_name = "api.myhealth.eric-n.com"  # From PROJECT_PLAN.md
# AND should be a variable, not hardcoded
```

### Current State

```
terraform/modules/api-gateway/main.tf (lines 24-27):

domain_name                = "myhealth.eric-n.com"  ❌
domain_name_certificate_arn = var.acm_certificate_arn
create_certificate         = false
create_domain_records      = false  ← Manual Route53 required

Problems:
1. Wrong domain (should be api.myhealth.eric-n.com)
2. Not a variable (can't change per environment)
3. Manual Route53 setup required (documented?)
```

### Solution

**Step 1: Add variable to `terraform/modules/api-gateway/variables.tf`:**

```hcl
variable "domain_name" {
  description = "Custom domain name for API Gateway"
  type        = string
  default     = ""
}

variable "create_domain_records" {
  description = "Create Route53 records automatically"
  type        = bool
  default     = false
}
```

**Step 2: Update `terraform/modules/api-gateway/main.tf`:**

```hcl
# Replace lines 24-27 with:

domain_name                = var.domain_name != "" ? var.domain_name : null
domain_name_certificate_arn = var.acm_certificate_arn != "" ? var.acm_certificate_arn : null
create_domain_records      = var.create_domain_records
```

**Step 3: Update `terraform/main.tf` module call:**

```hcl
module "api_gateway" {
  count = var.acm_certificate_arn != "" && var.backend_url != "" ? 1 : 0

  source              = "./modules/api-gateway"
  
  domain_name         = "api.eric-n.com"  # ← Add this
  backend_url         = var.backend_url
  acm_certificate_arn = var.acm_certificate_arn
  tags                = local.tags
}
```

---

## Issue #3: Redundant Password Generation

### The Problem

Two password resources are created but only one is used:

```hcl
# In secrets-manager/main.tf:

# Created: UNUSED (40 chars)
resource "random_password" "db_master_password" {
  length           = 40
  special          = true
  override_special = "@#%&*()-_=+[]{}<>:?"
}

# Created: UNUSED (32 chars)
resource "random_password" "jwt_secret" {
  length           = 32
  special          = true
  override_special = "@#%&*()-_=+[]{}<>:?"
}

# Used: Line 62
password = var.db_password != null ? var.db_password : random_password.db_master_password.result
↑
Gets var.db_password from main.tf, never needs the locally generated one!
```

### Solution

**Delete from `terraform/modules/secrets-manager/main.tf`:**

```hcl
# Remove these lines (6-8):
resource "random_password" "db_master_password" {
  length           = 40
  special          = true
  override_special = "@#%&*()-_=+[]{}<>:?"
}

# Remove these lines (13-17):
resource "random_password" "jwt_secret" {
  length           = 32
  special          = true
  override_special = "@#%&*()-_=+[]{}<>:?"
}

# Keep the secret_version resources as-is:
resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = var.db_username  ← Uses variable
    password = var.db_password  ← Uses variable (from main.tf)
    host     = var.db_host
    port     = 5432
    dbname   = "myhealth"
  })
}
```

---

## Issue #4: Missing Variable Validation

### The Problem

Variables accept any value without validation:

```hcl
# Current (WRONG):
variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "myhealth"
  # Can be anything!
}

# Someone could set: cluster_name = "INVALID@#$%"
# Terraform would accept it!
```

### Solution

**Update `terraform/variables.tf` with validation blocks:**

```hcl
variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "myhealth"
  
  validation {
    condition     = can(regex("^[a-z0-9-]{1,64}$", var.cluster_name))
    error_message = "Cluster name must be lowercase alphanumeric with hyphens (max 64 chars)"
  }
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
  
  validation {
    condition     = can(regex("^db\\.", var.rds_instance_class))
    error_message = "RDS instance class must start with 'db.'"
  }
}

variable "node_instance_types" {
  description = "EKS node instance types"
  type        = list(string)
  default     = ["t3.medium"]
  
  validation {
    condition     = alltrue([for t in var.node_instance_types : can(regex("^[a-z]\\d[a-z]", t))])
    error_message = "Invalid EC2 instance types"
  }
}
```

---

## Issue #5: Single NAT Gateway

### The Problem

```hcl
# In networking/main.tf:
single_nat_gateway = true  # Only one NAT gateway

Issues:
1. Single point of failure (if NAT fails, no internet)
2. Bottleneck for all traffic
3. Not HA for production

This is OK for dev but NOT for production!
```

### Solution

**Add flexibility variable to `terraform/variables.tf`:**

```hcl
variable "enable_nat_ha" {
  description = "Enable high availability NAT (one per AZ) vs cost-effective (single)"
  type        = bool
  default     = false  # Dev: single NAT (cost saving)
}
```

**Update `terraform/modules/networking/variables.tf`:**

```hcl
variable "enable_nat_ha" {
  description = "Enable high availability NAT"
  type        = bool
  default     = false
}
```

**Update `terraform/modules/networking/main.tf`:**

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 6.5"

  # ...existing config...

  single_nat_gateway = !var.enable_nat_ha  # false (single) for dev, true (ha) for prod

  # ...rest of config...
}
```

**Update `terraform/main.tf`:**

```hcl
module "networking" {
  source = "./modules/networking"

  cluster_name          = var.cluster_name
  vpc_cidr              = var.vpc_cidr
  enable_nat_ha         = var.environment == "prod" ? true : false
  tags                  = local.tags
  log_retention_in_days = var.log_retention_in_days
}
```

---

## Issue #6: Missing Output Descriptions

### The Problem

```hcl
# Current (WRONG):
output "rds_endpoint" {
  value = module.rds.db_endpoint
  # No description!
}

# terraform output gives no helpful information about what this means
```

### Solution

**Update `terraform/outputs.tf`:**

```hcl
output "vpc_id" {
  description = "VPC ID"
  value       = module.networking.vpc_id
}

output "private_subnets" {
  description = "Private subnet IDs"
  value       = module.networking.private_subnets
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  description = "EKS cluster API endpoint (HTTPS)"
  value       = module.eks.cluster_endpoint
  sensitive   = false
}

output "eks_oidc_provider_arn" {
  description = "OIDC provider ARN for IRSA"
  value       = module.eks.oidc_provider_arn
  sensitive   = false
}

output "rds_endpoint" {
  description = "RDS PostgreSQL endpoint (hostname:port format)"
  value       = module.rds.db_endpoint
  sensitive   = false
}

output "rds_database_name" {
  description = "RDS database name"
  value       = "myhealth"
}

output "ecr_registry_id" {
  description = "AWS account ID (ECR registry)"
  value       = module.ecr.registry_id
}

output "ecr_repositories" {
  description = "ECR repository URLs"
  value = {
    oura_collector  = module.ecr.oura_collector_url
    data_processor  = module.ecr.data_processor_url
    api_service     = module.ecr.api_service_url
  }
}

output "tf_state_bucket" {
  description = "S3 bucket for Terraform state (local for now)"
  value       = aws_s3_bucket.tf_state.bucket
}

output "api_gateway_endpoint" {
  description = "API Gateway invoke URL (or null if not deployed)"
  value       = try(module.api_gateway[0].api_endpoint, null)
  sensitive   = false
}

output "secrets_manager_secrets" {
  description = "Secrets created in AWS Secrets Manager"
  value = {
    oura_api_key = "myhealth/oura-api-key"
    db_credentials = "myhealth/db-credentials"
    jwt_secret = "myhealth/jwt-secret"
  }
}
```

---

## Implementation Priority & Timeline

### Phase 1: CRITICAL (Do before `terraform apply`)
- [ ] **Fix password generation** - 30 min
  - Consolidate to main.tf
  - Remove from modules
  
- [ ] **Fix API Gateway domain** - 15 min
  - Add variable
  - Update main.tf call

**Estimated time: 45 minutes**

### Phase 2: HIGH (Do before production)
- [ ] **Add variable validation** - 45 min
- [ ] **Add output descriptions** - 20 min

**Estimated time: 65 minutes**

### Phase 3: MEDIUM (Nice to have)
- [ ] **Add NAT HA variable** - 20 min
- [ ] **Extract locals.tf** - 10 min

**Estimated time: 30 minutes**

---

## Summary: Before Running `terraform apply`

```
✅ To Do:
[ ] Fix password generation conflict
[ ] Fix API Gateway domain configuration
[ ] Verify AWS credentials configured
[ ] Review terraform plan output for costs
[ ] Confirm you're ready to deploy

❌ Do NOT apply until:
- Password generation is consolidated
- API Gateway domain is configurable
```

---

## Questions to Answer

**Q1**: Should we fix all issues now, or just critical ones?  
**A**: Fix critical + high priority before apply. Medium priority can wait.

**Q2**: When should we migrate from local backend to S3?  
**A**: After first successful apply, before second apply.

**Q3**: Do you want me to make these changes, or review my changes?  
**A**: Tell me and I'll implement immediately.

