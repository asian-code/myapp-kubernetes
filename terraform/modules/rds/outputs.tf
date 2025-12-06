output "db_endpoint" {
  value       = module.db.db_instance_endpoint
  description = "RDS database endpoint"
}

output "db_address" {
  value       = module.db.db_instance_address
  description = "RDS database address"
}

output "db_name" {
  value       = module.db.db_instance_name
  description = "Name of the database"
}

output "db_username" {
  value       = module.db.db_instance_username
  sensitive   = true
  description = "Database username"
}

output "db_port" {
  value       = module.db.db_instance_port
  description = "Database port"
}
