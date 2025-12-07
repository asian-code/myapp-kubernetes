module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.10"

  name               = var.cluster_name
  kubernetes_version = var.cluster_version

  vpc_id                   = var.vpc_id
  subnet_ids               = var.private_subnets
  control_plane_subnet_ids = var.private_subnets

  endpoint_public_access  = true
  endpoint_private_access = true

  # Enable cluster creator admin permissions for Auto Mode
  enable_cluster_creator_admin_permissions = true

  # Enable API and ConfigMap authentication mode for Auto Mode
  authentication_mode = "API_AND_CONFIG_MAP"

  # EKS Auto Mode - Compute Configuration
  # The module automatically creates the required IAM role for Auto Mode nodes
  cluster_compute_config = {
    enabled    = true
    node_pools = ["general-purpose", "system"]
  }

  # Auto Mode manages add-ons automatically, but we can specify versions if needed
  cluster_addons = {
    # Auto Mode will manage these automatically
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    eks-pod-identity-agent = {
      most_recent = true
    }
  }

  tags = var.tags
}
