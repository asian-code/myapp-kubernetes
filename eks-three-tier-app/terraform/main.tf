# # KMS key for EKS cluster encryption
# resource "aws_kms_key" "eks" {
#   count = var.enable_cluster_encryption ? 1 : 0

#   description             = "EKS Secret Encryption Key for ${local.cluster_name}"
#   deletion_window_in_days = var.kms_key_deletion_window_in_days
#   enable_key_rotation     = true

#   tags = merge(var.tags, {
#     Name = "${local.cluster_name}-eks-encryption-key"
#   })
# }

# resource "aws_kms_alias" "eks" {
#   count = var.enable_cluster_encryption ? 1 : 0

#   name          = "alias/${local.cluster_name}-eks-encryption-key"
#   target_key_id = aws_kms_key.eks[0].key_id
# }

# # EKS Cluster - Optimized with only necessary inputs for the project
# module "eks" {
#   source  = "terraform-aws-modules/eks/aws"
#   version = "~> 21.0.0"
#   # Essential cluster configuration
#   cluster_name    = local.cluster_name
#   cluster_version = var.cluster_version

#   # VPC configuration
#   vpc_id                   = module.vpc.vpc_id
#   subnet_ids               = module.vpc.private_subnets
#   control_plane_subnet_ids = module.vpc.private_subnets

#   # Cluster endpoint access
#   cluster_endpoint_public_access       = var.cluster_endpoint_public_access
#   cluster_endpoint_private_access      = var.cluster_endpoint_private_access
#   cluster_endpoint_public_access_cidrs = var.cluster_endpoint_public_access_cidrs

#   # Cluster encryption (conditional)
#   cluster_encryption_config = var.enable_cluster_encryption ? {
#     provider_key_arn = aws_kms_key.eks[0].arn
#     resources        = ["secrets"]
#   } : {}

#   # IRSA (IAM Roles for Service Accounts)
#   enable_irsa = var.enable_irsa

#   # Control plane logging
#   cluster_enabled_log_types              = var.enable_logging ? ["api", "audit", "authenticator", "controllerManager", "scheduler"] : []
#   cloudwatch_log_group_retention_in_days = var.log_retention_in_days
#   create_cloudwatch_log_group            = var.enable_logging
#   cloudwatch_log_group_kms_key_id        = var.enable_cluster_encryption ? aws_kms_key.eks[0].arn : null

#   # EKS Managed Node Groups - simplified configuration
#   eks_managed_node_groups = {
#     for name, config in var.node_groups : name => {
#       instance_types = config.instance_types
#       min_size       = config.min_size
#       max_size       = config.max_size
#       desired_size   = config.desired_size
#       capacity_type  = config.capacity_type
#       ami_type       = config.ami_type
#       disk_size      = config.disk_size

#       # Launch template
#       create_launch_template = true
#       launch_template_name   = "${local.cluster_name}-${name}"

#       # Update configuration
#       update_config = {
#         max_unavailable_percentage = 25
#       }

#       # Labels and taints
#       labels = config.labels
#       taints = config.taints

#       # Security groups
#       vpc_security_group_ids = [aws_security_group.additional_node_sg.id]

#       tags = merge(var.tags, {
#         "kubernetes.io/cluster/${local.cluster_name}" = "owned"
#         "NodeGroup"                                   = name
#       })
#     }
#   }

#   # Essential EKS Add-ons
#   cluster_addons = {
#     coredns = {
#       most_recent = true
#     }
#     kube-proxy = {
#       most_recent = true
#     }
#     vpc-cni = {
#       most_recent           = true
#       configuration_values = jsonencode({
#         env = {
#           ENABLE_PREFIX_DELEGATION = "true"
#           WARM_PREFIX_TARGET       = "1"
#           ENABLE_POD_ENI          = "true"
#         }
#       })
#     }
#     aws-ebs-csi-driver = {
#       most_recent = true
#     }
#   }

#   # Security group rules - only essential ones
#   node_security_group_additional_rules = {
#     ingress_self_all = {
#       description = "Node to node all ports/protocols"
#       protocol    = "-1"
#       from_port   = 0
#       to_port     = 0
#       type        = "ingress"
#       self        = true
#     }
#     ingress_cluster_to_node_all = {
#       description                   = "Cluster API to Nodegroup all traffic"
#       protocol                     = "-1"
#       from_port                    = 0
#       to_port                      = 0
#       type                         = "ingress"
#       source_cluster_security_group = true
#     }
#     egress_all = {
#       description = "Node all egress"
#       protocol    = "-1"
#       from_port   = 0
#       to_port     = 0
#       type        = "egress"
#       cidr_blocks = ["0.0.0.0/0"]
#     }
#   }

#   tags = merge(var.tags, {
#     "kubernetes.io/cluster/${local.cluster_name}" = "owned"
#   })
# }

# # IRSA role for EBS CSI Driver
# module "ebs_csi_irsa_role" {
#   source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
#   version = "~> 5.0"

#   role_name             = "${local.cluster_name}-ebs-csi-driver"
#   attach_ebs_csi_policy = true

#   oidc_providers = {
#     main = {
#       provider_arn               = module.eks.oidc_provider_arn
#       namespace_service_accounts = ["kube-system:ebs-csi-controller-sa"]
#     }
#   }

#   tags = var.tags
# }

# # IRSA role for AWS Load Balancer Controller  
# module "load_balancer_controller_irsa_role" {
#   source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
#   version = "~> 5.0"

#   role_name                              = "${local.cluster_name}-aws-load-balancer-controller"
#   attach_load_balancer_controller_policy = true

#   oidc_providers = {
#     main = {
#       provider_arn               = module.eks.oidc_provider_arn
#       namespace_service_accounts = ["kube-system:aws-load-balancer-controller"]
#     }
#   }

#   tags = var.tags
# }

# # IRSA role for External Secrets Operator
# module "external_secrets_irsa_role" {
#   source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
#   version = "~> 5.0"

#   role_name = "${local.cluster_name}-external-secrets-operator"

#   role_policy_arns = {
#     policy = "arn:aws:iam::aws:policy/SecretsManagerReadWrite"
#   }

#   oidc_providers = {
#     main = {
#       provider_arn               = module.eks.oidc_provider_arn
#       namespace_service_accounts = ["external-secrets:external-secrets-operator"]
#     }
#   }

#   tags = var.tags
# }

# # IRSA role for Cluster Autoscaler
# module "cluster_autoscaler_irsa_role" {
#   source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
#   version = "~> 5.0"

#   role_name                        = "${local.cluster_name}-cluster-autoscaler"
#   attach_cluster_autoscaler_policy = true
#   cluster_autoscaler_cluster_names = [local.cluster_name]

#   oidc_providers = {
#     main = {
#       provider_arn               = module.eks.oidc_provider_arn
#       namespace_service_accounts = ["kube-system:cluster-autoscaler"]
#     }
#   }

#   tags = var.tags
# }


# # Install AWS Load Balancer Controller
# resource "helm_release" "aws_load_balancer_controller" {
#   name       = "aws-load-balancer-controller"
#   repository = "https://aws.github.io/eks-charts"
#   chart      = "aws-load-balancer-controller"
#   namespace  = "kube-system"
#   version    = "1.6.2"

#   set {
#     name  = "clusterName"
#     value = module.eks.cluster_name
#   }

#   set {
#     name  = "serviceAccount.create"
#     value = "true"
#   }

#   set {
#     name  = "serviceAccount.name"
#     value = "aws-load-balancer-controller"
#   }

#   set {
#     name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
#     value = module.load_balancer_controller_irsa_role.iam_role_arn
#   }

#   set {
#     name  = "region"
#     value = var.region
#   }

#   set {
#     name  = "vpcId"
#     value = module.vpc.vpc_id
#   }

#   depends_on = [module.eks, module.load_balancer_controller_irsa_role]
# }

# # Install kube-prometheus-stack for comprehensive observability
# resource "helm_release" "kube_prometheus_stack" {
#   name       = "kube-prometheus-stack"
#   repository = "https://prometheus-community.github.io/helm-charts"
#   chart      = "kube-prometheus-stack"
#   namespace  = "monitoring"
#   version    = "57.2.0"

#   create_namespace = true

#   values = [
#     yamlencode({
#       # Prometheus configuration
#       prometheus = {
#         prometheusSpec = {
#           retention = "${var.monitoring_config.prometheus.retention_days}d"
#           scrapeInterval = var.monitoring_config.prometheus.scrape_interval
#           storageSpec = {
#             volumeClaimTemplate = {
#               spec = {
#                 storageClassName = "gp3"
#                 resources = {
#                   requests = {
#                     storage = var.monitoring_config.prometheus.storage_size
#                   }
#                 }
#               }
#             }
#           }
#           # Spread across AZs
#           affinity = {
#             podAntiAffinity = {
#               requiredDuringSchedulingIgnoredDuringExecution = [{
#                 labelSelector = {
#                   matchLabels = {
#                     "app.kubernetes.io/name" = "prometheus"
#                   }
#                 }
#                 topologyKey = "topology.kubernetes.io/zone"
#               }]
#             }
#           }
#         }
#       }
#       # Grafana configuration
#       grafana = {
#         adminPassword = var.monitoring_config.grafana.admin_password
#         persistence = {
#           enabled = true
#           storageClassName = "gp3"
#           size = var.monitoring_config.grafana.storage_size
#         }
#         # Enable ingress for Grafana
#         ingress = {
#           enabled = true
#           ingressClassName = "alb"
#           annotations = {
#             "kubernetes.io/ingress.class" = "alb"
#             "alb.ingress.kubernetes.io/scheme" = "internet-facing"
#             "alb.ingress.kubernetes.io/target-type" = "ip"
#             "alb.ingress.kubernetes.io/listen-ports" = "[{\"HTTP\": 80}, {\"HTTPS\": 443}]"
#           }
#           hosts = ["grafana.${local.cluster_name}.local"]
#         }
#       }
#       # Alertmanager configuration  
#       alertmanager = {
#         alertmanagerSpec = {
#           storage = {
#             volumeClaimTemplate = {
#               spec = {
#                 storageClassName = "gp3"
#                 resources = {
#                   requests = {
#                     storage = var.monitoring_config.alertmanager.storage_size
#                   }
#                 }
#               }
#             }
#           }
#         }
#       }
#     })
#   ]

#   depends_on = [
#     module.eks,
#     kubernetes_storage_class.gp3
#   ]
# }

# # Install Fluent Bit for log shipping
# resource "helm_release" "fluent_bit" {
#   name       = "fluent-bit"
#   repository = "https://fluent.github.io/helm-charts"
#   chart      = "fluent-bit"
#   namespace  = "logging"
#   version    = "0.43.0"

#   create_namespace = true

#   values = [
#     yamlencode({
#       config = {
#         outputs = <<EOF
# [OUTPUT]
#     Name cloudwatch_logs
#     Match *
#     region ${var.region}
#     log_group_name /aws/eks/${local.cluster_name}/fluent-bit
#     auto_create_group On
#     log_stream_prefix fluent-bit-
# EOF
#         filters = <<EOF
# [FILTER]
#     Name kubernetes
#     Match kube.*
#     Kube_URL https://kubernetes.default.svc:443
#     Merge_Log On
#     K8S-Logging.Parser On
#     K8S-Logging.Exclude Off
# EOF
#       }
#       # Spread DaemonSet across all nodes
#       tolerations = [
#         {
#           operator = "Exists"
#         }
#       ]
#     })
#   ]

#   depends_on = [module.eks]
# }

# # Install External Secrets Operator
# resource "helm_release" "external_secrets" {
#   name       = "external-secrets"
#   repository = "https://charts.external-secrets.io"
#   chart      = "external-secrets"
#   namespace  = "external-secrets"
#   version    = "0.9.11"

#   create_namespace = true

#   values = [
#     yamlencode({
#       # Enable metrics for monitoring
#       metrics = {
#         enabled = true
#         service = {
#           port = 8080
#         }
#       }
#       serviceAccount = {
#         annotations = {
#           "eks.amazonaws.com/role-arn" = module.external_secrets_irsa_role.iam_role_arn
#         }
#       }
#     })
#   ]

#   depends_on = [
#     module.eks,
#     module.external_secrets_irsa_role
#   ]
# }

# # Install Cluster Autoscaler
# resource "helm_release" "cluster_autoscaler" {
#   name       = "cluster-autoscaler"
#   repository = "https://kubernetes.github.io/autoscaler"
#   chart      = "cluster-autoscaler"
#   namespace  = "kube-system"
#   version    = "9.29.0"

#   values = [
#     yamlencode({
#       autoDiscovery = {
#         clusterName = module.eks.cluster_name
#       }
#       awsRegion = var.region
#       rbac = {
#         serviceAccount = {
#           create = true
#           name = "cluster-autoscaler"
#           annotations = {
#             "eks.amazonaws.com/role-arn" = module.cluster_autoscaler_irsa_role.iam_role_arn
#           }
#         }
#       }
#       extraArgs = {
#         "scale-down-delay-after-add" = "10m"
#         "scale-down-unneeded-time" = "10m"
#         "scale-down-utilization-threshold" = "0.5"
#       }
#     })
#   ]

#   depends_on = [
#     module.eks, 
#     module.cluster_autoscaler_irsa_role
#   ]
# }

# # Install Calico CNI for Network Policies
# resource "helm_release" "calico" {
#   name       = "calico"
#   repository = "https://docs.projectcalico.org/charts"
#   chart      = "tigera-operator"
#   namespace  = "tigera-operator"
#   version    = "v3.26.1"

#   create_namespace = true

#   depends_on = [module.eks]
# }

# # Create namespaces for the simplified three-tier application
# resource "kubernetes_namespace" "namespaces" {
#   for_each = toset([
#     "frontend",           # Frontend React/Vue/Angular pods
#     "backend",            # Backend API pods  
#     "database",           # MongoDB pods (StatefulSet)
#     "monitoring",         # Prometheus, Grafana, Alertmanager
#     "logging",            # Fluent Bit
#     "external-secrets"    # External Secrets Operator
#   ])

#   metadata {
#     name = each.key

#     labels = {
#       # Pod security standards - MongoDB needs privileged for storage
#       "pod-security.kubernetes.io/enforce" = each.key == "database" ? "privileged" : "restricted"
#       "pod-security.kubernetes.io/audit"   = each.key == "database" ? "privileged" : "restricted" 
#       "pod-security.kubernetes.io/warn"    = each.key == "database" ? "privileged" : "restricted"

#       # Tier identification
#       "tier" = each.key

#       # Network policy selector
#       "network-policy" = "enabled"
#     }

#     annotations = {
#       "description" = {
#         "frontend"         = "Frontend web application pods"
#         "backend"          = "Backend API service pods"  
#         "database"         = "MongoDB database pods (StatefulSet)"
#         "monitoring"       = "Observability stack (Prometheus, Grafana)"
#         "logging"          = "Centralized logging (Fluent Bit)"
#         "external-secrets" = "Secret management operator"
#       }[each.key]
#     }
#   }

#   depends_on = [module.eks]
# }

# # Create storage classes
# resource "kubernetes_storage_class" "gp3" {
#   metadata {
#     name = "gp3"
#     annotations = {
#       "storageclass.kubernetes.io/is-default-class" = "true"
#     }
#   }

#   storage_provisioner    = "ebs.csi.aws.com"
#   reclaim_policy         = "Delete"
#   allow_volume_expansion = true
#   volume_binding_mode    = "WaitForFirstConsumer"

#   parameters = {
#     type      = "gp3"
#     fsType    = "ext4"
#     encrypted = "true"
#   }

#   depends_on = [module.eks]
# }

# # Remove default gp2 storage class
# resource "kubernetes_annotations" "gp2_default" {
#   api_version = "storage.k8s.io/v1"
#   kind        = "StorageClass"

#   metadata {
#     name = "gp2"
#   }

#   annotations = {
#     "storageclass.kubernetes.io/is-default-class" = "false"
#   }

#   depends_on = [kubernetes_storage_class.gp3]
# }
