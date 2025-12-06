output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "VPC ID"
}

output "private_subnets" {
  value       = module.vpc.private_subnets
  description = "List of private subnet IDs"
}

output "public_subnets" {
  value       = module.vpc.public_subnets
  description = "List of public subnet IDs"
}

output "node_security_group_id" {
  value       = aws_security_group.additional_node_sg.id
  description = "Security group ID for EKS nodes"
}
