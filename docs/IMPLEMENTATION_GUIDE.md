# myHealth - Implementation Guide
## Detailed Instructions for AI Agent Execution

**Project**: myHealth Oura Ring Kubernetes Application  
**Tech Stack**: EKS + Terraform + Go + Helm + Prometheus + Grafana + ArgoCD + GitHub Actions  
**Timeline**: 4-5 weeks (simplified dev environment)  
**Region**: us-east-1  
**Domain**: eric-n.com (api.myhealth.eric-n.com)

---

## Table of Contents
1. [Phase 1: Infrastructure Setup](#phase-1-infrastructure-setup)
2. [Phase 2: Microservices Development](#phase-2-microservices-development)
3. [Phase 3: Helm Chart Creation](#phase-3-helm-chart-creation)
4. [Phase 4: CI/CD Pipeline Setup](#phase-4-cicd-pipeline-setup)
5. [Phase 5: Monitoring & Observability](#phase-5-monitoring--observability)
6. [Phase 6: API Gateway & Testing](#phase-6-api-gateway--testing)
7. [General Guidelines](#general-guidelines)

---

# PHASE 1: Infrastructure Setup
**Duration**: 1 week  
**Objective**: Refactor existing Terraform code, add missing modules, provision dev EKS cluster

## Sprint 1.1: Terraform Structure Reorganization

### Task 1.1.1: Create Module Directory Structure
**What**: Organize Terraform into reusable modules  
**Where**: `terraform/modules/`  
**Expected Output**: Directory structure as below

```bash
terraform/
├── modules/
│   ├── eks/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── addons.tf
│   │   └── irsa.tf
│   ├── rds/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── ecr/
│   │   ├── main.tf
│   │   └── outputs.tf
│   ├── api-gateway/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── secrets-manager/
│   │   ├── main.tf
│   │   └── outputs.tf
│   └── networking/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
├── dev/
│   ├── main.tf
│   ├── terraform.tfvars
│   ├── backend.tf
│   └── variables.tf
├── main.tf
├── variables.tf
├── outputs.tf
└── versions.tf
```

**Instructions**:
1. Create all directories under `terraform/modules/`
2. Keep existing `networking.tf` content, but will refactor into `modules/networking/`
3. Don't move files yet - just create structure

**Validation**: Run `ls -R terraform/modules/` to verify structure exists

---

### Task 1.1.2: Extract Networking Module
**What**: Move VPC/networking code to proper module  
**Source**: Current `eks-three-tier-app/terraform/networking.tf`  
**Target**: `terraform/modules/networking/main.tf`

**Instructions**:

1. Create `terraform/modules/networking/main.tf` with:
   - Copy entire `vpc` module configuration from `networking.tf`
   - Copy security group configuration
   - Include all locals and data sources

2. Create `terraform/modules/networking/variables.tf` with:
   ```hcl
   variable "cluster_name" {
     type = string
   }
   
   variable "vpc_cidr" {
     type = string
   }
   
   variable "tags" {
     type = map(string)
   }
   
   variable "log_retention_in_days" {
     type = number
   }
   ```

3. Create `terraform/modules/networking/outputs.tf` with:
   ```hcl
   output "vpc_id" {
     value = module.vpc.vpc_id
   }
   
   output "private_subnets" {
     value = module.vpc.private_subnets
   }
   
   output "public_subnets" {
     value = module.vpc.public_subnets
   }
   
   output "node_security_group_id" {
     value = aws_security_group.additional_node_sg.id
   }
   ```

**Validation**: 
- Check file exists and contains VPC module
- Verify outputs are exported

---

### Task 1.1.3: Create EKS Module
**What**: Extract EKS cluster configuration into module  
**Source**: `eks-three-tier-app/terraform/k8s.tf` (uncomment and refactor)  
**Target**: `terraform/modules/eks/main.tf`

**Instructions**:

1. Create `terraform/modules/eks/main.tf`:
   ```hcl
   module "eks" {
     source  = "terraform-aws-modules/eks/aws"
     version = "~> 21.0.0"
     
     cluster_name    = var.cluster_name
     cluster_version = var.cluster_version
     
     vpc_id                   = var.vpc_id
     subnet_ids               = var.private_subnets
     control_plane_subnet_ids = var.private_subnets
     
     cluster_endpoint_public_access       = true
     cluster_endpoint_private_access      = true
     cluster_endpoint_public_access_cidrs = ["0.0.0.0/0"]  # Restrict in prod
     
     enable_irsa = true
     
     eks_managed_node_groups = {
       general = {
         name            = "general-group"
         instance_types  = ["t3.medium"]
         min_size        = 2
         max_size        = 10
         desired_size    = 2
         capacity_type   = "SPOT"
         
         launch_template_name = "${var.cluster_name}-general"
         
         update_config = {
           max_unavailable_percentage = 25
         }
         
         vpc_security_group_ids = [var.node_security_group_id]
         
         tags = merge(var.tags, {
           "NodeGroup" = "general"
         })
       }
     }
     
     cluster_addons = {
       coredns = {
         most_recent = true
       }
       kube-proxy = {
         most_recent = true
       }
       vpc-cni = {
         most_recent = true
         configuration_values = jsonencode({
           env = {
             ENABLE_PREFIX_DELEGATION = "true"
           }
         })
       }
       ebs-csi-driver = {
         most_recent = true
       }
     }
     
     cluster_enabled_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
     cloudwatch_log_group_retention_in_days = var.log_retention_in_days
     create_cloudwatch_log_group = true
     
     tags = var.tags
   }
   ```

2. Create `terraform/modules/eks/variables.tf`:
   ```hcl
   variable "cluster_name" {
     type = string
   }
   
   variable "cluster_version" {
     type = string
     default = "1.28"
   }
   
   variable "vpc_id" {
     type = string
   }
   
   variable "private_subnets" {
     type = list(string)
   }
   
   variable "node_security_group_id" {
     type = string
   }
   
   variable "log_retention_in_days" {
     type = number
     default = 7
   }
   
   variable "tags" {
     type = map(string)
   }
   ```

3. Create `terraform/modules/eks/outputs.tf`:
   ```hcl
   output "cluster_endpoint" {
     value = module.eks.cluster_endpoint
   }
   
   output "cluster_name" {
     value = module.eks.cluster_name
   }
   
   output "cluster_version" {
     value = module.eks.cluster_version
   }
   
   output "cluster_certificate_authority_data" {
     value = module.eks.cluster_certificate_authority_data
     sensitive = true
   }
   
   output "oidc_provider_arn" {
     value = module.eks.oidc_provider_arn
   }
   ```

4. Create `terraform/modules/eks/irsa.tf` (IAM Roles for Service Accounts):
   ```hcl
   # This file will be used for IRSA policies later
   # Placeholder for now
   ```

**Validation**:
- Files exist and contain proper module structure
- All variables defined
- All outputs exported

---

### Task 1.1.4: Create RDS Module
**What**: New PostgreSQL database for Oura metrics  
**Target**: `terraform/modules/rds/main.tf`

**Instructions**:

1. Create `terraform/modules/rds/main.tf`:
   ```hcl
   resource "aws_db_subnet_group" "myhealth" {
     name       = "${var.cluster_name}-db-subnet-group"
     subnet_ids = var.private_subnets
     
     tags = merge(var.tags, {
       Name = "${var.cluster_name}-db-subnet-group"
     })
   }
   
   resource "aws_security_group" "rds" {
     name        = "${var.cluster_name}-rds-sg"
     description = "Security group for RDS"
     vpc_id      = var.vpc_id
     
     ingress {
       from_port       = 5432
       to_port         = 5432
       protocol        = "tcp"
       security_groups = [var.eks_node_security_group_id]
     }
     
     egress {
       from_port   = 0
       to_port     = 0
       protocol    = "-1"
       cidr_blocks = ["0.0.0.0/0"]
     }
     
     tags = var.tags
   }
   
   resource "aws_db_instance" "myhealth" {
     identifier              = "${var.cluster_name}-db"
     engine                  = "postgres"
     engine_version          = "14.10"
     instance_class          = var.instance_class
     allocated_storage       = 20
     storage_type            = "gp3"
     storage_encrypted       = true
     
     db_name  = "myhealth"
     username = var.db_username
     password = var.db_password
     
     db_subnet_group_name            = aws_db_subnet_group.myhealth.name
     vpc_security_group_ids          = [aws_security_group.rds.id]
     publicly_accessible             = false
     
     multi_az            = var.multi_az
     backup_retention_period = var.backup_retention_days
     backup_window           = "03:00-04:00"
     maintenance_window      = "mon:04:00-mon:05:00"
     
     skip_final_snapshot       = var.skip_final_snapshot
     final_snapshot_identifier = "${var.cluster_name}-db-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
     
     tags = merge(var.tags, {
       Name = "${var.cluster_name}-db"
     })
   }
   ```

2. Create `terraform/modules/rds/variables.tf`:
   ```hcl
   variable "cluster_name" {
     type = string
   }
   
   variable "vpc_id" {
     type = string
   }
   
   variable "private_subnets" {
     type = list(string)
   }
   
   variable "eks_node_security_group_id" {
     type = string
   }
   
   variable "instance_class" {
     type    = string
     default = "db.t3.micro"
   }
   
   variable "db_username" {
     type      = string
     sensitive = true
   }
   
   variable "db_password" {
     type      = string
     sensitive = true
   }
   
   variable "multi_az" {
     type    = bool
     default = false
   }
   
   variable "backup_retention_days" {
     type    = number
     default = 7
   }
   
   variable "skip_final_snapshot" {
     type    = bool
     default = false
   }
   
   variable "tags" {
     type = map(string)
   }
   ```

3. Create `terraform/modules/rds/outputs.tf`:
   ```hcl
   output "db_endpoint" {
     value = aws_db_instance.myhealth.endpoint
   }
   
   output "db_name" {
     value = aws_db_instance.myhealth.db_name
   }
   
   output "db_username" {
     value     = aws_db_instance.myhealth.username
     sensitive = true
   }
   ```

**Validation**:
- Files exist
- Security group allows ingress from EKS nodes on port 5432

---

### Task 1.1.5: Create ECR Module
**What**: Container registries for 3 microservices  
**Target**: `terraform/modules/ecr/main.tf`

**Instructions**:

1. Create `terraform/modules/ecr/main.tf`:
   ```hcl
   resource "aws_ecr_repository" "oura_collector" {
     name                 = "myhealth/oura-collector"
     image_tag_mutability = "MUTABLE"
     image_scanning_configuration {
       scan_on_push = true
     }
     encryption_configuration {
       encryption_type = "AES256"
     }
     tags = var.tags
   }
   
   resource "aws_ecr_repository" "data_processor" {
     name                 = "myhealth/data-processor"
     image_tag_mutability = "MUTABLE"
     image_scanning_configuration {
       scan_on_push = true
     }
     encryption_configuration {
       encryption_type = "AES256"
     }
     tags = var.tags
   }
   
   resource "aws_ecr_repository" "api_service" {
     name                 = "myhealth/api-service"
     image_tag_mutability = "MUTABLE"
     image_scanning_configuration {
       scan_on_push = true
     }
     encryption_configuration {
       encryption_type = "AES256"
     }
     tags = var.tags
   }
   
   resource "aws_ecr_lifecycle_policy" "retention" {
     for_each = {
       oura_collector = aws_ecr_repository.oura_collector.name
       data_processor = aws_ecr_repository.data_processor.name
       api_service    = aws_ecr_repository.api_service.name
     }
     
     repository = each.value
     policy = jsonencode({
       rules = [{
         rulePriority = 1
         description  = "Keep last 10 images"
         selection = {
           tagStatus     = "any"
           countType     = "imageCountMoreThan"
           countNumber   = 10
         }
         action = {
           type = "expire"
         }
       }]
     })
   }
   ```

2. Create `terraform/modules/ecr/outputs.tf`:
   ```hcl
   output "oura_collector_url" {
     value = aws_ecr_repository.oura_collector.repository_url
   }
   
   output "data_processor_url" {
     value = aws_ecr_repository.data_processor.repository_url
   }
   
   output "api_service_url" {
     value = aws_ecr_repository.api_service.repository_url
   }
   
   output "registry_id" {
     value = aws_ecr_repository.oura_collector.registry_id
   }
   ```

**Validation**:
- Three repositories created
- Lifecycle policy set to keep 10 images

---

### Task 1.1.6: Create Secrets Manager Module
**What**: Secure storage for sensitive values  
**Target**: `terraform/modules/secrets-manager/main.tf`

**Instructions**:

1. Create `terraform/modules/secrets-manager/main.tf`:
   ```hcl
   resource "aws_secretsmanager_secret" "oura_api_key" {
     name                    = "myhealth/oura-api-key"
     recovery_window_in_days = 7
     tags                    = var.tags
   }
   
   resource "aws_secretsmanager_secret_version" "oura_api_key" {
     secret_id = aws_secretsmanager_secret.oura_api_key.id
     secret_string = jsonencode({
       api_key = var.oura_api_key != "" ? var.oura_api_key : "placeholder-to-be-updated"
     })
   }
   
   resource "aws_secretsmanager_secret" "db_credentials" {
     name                    = "myhealth/db-credentials"
     recovery_window_in_days = 7
     tags                    = var.tags
   }
   
   resource "aws_secretsmanager_secret_version" "db_credentials" {
     secret_id = aws_secretsmanager_secret.db_credentials.id
     secret_string = jsonencode({
       username = var.db_username
       password = var.db_password
       host     = var.db_host
       port     = 5432
       dbname   = "myhealth"
     })
   }
   
   resource "aws_secretsmanager_secret" "jwt_secret" {
     name                    = "myhealth/jwt-secret"
     recovery_window_in_days = 7
     tags                    = var.tags
   }
   
   resource "aws_secretsmanager_secret_version" "jwt_secret" {
     secret_id = aws_secretsmanager_secret.jwt_secret.id
     secret_string = jsonencode({
       signing_key = var.jwt_secret_key
     })
   }
   ```

2. Create `terraform/modules/secrets-manager/variables.tf`:
   ```hcl
   variable "oura_api_key" {
     type      = string
     default   = ""
     sensitive = true
   }
   
   variable "db_username" {
     type      = string
     sensitive = true
   }
   
   variable "db_password" {
     type      = string
     sensitive = true
   }
   
   variable "db_host" {
     type = string
   }
   
   variable "jwt_secret_key" {
     type      = string
     sensitive = true
   }
   
   variable "tags" {
     type = map(string)
   }
   ```

3. Create `terraform/modules/secrets-manager/outputs.tf`:
   ```hcl
   output "oura_api_key_arn" {
     value = aws_secretsmanager_secret.oura_api_key.arn
   }
   
   output "db_credentials_arn" {
     value = aws_secretsmanager_secret.db_credentials.arn
   }
   
   output "jwt_secret_arn" {
     value = aws_secretsmanager_secret.jwt_secret.arn
   }
   ```

**Validation**:
- Three secrets created
- Can be updated later with actual values

---

### Task 1.1.7: Create API Gateway Module
**What**: Public API entry point  
**Target**: `terraform/modules/api-gateway/main.tf`

**Instructions**:

1. Create `terraform/modules/api-gateway/main.tf`:
   ```hcl
   resource "aws_apigatewayv2_api" "myhealth" {
     name          = "myhealth-api"
     protocol_type = "HTTP"
     
     cors_configuration {
       allow_origins = ["https://eric-n.com", "https://*.eric-n.com"]
       allow_methods = ["GET", "POST", "OPTIONS"]
       allow_headers = ["Content-Type", "Authorization"]
     }
     
     tags = var.tags
   }
   
   resource "aws_apigatewayv2_stage" "dev" {
     api_id      = aws_apigatewayv2_api.myhealth.id
     name        = "dev"
     auto_deploy = true
     
     access_log_settings {
       destination_arn = aws_cloudwatch_log_group.api_gateway.arn
       format = jsonencode({
         requestId      = "$context.requestId"
         ip             = "$context.identity.sourceIp"
         requestTime    = "$context.requestTime"
         httpMethod     = "$context.httpMethod"
         resourcePath   = "$context.resourcePath"
         status         = "$context.status"
         protocol       = "$context.protocol"
         responseLength = "$context.responseLength"
       })
     }
     
     tags = var.tags
   }
   
   resource "aws_cloudwatch_log_group" "api_gateway" {
     name              = "/aws/apigateway/myhealth"
     retention_in_days = 7
     tags              = var.tags
   }
   
   resource "aws_apigatewayv2_integration" "example" {
     api_id             = aws_apigatewayv2_api.myhealth.id
     integration_type   = "HTTP_PROXY"
     integration_method = "ANY"
     integration_uri    = var.backend_url
     
     payload_format_version = "1.0"
   }
   
   resource "aws_apigatewayv2_route" "example" {
     api_id    = aws_apigatewayv2_api.myhealth.id
     route_key = "ANY /{proxy+}"
     target    = "integrations/${aws_apigatewayv2_integration.example.id}"
   }
   ```

2. Create `terraform/modules/api-gateway/variables.tf`:
   ```hcl
   variable "backend_url" {
     type = string
     description = "Backend service URL (Istio ingress)"
   }
   
   variable "tags" {
     type = map(string)
   }
   ```

3. Create `terraform/modules/api-gateway/outputs.tf`:
   ```hcl
   output "api_endpoint" {
     value = aws_apigatewayv2_stage.dev.invoke_url
   }
   
   output "api_id" {
     value = aws_apigatewayv2_api.myhealth.id
   }
   ```

**Validation**:
- API Gateway created
- CORS configured for eric-n.com

---

### Task 1.1.8: Create Root Terraform Configuration
**What**: Main Terraform files that orchestrate all modules  
**Target**: Root `terraform/` directory

**Instructions**:

1. Create `terraform/versions.tf`:
   ```hcl
   terraform {
     required_version = ">= 1.6"
     required_providers {
       aws = {
         source  = "hashicorp/aws"
         version = "~> 5.30"
       }
     }
     
     backend "s3" {
       bucket         = "myhealth-terraform-state"
       key            = "dev/terraform.tfstate"
       region         = "us-east-1"
       encrypt        = true
       dynamodb_table = "terraform-locks"
     }
   }
   ```

2. Create `terraform/main.tf`:
   ```hcl
   module "networking" {
     source = "./modules/networking"
     
     cluster_name         = var.cluster_name
     vpc_cidr             = var.vpc_cidr
     tags                 = var.tags
     log_retention_in_days = var.log_retention_in_days
   }
   
   module "eks" {
     source = "./modules/eks"
     
     cluster_name              = var.cluster_name
     cluster_version           = var.cluster_version
     vpc_id                    = module.networking.vpc_id
     private_subnets          = module.networking.private_subnets
     node_security_group_id   = module.networking.node_security_group_id
     log_retention_in_days    = var.log_retention_in_days
     tags                     = var.tags
   }
   
   module "rds" {
     source = "./modules/rds"
     
     cluster_name                = var.cluster_name
     vpc_id                      = module.networking.vpc_id
     private_subnets            = module.networking.private_subnets
     eks_node_security_group_id = module.networking.node_security_group_id
     instance_class             = var.rds_instance_class
     db_username                = var.db_username
     db_password                = var.db_password
     multi_az                   = false
     backup_retention_days      = 7
     skip_final_snapshot        = true
     tags                       = var.tags
   }
   
   module "ecr" {
     source = "./modules/ecr"
     tags   = var.tags
   }
   
   module "secrets_manager" {
     source = "./modules/secrets-manager"
     
     db_username       = var.db_username
     db_password       = var.db_password
     db_host           = split(":", module.rds.db_endpoint)[0]
     jwt_secret_key    = var.jwt_secret_key
     oura_api_key      = var.oura_api_key
     tags              = var.tags
   }
   ```

3. Create `terraform/variables.tf`:
   ```hcl
   variable "region" {
     type    = string
     default = "us-east-1"
   }
   
   variable "cluster_name" {
     type    = string
     default = "myhealth"
   }
   
   variable "cluster_version" {
     type    = string
     default = "1.28"
   }
   
   variable "vpc_cidr" {
     type    = string
     default = "10.0.0.0/16"
   }
   
   variable "environment" {
     type    = string
     default = "dev"
   }
   
   variable "rds_instance_class" {
     type    = string
     default = "db.t3.micro"
   }
   
   variable "db_username" {
     type      = string
     sensitive = true
   }
   
   variable "db_password" {
     type      = string
     sensitive = true
   }
   
   variable "jwt_secret_key" {
     type      = string
     sensitive = true
   }
   
   variable "oura_api_key" {
     type      = string
     default   = ""
     sensitive = true
   }
   
   variable "log_retention_in_days" {
     type    = number
     default = 7
   }
   
   variable "tags" {
     type = map(string)
     default = {
       Environment = "dev"
       Project     = "myhealth"
       ManagedBy   = "terraform"
     }
   }
   ```

4. Create `terraform/outputs.tf`:
   ```hcl
   output "eks_cluster_endpoint" {
     value = module.eks.cluster_endpoint
   }
   
   output "eks_cluster_name" {
     value = module.eks.cluster_name
   }
   
   output "eks_oidc_provider_arn" {
     value = module.eks.oidc_provider_arn
   }
   
   output "rds_endpoint" {
     value = module.rds.db_endpoint
   }
   
   output "ecr_oura_collector_url" {
     value = module.ecr.oura_collector_url
   }
   
   output "ecr_data_processor_url" {
     value = module.ecr.data_processor_url
   }
   
   output "ecr_api_service_url" {
     value = module.ecr.api_service_url
   }
   ```

**Validation**:
- Root Terraform files created
- All modules referenced in main.tf

---

## Sprint 1.2: Terraform Initialization & Validation

### Task 1.2.1: Setup S3 Backend State
**What**: Create S3 bucket for Terraform state storage  
**Action Items**:

1. Manually create S3 bucket:
   ```bash
   aws s3api create-bucket \
     --bucket myhealth-terraform-state \
     --region us-east-1 \
     --acl private
   ```

2. Enable versioning:
   ```bash
   aws s3api put-bucket-versioning \
     --bucket myhealth-terraform-state \
     --versioning-configuration Status=Enabled
   ```

3. Enable encryption:
   ```bash
   aws s3api put-bucket-encryption \
     --bucket myhealth-terraform-state \
     --server-side-encryption-configuration '{
       "Rules": [{
         "ApplyServerSideEncryptionByDefault": {
           "SSEAlgorithm": "AES256"
         }
       }]
     }'
   ```

4. Create DynamoDB table for state locking:
   ```bash
   aws dynamodb create-table \
     --table-name terraform-locks \
     --attribute-definitions AttributeName=LockID,AttributeType=S \
     --key-schema AttributeName=LockID,KeyType=HASH \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
     --region us-east-1
   ```

**Validation**: 
- `aws s3 ls | grep myhealth-terraform-state` returns bucket
- `aws dynamodb list-tables --region us-east-1 | grep terraform-locks` returns table

---

### Task 1.2.2: Initialize Terraform
**What**: Initialize Terraform working directory  
**Where**: Run from `terraform/` directory

**Commands**:
```bash
cd terraform/
terraform init
```

**Expected Output**:
```
Initializing the backend...
...
Terraform has been successfully initialized!
```

**Validation**:
- `.terraform/` directory created
- `.terraform.lock.hcl` file exists

---

### Task 1.2.3: Validate Terraform Configuration
**What**: Check for syntax errors and consistency  
**Commands**:
```bash
cd terraform/
terraform fmt -recursive .
terraform validate
```

**Expected Output**:
```
Success! The configuration is valid.
```

**Troubleshooting**:
- If validation fails, check error messages and fix variable types/references
- Ensure all module outputs are properly referenced

---

### Task 1.2.4: Generate Terraform Plan
**What**: Preview what will be created  
**Commands**:
```bash
cd terraform/

# Create terraform.tfvars with sensitive values
cat > terraform.tfvars << EOF
db_username = "myhealth_user"
db_password = "$(openssl rand -base64 32)"
jwt_secret_key = "$(openssl rand -base64 32)"
cluster_name = "myhealth"
environment = "dev"
EOF

# Plan infrastructure
terraform plan -out=tfplan
```

**Expected Output**:
- Plan shows ~50-60 resources to be created (VPC, subnets, EKS, RDS, ECR, etc.)
- No errors or warnings

**Review**:
- Look for correct resource types
- Verify resource naming convention
- Check that databases/registries are properly configured

---

## Sprint 1.3: AWS Credentials & Provider Setup

### Task 1.3.1: Configure AWS Credentials
**What**: Ensure Terraform can authenticate to AWS  
**Instructions**:

Option A - Using AWS CLI credentials:
```bash
aws configure

# Enter:
# AWS Access Key ID: [your-key]
# AWS Secret Access Key: [your-secret]
# Default region name: us-east-1
# Default output format: json
```

Option B - Using environment variables:
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-east-1"
```

**Validation**:
```bash
aws sts get-caller-identity
```

Expected output shows your AWS account ID

---

## Sprint 1.4: Terraform Apply

### Task 1.4.1: Apply Infrastructure
**What**: Create all AWS resources  
**Commands**:
```bash
cd terraform/

# Apply the plan
terraform apply tfplan
```

**Duration**: 15-20 minutes (EKS cluster takes time)

**Expected Output**:
```
Apply complete! Resources: 57 added, 0 changed, 0 destroyed.

Outputs:
eks_cluster_endpoint = ...
eks_cluster_name = ...
rds_endpoint = ...
```

**Important**: Save this output - you'll need the endpoints

**Validation**:
```bash
# Check EKS cluster
aws eks describe-cluster --name myhealth --region us-east-1

# Check RDS
aws rds describe-db-instances --db-instance-identifier myhealth-db --region us-east-1

# Check ECR
aws ecr describe-repositories --region us-east-1
```

---

## Sprint 1.5: Kubernetes Configuration

### Task 1.5.1: Configure kubectl Access
**What**: Setup local kubectl to access EKS cluster  
**Commands**:
```bash
# Update kubeconfig
aws eks update-kubeconfig \
  --name myhealth \
  --region us-east-1 \
  --alias myhealth-dev

# Verify access
kubectl --context myhealth-dev get nodes
```

**Expected Output**:
```
NAME                           STATUS   ROLES    AGE   VERSION
ip-10-0-xxx-xxx.ec2.internal  Ready    <none>   2m    v1.28.x
ip-10-0-xxx-xxx.ec2.internal  Ready    <none>   2m    v1.28.x
```

**Validation**: Both nodes are in `Ready` state

---

### Task 1.5.2: Create Namespaces
**What**: Organize Kubernetes resources  
**Commands**:
```bash
kubectl --context myhealth-dev create namespace myhealth
kubectl --context myhealth-dev create namespace monitoring
kubectl --context myhealth-dev create namespace istio-system

# Verify
kubectl --context myhealth-dev get namespaces
```

---

### Task 1.5.3: Install Istio
**What**: Service mesh for traffic management  
**Commands**:
```bash
# Download Istio
curl -L https://istio.io/downloadIstio | sh -
cd istio-1.20.x
export PATH=$PWD/bin:$PATH

# Install Istio
istioctl install --set profile=demo -y

# Verify installation
kubectl --context myhealth-dev -n istio-system get pods
```

**Expected Output**:
```
NAME                    READY   STATUS    RESTARTS   AGE
istiod-xxx              1/1     Running   0          2m
istio-ingressgateway... 1/1     Running   0          2m
```

---

### Task 1.5.4: Install External Secrets Operator
**What**: Sync AWS Secrets Manager to K8s Secrets  
**Commands**:
```bash
helm repo add external-secrets https://charts.external-secrets.io
helm repo update

helm install external-secrets \
  external-secrets/external-secrets \
  -n external-secrets-system \
  --create-namespace \
  --set installCRDs=true \
  --context myhealth-dev

# Verify
kubectl --context myhealth-dev -n external-secrets-system get pods
```

---

## Sprint 1.6: Validation & Documentation

### Task 1.6.1: Final Infrastructure Validation
**What**: Ensure all components working  
**Checklist**:
- [ ] EKS cluster running (2 nodes ready)
- [ ] RDS database accessible from cluster
- [ ] ECR repositories created (3 repos)
- [ ] Secrets Manager has 3 secrets
- [ ] Istio installed and running
- [ ] External Secrets Operator running
- [ ] kubectl can access cluster

**Validation Commands**:
```bash
# Check cluster
kubectl --context myhealth-dev get nodes -o wide
kubectl --context myhealth-dev get namespaces

# Check services in each namespace
kubectl --context myhealth-dev -n istio-system get pods
kubectl --context myhealth-dev -n external-secrets-system get pods

# Test DB connection from pod (one-time)
kubectl --context myhealth-dev run -it --rm debug --image=postgres:14 --restart=Never -- \
  psql -h <RDS_ENDPOINT> -U myhealth_user -d myhealth -c "\dt"
```

---

### Task 1.6.2: Document Infrastructure
**What**: Create operational documentation  
**File**: `docs/infrastructure-setup.md`

**Contents**:
```markdown
# Infrastructure Setup Documentation

## AWS Resources Created

### EKS Cluster
- Name: myhealth
- Version: 1.28
- Nodes: 2x t3.medium (Spot)
- Endpoint: [from terraform output]

### RDS PostgreSQL
- Endpoint: [from terraform output]
- Username: myhealth_user
- Database: myhealth

### ECR Repositories
1. myhealth/oura-collector
2. myhealth/data-processor
3. myhealth/api-service

### Secrets Manager
1. myhealth/oura-api-key
2. myhealth/db-credentials
3. myhealth/jwt-secret

## Kubernetes Add-ons
- Istio Service Mesh v1.20
- External Secrets Operator v0.9

## Access Commands
kubectl context: myhealth-dev

## Troubleshooting
[Common issues and solutions]
```

---

## Phase 1 Completion Criteria

✅ All tasks completed when:
1. Terraform files organized into modules
2. Infrastructure provisioned successfully (no errors on `terraform apply`)
3. EKS cluster running with 2 nodes
4. RDS PostgreSQL accessible
5. ECR repositories created
6. Secrets Manager populated
7. Istio installed on cluster
8. External Secrets Operator running
9. kubectl configured and accessing cluster
10. Documentation created

**Estimated Time**: 1 week (including troubleshooting)

---

# PHASE 2: MICROSERVICES DEVELOPMENT
**Duration**: 2 weeks  
**Objective**: Build 3 Go microservices with proper structure, tests, and Docker containers

---

## Overview of Three Services

| Service | Purpose | Type | Trigger |
|---------|---------|------|---------|
| oura-collector | Fetch Oura Ring data | CronJob | Every 5 minutes |
| data-processor | Transform & store data | Deployment | HTTP service (port 8080) |
| api-service | Serve API to users | Deployment | HTTP service (port 8080) |

---

## Sprint 2.1: Go Project Setup & Shared Libraries

### Task 2.1.1: Create Go Module Structure
**What**: Setup root Go modules and shared packages  
**Directory Structure**:

```
services/
├── go.work
├── shared/
│   ├── go.mod
│   ├── logger/
│   ├── database/
│   ├── secrets/
│   └── metrics/
├── oura-collector/
│   ├── go.mod
│   ├── cmd/main.go
│   └── internal/...
├── data-processor/
│   ├── go.mod
│   ├── cmd/main.go
│   └── internal/...
└── api-service/
    ├── go.mod
    ├── cmd/main.go
    └── internal/...
```

**Instructions**:

1. Create `services/go.work`:
   ```
   go 1.21
   
   use (
     ./shared
     ./oura-collector
     ./data-processor
     ./api-service
   )
   ```

2. Create `services/shared/go.mod`:
   ```
   module github.com/asian-code/myapp-kubernetes/services/shared
   
   go 1.21
   
   require (
     github.com/sirupsen/logrus v1.9.3
     github.com/lib/pq v1.10.9
     github.com/prometheus/client_golang v1.18.0
     github.com/aws/aws-sdk-go-v2 v1.24.0
     github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.15.7
   )
   ```

3. Create directories:
   ```bash
   mkdir -p services/shared/{logger,database,secrets,metrics}
   mkdir -p services/{oura-collector,data-processor,api-service}/{cmd,internal}
   ```

---

### Task 2.1.2: Create Shared Logger Package
**What**: Centralized logging for all services  
**File**: `services/shared/logger/logger.go`

```go
package logger

import (
  "context"
  "os"
  
  log "github.com/sirupsen/logrus"
)

func Init(serviceName string) *log.Entry {
  l := log.New()
  
  // JSON format for structured logging
  l.SetFormatter(&log.JSONFormatter{
    TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
  })
  
  l.SetOutput(os.Stdout)
  l.SetLevel(log.InfoLevel)
  
  if os.Getenv("LOG_LEVEL") == "debug" {
    l.SetLevel(log.DebugLevel)
  }
  
  return l.WithField("service", serviceName)
}

func WithContext(ctx context.Context, logger *log.Entry) *log.Entry {
  if requestID := ctx.Value("request-id"); requestID != nil {
    logger = logger.WithField("request_id", requestID)
  }
  return logger
}
```

---

### Task 2.1.3: Create Database Connection Package
**What**: PostgreSQL connection pool management  
**File**: `services/shared/database/connection.go`

```go
package database

import (
  "context"
  "fmt"
  
  "github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
  Host     string
  Port     int
  User     string
  Password string
  Database string
  MaxConns int
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
  dsn := fmt.Sprintf(
    "postgres://%s:%s@%s:%d/%s?sslmode=disable",
    cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
  )
  
  poolCfg, err := pgxpool.ParseConfig(dsn)
  if err != nil {
    return nil, err
  }
  
  poolCfg.MaxConns = int32(cfg.MaxConns)
  
  return pgxpool.NewWithConfig(ctx, poolCfg)
}
```

---

### Task 2.1.4: Create Secrets Manager Package
**What**: AWS Secrets Manager integration  
**File**: `services/shared/secrets/secrets.go`

```go
package secrets

import (
  "context"
  "encoding/json"
  
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretClient struct {
  client *secretsmanager.Client
}

func New(ctx context.Context) (*SecretClient, error) {
  cfg, err := config.LoadDefaultConfig(ctx)
  if err != nil {
    return nil, err
  }
  
  return &SecretClient{
    client: secretsmanager.NewFromConfig(cfg),
  }, nil
}

func (s *SecretClient) GetSecret(ctx context.Context, name string) (map[string]string, error) {
  result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
    SecretId: &name,
  })
  if err != nil {
    return nil, err
  }
  
  var secret map[string]string
  if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
    return nil, err
  }
  
  return secret, nil
}
```

---

### Task 2.1.5: Create Prometheus Metrics Package
**What**: Standardized metrics for all services  
**File**: `services/shared/metrics/metrics.go`

```go
package metrics

import (
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
  HTTPRequestsTotal   prometheus.Counter
  HTTPRequestDuration prometheus.Histogram
  DatabaseQueryDuration prometheus.Histogram
}

func New(serviceName string) *Metrics {
  return &Metrics{
    HTTPRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
      Name: "http_requests_total",
      Help: "Total HTTP requests",
      ConstLabels: map[string]string{
        "service": serviceName,
      },
    }),
    HTTPRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
      Name: "http_request_duration_seconds",
      Help: "HTTP request duration",
      Buckets: []float64{.001, .01, .1, 1, 10},
      ConstLabels: map[string]string{
        "service": serviceName,
      },
    }),
    DatabaseQueryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
      Name: "db_query_duration_seconds",
      Help: "Database query duration",
      Buckets: []float64{.001, .01, .1, 1},
      ConstLabels: map[string]string{
        "service": serviceName,
      },
    }),
  }
}
```

---

## Sprint 2.2: oura-collector Service

### Task 2.2.1: Create oura-collector Structure
**What**: Setup project scaffold  
**Files**:

1. `services/oura-collector/go.mod`:
   ```
   module github.com/asian-code/myapp-kubernetes/services/oura-collector
   
   go 1.21
   
   require (
     github.com/asian-code/myapp-kubernetes/services/shared v0.0.0
     github.com/sirupsen/logrus v1.9.3
     github.com/joho/godotenv v1.5.1
   )
   
   replace github.com/asian-code/myapp-kubernetes/services/shared => ../shared
   ```

2. `services/oura-collector/internal/config/config.go`:
   ```go
   package config
   
   import (
     "os"
   )
   
   type Config struct {
     OuraAPIKey    string
     ProcessorURL  string
     LogLevel      string
   }
   
   func Load() *Config {
     return &Config{
       OuraAPIKey:   os.Getenv("OURA_API_KEY"),
       ProcessorURL: os.Getenv("PROCESSOR_URL"),
       LogLevel:     os.Getenv("LOG_LEVEL"),
     }
   }
   ```

3. `services/oura-collector/internal/client/oura.go`:
   ```go
   package client
   
   import (
     "context"
     "encoding/json"
     "fmt"
     "net/http"
     "time"
   
     log "github.com/sirupsen/logrus"
   )
   
   const OuraBaseURL = "https://api.ouraring.com/v2/usercollection"
   
   type SleepData struct {
     ID    string    `json:"id"`
     Day   string    `json:"day"`
     Score int       `json:"score"`
     Duration int    `json:"duration"`
   }
   
   type OuraClient struct {
     apiKey string
     client *http.Client
     logger *log.Entry
   }
   
   func New(apiKey string, logger *log.Entry) *OuraClient {
     return &OuraClient{
       apiKey: apiKey,
       client: &http.Client{Timeout: 10 * time.Second},
       logger: logger,
     }
   }
   
   func (c *OuraClient) GetSleepData(ctx context.Context, date string) (*SleepData, error) {
     url := fmt.Sprintf("%s/daily_sleep?date=%s", OuraBaseURL, date)
     
     req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
     if err != nil {
       return nil, err
     }
     
     req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
     
     resp, err := c.client.Do(req)
     if err != nil {
       return nil, err
     }
     defer resp.Body.Close()
     
     if resp.StatusCode != http.StatusOK {
       return nil, fmt.Errorf("oura API returned %d", resp.StatusCode)
     }
     
     var data SleepData
     if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
       return nil, err
     }
     
     return &data, nil
   }
   ```

4. `services/oura-collector/cmd/main.go`:
   ```go
   package main
   
   import (
     "context"
     "time"
   
     log "github.com/sirupsen/logrus"
     "github.com/asian-code/myapp-kubernetes/services/shared/logger"
     "github.com/asian-code/myapp-kubernetes/services/oura-collector/internal/config"
     "github.com/asian-code/myapp-kubernetes/services/oura-collector/internal/client"
   )
   
   func main() {
     logger := logger.Init("oura-collector")
     cfg := config.Load()
     
     logger.Info("Starting oura-collector")
     
     // Fetch data from Oura
     ouraClient := client.New(cfg.OuraAPIKey, logger)
     
     ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
     defer cancel()
     
     date := time.Now().Format("2006-01-02")
     sleepData, err := ouraClient.GetSleepData(ctx, date)
     if err != nil {
       logger.WithError(err).Error("Failed to fetch sleep data")
       return
     }
     
     logger.WithField("sleep_score", sleepData.Score).Info("Successfully fetched sleep data")
     
     // TODO: Send to data-processor
   }
   ```

---

### Task 2.2.2: Create oura-collector Tests
**What**: Unit tests for Oura client  
**File**: `services/oura-collector/internal/client/oura_test.go`

```go
package client

import (
  "context"
  "testing"
  
  log "github.com/sirupsen/logrus"
)

func TestOuraClientNew(t *testing.T) {
  logger := log.NewEntry(log.New())
  client := New("test-key", logger)
  
  if client.apiKey != "test-key" {
    t.Errorf("Expected apiKey to be 'test-key', got %s", client.apiKey)
  }
}
```

---

### Task 2.2.3: Create oura-collector Dockerfile
**File**: `services/oura-collector/Dockerfile`

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.work ./
COPY shared ./shared
COPY oura-collector ./oura-collector

WORKDIR /app/oura-collector

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/oura-collector/main .

CMD ["./main"]
```

---

## Sprint 2.3: data-processor Service

**Similar structure to oura-collector**

### Task 2.3.1: Create data-processor Service
**Files to create**:
- `services/data-processor/go.mod`
- `services/data-processor/cmd/main.go`
- `services/data-processor/internal/handler/handler.go`
- `services/data-processor/internal/repository/postgres.go`
- `services/data-processor/internal/service/service.go`
- `services/data-processor/Dockerfile`

**Key Responsibilities**:
1. HTTP server on port 8080
2. POST `/api/v1/ingest` - receive data
3. GET `/api/v1/metrics/{type}` - query data
4. Prometheus `/metrics` endpoint
5. Health check `/health`

**Code outline** (abbreviated):
```go
// cmd/main.go
package main

import (
  "context"
  "fmt"
  "net/http"
  
  "github.com/gorilla/mux"
  "github.com/asian-code/myapp-kubernetes/services/shared/logger"
  "github.com/asian-code/myapp-kubernetes/services/shared/database"
  "github.com/asian-code/myapp-kubernetes/services/data-processor/internal/handler"
)

func main() {
  l := logger.Init("data-processor")
  
  // Connect to database
  db, _ := database.NewPool(context.Background(), database.Config{
    Host:     "...",
    Port:     5432,
    User:     "...",
    Password: "...",
    Database: "myhealth",
  })
  defer db.Close()
  
  // Create router
  router := mux.NewRouter()
  h := handler.New(db, l)
  
  router.HandleFunc("/api/v1/ingest", h.Ingest).Methods("POST")
  router.HandleFunc("/api/v1/metrics/{type}", h.GetMetrics).Methods("GET")
  router.HandleFunc("/health", h.Health).Methods("GET")
  router.HandleFunc("/metrics", h.PrometheusMetrics).Methods("GET")
  
  l.Info("Starting data-processor on :8080")
  http.ListenAndServe(":8080", router)
}
```

---

## Sprint 2.4: api-service Service

### Task 2.4.1: Create api-service Service
**Files to create**:
- `services/api-service/go.mod`
- `services/api-service/cmd/main.go`
- `services/api-service/internal/handler/stats.go`
- `services/api-service/internal/auth/auth.go`
- `services/api-service/api/openapi.yaml`
- `services/api-service/Dockerfile`

**Key Responsibilities**:
1. REST API on port 8080
2. JWT authentication
3. GET `/api/v1/dashboard` - aggregated metrics
4. GET `/api/v1/sleep` - sleep metrics
5. GET `/api/v1/activity` - activity metrics
6. POST `/auth/login` - login endpoint
7. Prometheus `/metrics` endpoint

---

## Phase 2 Completion Criteria

✅ All tasks completed when:
1. Shared Go packages created (logger, database, secrets, metrics)
2. oura-collector service built and tested
3. data-processor service built and tested
4. api-service service built and tested
5. All services have >80% test coverage
6. All services have Dockerfiles
7. No compilation errors
8. Services can be built to Docker images

---

# PHASE 3: HELM CHART CREATION
**Duration**: 3 days  
**Objective**: Create single Helm chart to deploy entire stack

[Instructions for Helm chart creation...]

---

# PHASE 4: CI/CD PIPELINE SETUP
**Duration**: 3 days  
**Objective**: Automate builds and deployments

[Instructions for GitHub Actions workflows...]

---

# PHASE 5: MONITORING & OBSERVABILITY
**Duration**: 3 days  
**Objective**: Setup Prometheus + Grafana dashboards

[Instructions for monitoring stack...]

---

# PHASE 6: API GATEWAY & TESTING
**Duration**: 1 week  
**Objective**: Complete API Gateway setup and perform testing

[Instructions for API Gateway and testing...]

---

# General Guidelines

## Code Quality Standards

All code must follow:
- Go: `go fmt`, `go vet`, `golangci-lint`
- >80% test coverage
- Structured error handling
- Prometheus metrics on all services
- JSON logging

## Deployment Checklist

Before moving to next phase:
- [ ] All tests pass
- [ ] No linting errors
- [ ] Docker images build successfully
- [ ] Terraform plan shows expected resources
- [ ] Documentation updated

## Troubleshooting Template

For each phase, document:
1. **Problem**: Clear description
2. **Root Cause**: Why it happened
3. **Solution**: How to fix it
4. **Prevention**: How to avoid in future

## Communication

- Update progress on each task completion
- Flag blockers immediately
- Test each component before moving to next
- Document any deviations from plan

---

**Next Steps**: 
1. Begin Phase 1 - Infrastructure Setup
2. Report completion status after each sprint
3. Escalate any blockers to human supervisor

