module "api_gateway" {
  source  = "terraform-aws-modules/apigateway-v2/aws"
  version = "~> 6.0"

  name          = "myhealth-api"
  description   = "myHealth API Gateway"
  protocol_type = "HTTP"

  cors_configuration = {
    allow_credentials = true
    allow_headers     = ["*"]
    allow_methods     = ["*"]
    allow_origins     = ["https://eric-n.com"]
    expose_headers    = ["*"]
    max_age           = 86400
  }

  # Custom domain
  domain_name                 = var.domain_name != "" ? var.domain_name : null
  domain_name_certificate_arn = var.acm_certificate_arn != "" ? var.acm_certificate_arn : null
  create_certificate          = false
  create_domain_records       = var.create_domain_records

  # Routes & Integration
  routes = {
    "$default" = {
      integration = {
        type               = "HTTP_PROXY"
        integration_method = "ANY"
        uri                = var.backend_url
        payload_format_version = "1.0"
      }
    }
  }

  # Stage
  stage_name = "dev"
  
  stage_access_log_settings = {
    create_log_group            = true
    log_group_retention_in_days = 7
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      resourcePath   = "$context.resourcePath"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }

  # API mapping with /api path
  api_mapping_key = "api"

  tags = var.tags
}

# Optional Route53 record for the custom domain
resource "aws_route53_record" "api_gateway" {
  count = var.create_domain_records && var.domain_name != "" && var.route53_zone_id != "" ? 1 : 0

  zone_id = var.route53_zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = module.api_gateway.domain_name_target_domain_name
    zone_id                = module.api_gateway.domain_name_hosted_zone_id
    evaluate_target_health = false
  }
}
