module "eks" {
  # Provider module
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.10"

  # Cluster identity
  name               = var.cluster_name
  kubernetes_version = var.cluster_version

  # Networking
  vpc_id     = var.vpc_id
  subnet_ids = var.private_subnets
  # control_plane_subnet_ids = var.private_subnets

  # Security groups - use only the centrally managed cluster SG
  create_security_group = false
  security_group_id     = var.cluster_security_group_id
  create_node_security_group    = false

  # Access / IAM
  # Auto Mode requires creation of some IAM resources for the controller/operator
  create_auto_mode_iam_resources = true
  # Allow admin permissions to the cluster creator when using Auto Mode
  enable_cluster_creator_admin_permissions = true

  # API access restrictions
  endpoint_public_access_cidrs = var.cluster_endpoint_public_access_cidrs

  # Compute (EKS Auto Mode)
  compute_config = {
    enabled    = true
    node_pools = ["general-purpose"]
  }

  # Managed cluster add-ons.
  addons = {
    coredns = {
      most_recent = true
    }

    kube-proxy = {
      most_recent = true
    }

    # AWS EBS CSI Driver needs an IAM role provided by this module (IRSA)
    aws-ebs-csi-driver = {
      most_recent              = true
      service_account_role_arn = aws_iam_role.ebs_csi_driver.arn
    }
  }

  # Metadata tags applied to created resources
  tags = var.tags
}
