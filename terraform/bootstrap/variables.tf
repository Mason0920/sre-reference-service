variable "project" {
  description = "Project name prefix for all resources"
  type        = string
  default     = "hslee"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-2"
}

variable "aws_account_id" {
  description = "AWS account ID (used to make bucket name globally unique)"
  type        = string
}
