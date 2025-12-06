output "cluster_endpoint" {
  value       = module.eks.cluster_endpoint
  description = "Endpoint for your EKS Kubernetes API"
}

output "cluster_name" {
  value       = module.eks.cluster_name
  description = "The name of the EKS cluster"
}

output "cluster_version" {
  value       = module.eks.cluster_version
  description = "The Kubernetes server version for the cluster"
}

output "cluster_certificate_authority_data" {
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
  description = "Base64 encoded certificate data required to communicate with the cluster"
}

output "oidc_provider_arn" {
  value       = module.eks.oidc_provider_arn
  description = "ARN of the OIDC Provider for IAM Roles for Service Accounts"
}

output "oidc_provider_url" {
  value       = module.eks.oidc_provider
  description = "URL of the OIDC Provider"
}

output "cluster_oidc_issuer_url" {
  value       = module.eks.cluster_oidc_issuer_url
  description = "Issuer URL for the EKS cluster OIDC provider"
}
