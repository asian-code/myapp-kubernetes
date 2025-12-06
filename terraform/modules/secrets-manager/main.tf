# Generate database master password
resource "random_password" "db_master_password" {
  length           = 40
  special          = true
  override_special = "@#%&*()-_=+[]{}<>:?"
}

# Generate database username
resource "random_string" "db_username" {
  length  = 16
  special = false
}

# Generate JWT secret
resource "random_password" "jwt_secret" {
  length           = 32
  special          = true
  override_special = "@#%&*()-_=+[]{}<>:?"
}

# Oura API Key Secret
resource "aws_secretsmanager_secret" "oura_api_key" {
  name                    = "myhealth/oura-api-key"
  description             = "Oura Ring API Key"
  recovery_window_in_days = 7

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "oura_api_key" {
  secret_id = aws_secretsmanager_secret.oura_api_key.id
  secret_string = jsonencode({
    api_key = var.oura_api_key
  })
}

# Database Master Password Secret
resource "aws_secretsmanager_secret" "db_master_password" {
  name                    = "myhealth/rds/master-password"
  description             = "RDS Master Database Password"
  recovery_window_in_days = 7

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "db_master_password" {
  secret_id     = aws_secretsmanager_secret.db_master_password.id
  secret_string = random_password.db_master_password.result
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
    username = var.db_username != null ? var.db_username : random_string.db_username.result
    password = var.db_password != null ? var.db_password : random_password.db_master_password.result
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
    secret_key = var.jwt_secret_key != null ? var.jwt_secret_key : random_password.jwt_secret.result
  })
}
