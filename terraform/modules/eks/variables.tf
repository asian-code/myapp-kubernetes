variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster"
}

variable "cluster_version" {
  type        = string
  description = "Kubernetes version to use for the EKS cluster"
  default     = "1.28"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where the cluster will be created"
}

variable "private_subnets" {
  type        = list(string)
  description = "List of private subnet IDs for the cluster"
}

variable "node_security_group_id" {
  type        = string
  description = "Security group ID for EKS nodes"
}

variable "node_instance_types" {
  type        = list(string)
  description = "Instance types for the EKS node group"
  default     = ["t3.medium"]
}

variable "node_min_size" {
  type        = number
  description = "Minimum number of nodes"
  default     = 2
}

variable "node_max_size" {
  type        = number
  description = "Maximum number of nodes"
  default     = 4
}

variable "node_desired_size" {
  type        = number
  description = "Desired number of nodes"
  default     = 2
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
}
