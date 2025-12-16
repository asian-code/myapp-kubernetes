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

module "s3_state" {
  source = "./modules/s3-state"

  bucket_name = "myhealth-terraform-state"
  tags        = local.tags
}

# Networking module
module "networking" {
  source = "./modules/networking"

  cluster_name          = var.cluster_name
  vpc_cidr              = var.vpc_cidr
  tags                  = local.tags
  log_retention_in_days = var.log_retention_in_days
}

# Security Groups module (centralized)
module "security_groups" {
  source = "./modules/security-groups"

  cluster_name = var.cluster_name
  vpc_id       = module.networking.vpc_id
  vpc_cidr     = var.vpc_cidr
  tags         = local.tags
}

# Reuse centrally managed security groups across all tiers
locals {
  security_group_ids = {
    eks_cluster   = module.security_groups.eks_cluster_security_group_id
    rds           = module.security_groups.rds_security_group_id
    vpc_endpoints = module.security_groups.vpc_endpoints_security_group_id
  }
}

# VPC Endpoints module
module "vpc_endpoints" {
  source = "./modules/vpc-endpoints"

  cluster_name            = var.cluster_name
  vpc_id                  = module.networking.vpc_id
  vpc_cidr                = var.vpc_cidr
  private_subnets         = module.networking.private_subnets
  private_route_table_ids = module.networking.private_route_table_ids
  region                  = var.region
  security_group_id       = local.security_group_ids.vpc_endpoints
  tags                    = local.tags
}

# EKS module
module "eks" {
  source = "./modules/eks"

  cluster_name                         = var.cluster_name
  cluster_version                      = var.cluster_version
  vpc_id                               = module.networking.vpc_id
  private_subnets                      = module.networking.private_subnets
  cluster_security_group_id            = local.security_group_ids.eks_cluster
  tags                                 = local.tags
  cluster_endpoint_public_access_cidrs = var.cluster_endpoint_public_access_cidrs
}

# RDS module
module "rds" {
  source = "./modules/rds"

  cluster_name              = var.cluster_name
  vpc_id                    = module.networking.vpc_id
  db_subnets                = module.networking.database_subnets
  security_group_id         = local.security_group_ids.rds
  instance_class            = var.rds_instance_class
  db_username               = "myhealth_user"
  multi_az                  = var.environment == "prod" ? true : false
  backup_retention_days     = var.environment == "prod" ? 30 : 7
  skip_final_snapshot       = var.environment == "prod" ? false : true
  final_snapshot_identifier = var.environment == "prod" ? "${var.cluster_name}-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}" : null
  deletion_protection       = var.environment == "prod" ? true : false
  tags                      = local.tags
}

# ECR module
module "ecr" {
  source = "./modules/ecr"
  tags   = local.tags
}

# ACM Certificate (Manual Validation via Cloudflare)
module "acm" {
  source = "./modules/acm"

  domain_name = "myhealth.eric-n.com"
  tags        = local.tags
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
