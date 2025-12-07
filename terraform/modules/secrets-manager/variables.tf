variable "oura_client_id" {
  type        = string
  sensitive   = true
  description = "Oura OAuth Client ID"
  # No default - must be provided by caller module
}

variable "oura_client_secret" {
  type        = string
  sensitive   = true
  description = "Oura OAuth Client Secret"
  # No default - must be provided by caller module
}

variable "db_username" {
  type        = string
  sensitive   = true
  description = "Database username. Must match the RDS configuration."

  validation {
    condition     = var.db_username != null && length(var.db_username) > 0
    error_message = "db_username is required and cannot be empty."
  }
}

variable "db_password" {
  type        = string
  sensitive   = true
  description = "Database password from the RDS module."

  validation {
    condition     = var.db_password != null && length(var.db_password) > 0
    error_message = "db_password is required and cannot be empty."
  }
}

variable "db_host" {
  type        = string
  description = "Database host address"
}

variable "jwt_secret_key" {
  type        = string
  sensitive   = true
  default     = null
  description = "JWT signing secret key. If not provided, a random secret will be generated."
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
}
