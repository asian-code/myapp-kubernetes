variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster"
}

variable "cluster_version" {
  type        = string
  description = "Kubernetes version to use for the EKS cluster"
  default     = "1.31"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where the cluster will be created"
}

variable "private_subnets" {
  type        = list(string)
  description = "List of private subnet IDs for the cluster"
}

variable "cluster_security_group_id" {
  type        = string
  description = "Security group ID for the EKS cluster"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
}

variable "cluster_endpoint_public_access_cidrs" {
  type        = list(string)
  description = "List of CIDR blocks allowed to access the cluster endpoint publicly (only me)"
}