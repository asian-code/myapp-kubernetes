variable "aws_profile" {
  description = "AWS CLI profile to use"
  type        = string
  default     = "me"
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "myhealth"
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
}

variable "log_retention_in_days" {
  description = "CloudWatch log retention"
  type        = number
  default     = 7
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "node_instance_types" {
  description = "EKS node instance types"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "node_min_size" {
  description = "EKS node group min size"
  type        = number
  default     = 2
}

variable "node_max_size" {
  description = "EKS node group max size"
  type        = number
  default     = 4
}

variable "node_desired_size" {
  description = "EKS node group desired size"
  type        = number
  default     = 2
}

variable "acm_certificate_arn" {
  description = "ACM certificate ARN for api.myhealth.eric-n.com"
  type        = string
  default     = ""
}

variable "backend_url" {
  description = "Backend service URL (Istio ingress)"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Additional tags"
  type        = map(string)
  default = {
    Project = "myhealth"
  }
}
