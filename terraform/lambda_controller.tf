resource "aws_iam_role" "handler_lambda_exec" {
  name = "handler-lambda"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "log_group_for_controller" {
  name              = "/lambda/controller"
  retention_in_days = 3
}

data "archive_file" "controller_binary" {
  type        = "zip"
  source_file = "${path.module}/../release/binary/controller/bootstrap"
  output_path = "${path.module}/../release/lambda/controller.zip"
}

resource "aws_lambda_function" "controller" {

  filename      = "${path.module}/../release/lambda/controller.zip"
  function_name = "controller"
  role          = aws_iam_role.handler_lambda_exec.arn

  source_code_hash = data.archive_file.controller_binary.output_base64sha256

  runtime     = "provided.al2"
  memory_size = 1024
  timeout     = 15
  handler     = "bootstrap"

  logging_config {
    log_format = "JSON"
    log_group  = aws_cloudwatch_log_group.log_group_for_controller.id
  }
}