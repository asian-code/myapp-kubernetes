variable "oura_api_key" {
  type        = string
  sensitive   = true
  description = "Oura Ring API Key"
}

variable "db_username" {
  type        = string
  sensitive   = true
  default     = null
  description = "Database username. If not provided, a random username will be generated."
}

variable "db_password" {
  type        = string
  sensitive   = true
  default     = null
  description = "Database password. If not provided, a random password will be generated."
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
