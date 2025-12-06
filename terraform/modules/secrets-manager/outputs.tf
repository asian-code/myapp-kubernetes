output "oura_api_key_arn" {
  value       = aws_secretsmanager_secret.oura_api_key.arn
  description = "ARN of the Oura API key secret"
}

output "db_master_password_arn" {
  value       = aws_secretsmanager_secret.db_master_password.arn
  description = "ARN of the generated RDS master password secret"
}

output "db_credentials_arn" {
  value       = aws_secretsmanager_secret.db_credentials.arn
  description = "ARN of the database credentials secret"
}

output "db_credentials_username" {
  value       = var.db_username != null ? var.db_username : random_string.db_username.result
  sensitive   = true
  description = "Database username (generated if not provided)"
}

output "db_credentials_password" {
  value       = var.db_password != null ? var.db_password : random_password.db_master_password.result
  sensitive   = true
  description = "Database password (generated if not provided)"
}

output "jwt_secret_arn" {
  value       = aws_secretsmanager_secret.jwt_secret.arn
  description = "ARN of the JWT secret"
}

output "jwt_secret_key" {
  value       = var.jwt_secret_key != null ? var.jwt_secret_key : random_password.jwt_secret.result
  sensitive   = true
  description = "JWT secret key (generated if not provided)"
}
