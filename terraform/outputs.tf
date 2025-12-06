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

output "tf_state_bucket" {
  value = aws_s3_bucket.tf_state.bucket
}

output "api_gateway_endpoint" {
  value       = try(module.api_gateway[0].api_endpoint, null)
  description = "Invoke URL for API Gateway dev stage"
}
