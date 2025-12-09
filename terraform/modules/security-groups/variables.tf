variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster (used for resource naming)"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where security groups will be created"
}

variable "vpc_cidr" {
  type        = string
  description = "VPC CIDR block for additional access rules"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all security groups"
  default     = {}
}
