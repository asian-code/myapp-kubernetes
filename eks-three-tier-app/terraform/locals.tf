# Local variables
locals {
  cluster_name = var.cluster_name
  azs          = slice(data.aws_availability_zones.available.names, 0, 3)

  node_security_group_tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  }
}