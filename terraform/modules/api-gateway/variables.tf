variable "backend_url" {
  type        = string
  description = "Backend service URL (Istio ingress)"
}

variable "acm_certificate_arn" {
  type        = string
  description = "ARN of the ACM certificate for myhealth.eric-n.com"
}

variable "domain_name" {
  type        = string
  default     = ""
  description = "Custom domain name for the API Gateway"
}

variable "create_domain_records" {
  type        = bool
  default     = false
  description = "Whether to create Route53 records for the custom domain"
}

variable "route53_zone_id" {
  type        = string
  default     = ""
  description = "Route53 hosted zone ID for creating the API Gateway custom domain record"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
}
