output "state_bucket_name" {
  description = "S3 bucket name for Terraform remote state"
  value       = aws_s3_bucket.terraform_state.id
}

output "lock_table_name" {
  description = "DynamoDB table name for Terraform state locking"
  value       = aws_dynamodb_table.terraform_lock.name
}

output "github_oidc_provider_arn" {
  description = "ARN of the GitHub Actions OIDC provider (reference in environment IAM modules)"
  value       = aws_iam_openid_connect_provider.github_actions.arn
}
