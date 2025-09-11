# Variables for EKS Three-Tier Application

variable "region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-west-2"
}

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
  default     = "en-mspy"
}

variable "cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.28"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "node_groups" {
  description = "Configuration for EKS managed node groups"
  type = map(object({
    instance_types = list(string)
    min_size       = number
    max_size       = number
    desired_size   = number
    capacity_type  = optional(string, "ON_DEMAND")
    ami_type       = optional(string, "AL2_x86_64")
    disk_size      = optional(number, 50)
    labels         = optional(map(string), {})
    taints = optional(list(object({
      key    = string
      value  = string
      effect = string
    })), [])
  }))
  default = {
    general = {
      instance_types = ["t3.medium", "t3a.medium"]
      min_size       = 2
      max_size       = 10
      desired_size   = 3
      capacity_type  = "ON_DEMAND"
    }
    spot = {
      instance_types = ["t3.medium", "t3a.medium", "t3.large", "t3a.large"]
      min_size       = 0
      max_size       = 20
      desired_size   = 2
      capacity_type  = "SPOT"
      labels = {
        "node-type" = "spot"
      }
      taints = [
        {
          key    = "spot-instance"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
    }
  }
}

variable "enable_monitoring" {
  description = "Enable monitoring stack (Prometheus, Grafana)"
  type        = bool
  default     = true
}

variable "enable_logging" {
  description = "Enable EKS control plane logging"
  type        = bool
  default     = true
}

variable "log_retention_in_days" {
  description = "CloudWatch log retention period"
  type        = number
  default     = 30
}

variable "enable_irsa" {
  description = "Enable IAM Roles for Service Accounts"
  type        = bool
  default     = true
}

variable "cluster_endpoint_private_access" {
  description = "Enable private API server endpoint"
  type        = bool
  default     = true
}

variable "cluster_endpoint_public_access" {
  description = "Enable public API server endpoint"
  type        = bool
  default     = true
}

variable "cluster_endpoint_public_access_cidrs" {
  description = "List of CIDR blocks that can access the public API server endpoint"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "enable_cluster_encryption" {
  description = "Enable envelope encryption of Kubernetes secrets"
  type        = bool
  default     = true
}

variable "kms_key_deletion_window_in_days" {
  description = "KMS Key deletion window"
  type        = number
  default     = 7
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default = {
    Environment = "production"
    Project     = "three-tier-app"
    Owner       = "platform-team"
    ManagedBy   = "terraform"
    Application = "kubernetes"
  }
}

# Database variables (for MongoDB running in EKS)
variable "enable_mongodb" {
  description = "Enable MongoDB StatefulSet in EKS cluster"
  type        = bool
  default     = true
}

variable "mongodb_config" {
  description = "MongoDB StatefulSet configuration"
  type = object({
    replica_set_name = string
    port             = number
    storage_size     = string
    replicas         = number
    version          = string
    storage_class    = string
  })
  default = {
    replica_set_name = "rs0"
    port             = 27017
    storage_size     = "20Gi"
    replicas         = 3 # One per AZ for HA
    version          = "7.0"
    storage_class    = "gp3"
  }
}

# Application configuration
variable "app_config" {
  description = "Application configuration"
  type = object({
    frontend = object({
      replicas = number
      image    = string
      tag      = string
      port     = number
    })
    backend = object({
      replicas = number
      image    = string
      tag      = string
      port     = number
    })
    database = object({
      replicas     = number
      image        = string
      tag          = string
      port         = number
      storage_size = string
    })
  })
  default = {
    frontend = {
      replicas = 3
      image    = "nginx"
      tag      = "latest"
      port     = 3000
    }
    backend = {
      replicas = 3
      image    = "node"
      tag      = "18-alpine"
      port     = 5000
    }
    database = {
      replicas     = 3 # MongoDB replica set across 3 AZs
      image        = "mongo"
      tag          = "7.0"
      port         = 27017
      storage_size = "20Gi"
    }
  }
}

# Monitoring configuration
variable "monitoring_config" {
  description = "Comprehensive observability stack configuration"
  type = object({
    prometheus = object({
      retention_days  = number
      storage_size    = string
      scrape_interval = string
    })
    grafana = object({
      admin_password = string
      storage_size   = string
    })
    alertmanager = object({
      storage_size = string
    })
    enable_fluent_bit  = bool
    log_retention_days = number
  })
  default = {
    prometheus = {
      retention_days  = 30
      storage_size    = "50Gi"
      scrape_interval = "30s"
    }
    grafana = {
      admin_password = "admin" # Change this in production!
      storage_size   = "10Gi"
    }
    alertmanager = {
      storage_size = "10Gi"
    }
    enable_fluent_bit  = true
    log_retention_days = 30
  }
  sensitive = true
}

# Security configuration
variable "security_config" {
  description = "Security and network policy configuration"
  type = object({
    enable_pod_security_standards = bool
    enable_network_policies       = bool
    enable_external_secrets       = bool
    secrets_manager_region        = string
    enable_zero_trust_policies    = bool
    enable_calico_cni             = bool
  })
  default = {
    enable_pod_security_standards = true
    enable_network_policies       = true
    enable_external_secrets       = true
    secrets_manager_region        = "us-west-2"
    enable_zero_trust_policies    = true # Frontend→API, API→MongoDB only
    enable_calico_cni             = true
  }
}
