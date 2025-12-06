variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster"
}

variable "vpc_cidr" {
  type        = string
  description = "CIDR block for the VPC"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
}

variable "log_retention_in_days" {
  type        = number
  description = "CloudWatch log retention in days"
  default     = 7
}
