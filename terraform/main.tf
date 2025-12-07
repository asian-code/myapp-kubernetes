locals {
  tags = merge(
    {
      Environment = var.environment
      ManagedBy   = "terraform"
      Project     = var.tags["Project"]
    },
    var.tags
  )
}

# S3 bucket for Terraform state (backend can be configured after first apply)
resource "random_id" "tfstate_suffix" {
  byte_length = 4
}

resource "aws_s3_bucket" "tf_state" {
  bucket = "myhealth-terraform-state-${random_id.tfstate_suffix.hex}"
}

resource "aws_s3_bucket_versioning" "tf_state" {
  bucket = aws_s3_bucket.tf_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "tf_state" {
  bucket = aws_s3_bucket.tf_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "tf_state" {
  bucket = aws_s3_bucket.tf_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Networking module
module "networking" {
  source = "./modules/networking"

  cluster_name          = var.cluster_name
  vpc_cidr              = var.vpc_cidr
  tags                  = local.tags
  log_retention_in_days = var.log_retention_in_days
}

# EKS module
module "eks" {
  source = "./modules/eks"

  cluster_name    = var.cluster_name
  cluster_version = var.cluster_version
  vpc_id          = module.networking.vpc_id
  private_subnets = module.networking.private_subnets
  tags            = local.tags
}

# RDS module
module "rds" {
  source = "./modules/rds"

  cluster_name          = var.cluster_name
  vpc_id                = module.networking.vpc_id
  private_subnets       = module.networking.private_subnets
  allowed_cidr_blocks   = [var.vpc_cidr]
  instance_class        = var.rds_instance_class
  db_username           = "myhealth_user"
  multi_az              = var.environment == "prod" ? true : false
  backup_retention_days = var.environment == "prod" ? 30 : 7
  skip_final_snapshot   = var.environment == "prod" ? false : true
  final_snapshot_identifier = var.environment == "prod" ? "${var.cluster_name}-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}" : null
  deletion_protection   = var.environment == "prod" ? true : false
  tags                  = local.tags
}

# ECR module
module "ecr" {
  source = "./modules/ecr"
  tags   = local.tags
}

# Secrets Manager module
module "secrets_manager" {
  source = "./modules/secrets-manager"

  db_username        = "myhealth_user"
  db_password        = module.rds.db_password
  db_host            = module.rds.db_address
  oura_client_id     = var.oura_client_id
  oura_client_secret = var.oura_client_secret
  tags               = local.tags
}

# API Gateway module (optional until ACM cert + backend provided)
module "api_gateway" {
  count = var.acm_certificate_arn != "" && var.backend_url != "" ? 1 : 0

  source              = "./modules/api-gateway"
  domain_name         = var.api_gateway_domain_name
  create_domain_records = var.api_gateway_create_domain_records
  route53_zone_id       = var.api_gateway_route53_zone_id
  backend_url           = var.backend_url
  acm_certificate_arn   = var.acm_certificate_arn
  tags                  = local.tags
}
