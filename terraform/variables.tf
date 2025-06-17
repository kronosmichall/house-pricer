variable "aws_region" {
  type = string
  description = "AWS region"
}

variable "dynamodb_table_name" {
  type = string
  description = "DynamoDB table name"
}

variable "project_name" {
  type = string
  description = "AWS project name"
  default = "House pricer"
}

variable "lambda_timeout" {
  type = number
  description = "lambda timeout"
}

variable "lambda_memory_size" {
  type = number
  description = "max ram in MB"
}
