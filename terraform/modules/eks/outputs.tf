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

output "cluster_oidc_issuer_url" {
  value       = module.eks.cluster_oidc_issuer_url
  description = "Issuer URL for the EKS cluster OIDC provider"
}

output "auto_mode_node_role_arn" {
  value       = module.eks.node_iam_role_arn
  description = "ARN of the IAM role for Auto Mode nodes (created by the module)"
}

output "auto_mode_node_role_name" {
  value       = module.eks.node_iam_role_name
  description = "Name of the IAM role for Auto Mode nodes (created by the module)"
}
