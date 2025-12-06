variable "backend_url" {
  type        = string
  description = "Backend service URL (Istio ingress)"
}

variable "acm_certificate_arn" {
  type        = string
  description = "ARN of the ACM certificate for myhealth.eric-n.com"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
}
