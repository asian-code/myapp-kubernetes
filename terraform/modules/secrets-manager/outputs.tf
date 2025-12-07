output "oura_credentials_arn" {
  value       = aws_secretsmanager_secret.oura_credentials.arn
  description = "ARN of the Oura credentials secret"
}

output "db_credentials_arn" {
  value       = aws_secretsmanager_secret.db_credentials.arn
  description = "ARN of the database credentials secret"
}

output "db_credentials_username" {
  value       = var.db_username
  sensitive   = true
  description = "Database username"
}

output "db_credentials_password" {
  value       = var.db_password
  sensitive   = true
  description = "Database password"
}

output "jwt_secret_arn" {
  value       = aws_secretsmanager_secret.jwt_secret.arn
  description = "ARN of the JWT secret"
}

output "jwt_secret_key" {
  value       = local.jwt_secret_value
  sensitive   = true
  description = "JWT secret key (generated if not provided)"
}
