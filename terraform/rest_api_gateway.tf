#data "aws_iam_policy_document" "my_api_policy" {
#  statement {
#    principals {
#      identifiers = ["*"]
#      type        = "*"
#    }
#
#    effect = "Allow"
#
#    actions = [
#      "execute-api:Invoke"
#    ]
#
#    resources = [
#      "arm:aws:execute-api:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*/*"
#    ]
#  }
#}
#
#resource "aws_api_gateway_rest_api" "my_api" {
#  name        = "my_api"
#  description = "This is my demo for EC2 orchestrator"
#  policy      = data.aws_iam_policy_document.my_api_policy.json
#
#  endpoint_configuration {
#    types = ["REGIONAL"]
#  }
#}
#
#resource "aws_api_gateway_resource" "proxy" {
#  rest_api_id = aws_api_gateway_rest_api.my_api.id
#  parent_id   = aws_api_gateway_rest_api.my_api.root_resource_id
#  path_part   = "{proxy+}"
#}
#
#resource "aws_api_gateway_method" "proxy" {
#  rest_api_id   = aws_api_gateway_rest_api.my_api.id
#  resource_id   = aws_api_gateway_resource.proxy.id
#  http_method   = "POST"
#  authorization = "AWS_IAM"
#}
#
#resource "aws_api_gateway_integration" "lambda_integration" {
#  rest_api_id             = aws_api_gateway_rest_api.my_api.id
#  resource_id             = aws_api_gateway_resource.proxy.id
#  http_method             = aws_api_gateway_method.proxy.http_method
#  integration_http_method = "POST"
#  type                    = "AWS_PROXY"
#  uri                     = aws_lambda_function.controller.invoke_arn
#}
#
#resource "aws_api_gateway_method" "lambda_root" {
#  rest_api_id = aws_api_gateway_rest_api.my_api.id
#  resource_id = aws_api_gateway_method.proxy_root.resource_id
#}
#
#
#resource "aws_api_gateway_deployment" "deployment" {
#  depends_on = [
#    aws_api_gateway_integration.lambda_integration,
#  ]
#
#  rest_api_id = aws_api_gateway_rest_api.my_api.id
#  stage_name  = "dev"
#}