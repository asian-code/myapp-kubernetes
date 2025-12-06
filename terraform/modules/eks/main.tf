module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.10"

  name            = var.cluster_name
  kubernetes_version = var.cluster_version

  vpc_id             = var.vpc_id
  subnet_ids         = var.private_subnets
  control_plane_subnet_ids = var.private_subnets

  endpoint_public_access  = true
  endpoint_private_access = true

  addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    ebs-csi-driver = {
      most_recent              = true
      service_account_role_arn = module.ebs_csi_irsa.iam_role_arn
    }
  }

  eks_managed_node_groups = {
    general = {
      name            = "${var.cluster_name}-node-group"
      use_name_prefix = true
      capacity_type   = "SPOT"

      instance_types = var.node_instance_types

      min_size     = var.node_min_size
      max_size     = var.node_max_size
      desired_size = var.node_desired_size

      vpc_security_group_ids = [var.node_security_group_id]

      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = 50
            volume_type           = "gp3"
            iops                  = 3000
            throughput            = 125
            encrypted             = true
            delete_on_termination = true
          }
        }
      }

      tags = var.tags
    }
  }

  tags = var.tags
}

# IRSA (IAM Roles for Service Accounts) for EBS CSI Driver
module "ebs_csi_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.40"

  role_name_prefix            = "${var.cluster_name}-ebs-csi-"
  attach_ebs_csi_policy       = true

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:ebs-csi-controller-sa"]
    }
  }

  tags = var.tags
}
