# Generate secrets when not provided by the caller
resource "random_password" "jwt_secret" {
  length           = 64
  special          = true
  override_special = "@#%&*()-_=+[]{}<>:?"
}

locals {
  jwt_secret_value = coalesce(var.jwt_secret_key, random_password.jwt_secret.result)
}

# Oura OAuth Credentials Secret
resource "aws_secretsmanager_secret" "oura_credentials" {
  name                    = "myhealth/oura-credentials"
  description             = "Oura OAuth Client ID and Secret"
  recovery_window_in_days = 7

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "oura_credentials" {
  secret_id = aws_secretsmanager_secret.oura_credentials.id
  secret_string = jsonencode({
    client_id     = var.oura_client_id
    client_secret = var.oura_client_secret
  })
}

# Database Credentials Secret (generated + optional user-provided)
resource "aws_secretsmanager_secret" "db_credentials" {
  name                    = "myhealth/db-credentials"
  description             = "Database credentials"
  recovery_window_in_days = 7

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = var.db_username
    password = var.db_password
    host     = var.db_host
    port     = 5432
    dbname   = "myhealth"
  })
}

# JWT Secret (generated + optional user-provided)
resource "aws_secretsmanager_secret" "jwt_secret" {
  name                    = "myhealth/jwt-secret"
  description             = "JWT signing secret"
  recovery_window_in_days = 7

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "jwt_secret" {
  secret_id = aws_secretsmanager_secret.jwt_secret.id
  secret_string = jsonencode({
    secret_key = local.jwt_secret_value
  })
}
