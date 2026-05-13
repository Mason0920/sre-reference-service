output "cicd_role_arn" {
  description = "ARN of the GitHub Actions CI/CD IAM role"
  value       = aws_iam_role.cicd.arn
}

output "cicd_role_name" {
  description = "Name of the GitHub Actions CI/CD IAM role"
  value       = aws_iam_role.cicd.name
}

output "github_oidc_provider_arn" {
  description = "ARN of the GitHub Actions OIDC provider"
  value       = aws_iam_openid_connect_provider.github_actions.arn
}
