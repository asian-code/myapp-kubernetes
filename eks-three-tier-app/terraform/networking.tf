module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 6.0"

  # Simplified 2-tier architecture for EKS with all workloads in private subnets
  name = "${local.cluster_name}-vpc"
  cidr = var.vpc_cidr
  azs  = local.azs

  # 2-Subnet Architecture: Public (ALB + NAT), Private (EKS all workloads)
  public_subnets  = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k)]     # /24 subnets for ALB & NAT
  private_subnets = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 4, k + 1)] # /20 subnets for all EKS workloads

  # Essential networking features
  enable_nat_gateway   = true
  single_nat_gateway   = false # HA: One NAT per AZ for redundancy
  enable_dns_hostnames = true
  enable_dns_support   = true

  # Kubernetes subnet tags for AWS Load Balancer Controller discovery
  public_subnet_tags = {
    "kubernetes.io/role/elb"                      = "1"
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb"             = "1"
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  }

  # VPC Flow Logs for security monitoring
  enable_flow_log                                 = true
  create_flow_log_cloudwatch_iam_role             = true
  create_flow_log_cloudwatch_log_group            = true
  flow_log_cloudwatch_log_group_retention_in_days = var.log_retention_in_days

  tags = var.tags
}
# Additional Security Group for EKS nodes - All workloads
resource "aws_security_group" "additional_node_sg" {
  name        = "${local.cluster_name}-node-sg"
  description = "Additional security group for EKS nodes (all workloads)"
  vpc_id      = module.vpc.vpc_id

  # Allow all traffic between nodes (for pod-to-pod communication)
  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
  }

  # Allow HTTPS traffic from ALB to nodes (for frontend/backend ingress)
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = module.vpc.public_subnets_cidr_blocks
  }

  # Allow HTTP traffic from ALB to nodes (for frontend/backend ingress)
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = module.vpc.public_subnets_cidr_blocks
  }

  # NodePort range for services
  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = module.vpc.public_subnets_cidr_blocks
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name                                          = "${local.cluster_name}-node-sg"
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  })
}

# Additional Security Group for Application Load Balancer
resource "aws_security_group" "alb" {
  name        = "${local.cluster_name}-alb-sg"
  description = "Security group for Application Load Balancer (Frontend & API)"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "HTTP from internet"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS from internet"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "All outbound traffic to EKS nodes"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name    = "${local.cluster_name}-alb-sg"
    Purpose = "frontend-and-api-ingress"
  })
}
