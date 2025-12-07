module "db" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 6.13"

  identifier = "${var.cluster_name}-db"

  engine               = "postgres"
  engine_version       = "15.4"
  family               = "postgres15"
  major_engine_version = "15"
  instance_class       = var.instance_class

  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = "myhealth"
  username = var.db_username
  password = coalesce(var.db_password, random_password.db_password.result)
  port     = 5432

  multi_az               = var.multi_az
  db_subnet_group_name   = aws_db_subnet_group.myhealth.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  maintenance_window      = "sun:04:00-sun:05:00"
  backup_window           = "03:00-04:00"
  backup_retention_period = var.backup_retention_days

  skip_final_snapshot       = var.skip_final_snapshot

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  create_cloudwatch_log_group     = true

  deletion_protection   = var.deletion_protection
  copy_tags_to_snapshot = true

  tags = var.tags
}

# Generate a database password if one is not provided
resource "random_password" "db_password" {
  length  = 32
  special = true
}

# DB Subnet Group (still created manually for flexibility)
resource "aws_db_subnet_group" "myhealth" {
  name       = "${var.cluster_name}-db-subnet-group"
  subnet_ids = var.private_subnets

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-db-subnet-group"
  })
}

# Security Group for RDS
resource "aws_security_group" "rds" {
  name        = "${var.cluster_name}-rds-sg"
  description = "Security group for RDS PostgreSQL"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
    description = "PostgreSQL from VPC (EKS Auto Mode pods)"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

