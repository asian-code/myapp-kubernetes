variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "myhealth"

  validation {
    condition     = can(regex("^[a-z0-9-]{1,64}$", var.cluster_name))
    error_message = "Cluster name must be lowercase alphanumeric with hyphens (max 64 chars)."
  }
}

variable "cluster_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.33"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.10.0.0/16"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "log_retention_in_days" {
  description = "CloudWatch log retention"
  type        = number
  default     = 7
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.small"

  validation {
    condition     = can(regex("^db\\.", var.rds_instance_class))
    error_message = "RDS instance class must start with 'db.'."
  }
}

variable "tags" {
  description = "Additional tags"
  type        = map(string)
  default = {
    Project = "myhealth"
  }
}

variable "oura_client_id" {
  description = "Oura OAuth Client ID"
  type        = string
  sensitive   = true
}

variable "oura_client_secret" {
  description = "Oura OAuth Client Secret"
  type        = string
  sensitive   = true
}
variable "cluster_endpoint_public_access_cidrs" {
  type        = list(string)
  description = "List of CIDR blocks allowed to access the cluster endpoint publicly"
}