output "vpc_id" {
  value = module.networking.vpc_id
}

output "private_subnets" {
  value = module.networking.private_subnets
}

output "eks_cluster_name" {
  value = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "rds_endpoint" {
  value = module.rds.db_endpoint
}

output "ecr_registry_id" {
  value = module.ecr.registry_id
}

output "acm_certificate_arn" {
  description = "The ARN of the SSL certificate. Use this in your Istio Gateway Service annotation."
  value       = module.acm.certificate_arn
}

output "acm_validation_records" {
  description = "Add these CNAME records to Cloudflare to validate the certificate."
  value       = module.acm.validation_records
}
