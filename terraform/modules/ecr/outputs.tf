output "oura_collector_url" {
  value       = aws_ecr_repository.oura_collector.repository_url
  description = "URL of the oura-collector ECR repository"
}

output "data_processor_url" {
  value       = aws_ecr_repository.data_processor.repository_url
  description = "URL of the data-processor ECR repository"
}

output "api_service_url" {
  value       = aws_ecr_repository.api_service.repository_url
  description = "URL of the api-service ECR repository"
}

output "registry_id" {
  value       = aws_ecr_repository.oura_collector.registry_id
  description = "The account ID of the registry"
}
