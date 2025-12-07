output "api_endpoint" {
  value       = module.api_gateway.stage_invoke_url
  description = "API Gateway invoke URL"
}

output "api_id" {
  value       = module.api_gateway.api_id
  description = "API Gateway API ID"
}

output "domain_name_target" {
  value       = module.api_gateway.domain_name_target_domain_name
  description = "Target domain name for DNS CNAME record"
}

output "route53_record_fqdn" {
  value       = one(aws_route53_record.api_gateway[*].fqdn)
  description = "FQDN of the Route53 record created for the API Gateway custom domain"
  sensitive   = false
}
