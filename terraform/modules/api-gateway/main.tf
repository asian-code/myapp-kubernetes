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
  domain_name                = "myhealth.eric-n.com"
  domain_name_certificate_arn = var.acm_certificate_arn
  create_certificate         = false
  create_domain_records      = false

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
