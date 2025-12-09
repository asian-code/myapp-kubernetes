output "s3_endpoint_id" {
  value       = aws_vpc_endpoint.s3.id
  description = "ID of the S3 VPC endpoint"
}

output "ecr_api_endpoint_id" {
  value       = aws_vpc_endpoint.ecr_api.id
  description = "ID of the ECR API VPC endpoint"
}

output "ecr_dkr_endpoint_id" {
  value       = aws_vpc_endpoint.ecr_dkr.id
  description = "ID of the ECR DKR VPC endpoint"
}

output "secretsmanager_endpoint_id" {
  value       = aws_vpc_endpoint.secretsmanager.id
  description = "ID of the Secrets Manager VPC endpoint"
}

output "logs_endpoint_id" {
  value       = aws_vpc_endpoint.logs.id
  description = "ID of the CloudWatch Logs VPC endpoint"
}

output "vpc_endpoints_security_group_id" {
  value       = var.security_group_id
  description = "Security group ID provided to VPC endpoints"
}
