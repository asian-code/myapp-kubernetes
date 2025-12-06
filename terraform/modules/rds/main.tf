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
  password = var.db_password
  port     = 5432

  multi_az               = var.multi_az
  db_subnet_group_name   = aws_db_subnet_group.myhealth.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  maintenance_window      = "sun:04:00-sun:05:00"
  backup_window           = "03:00-04:00"
  backup_retention_period = var.backup_retention_days

  skip_final_snapshot = var.skip_final_snapshot

  enabled_cloudwatch_logs_exports = ["postgresql"]
  create_cloudwatch_log_group     = true

  deletion_protection   = false
  copy_tags_to_snapshot = true

  tags = var.tags
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
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.eks_node_security_group_id]
    description     = "PostgreSQL from EKS nodes"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

