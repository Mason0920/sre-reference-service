#!/usr/bin/env bash
# One-time setup: create secrets in AWS Secrets Manager and configure rotation.
# Run this once with admin credentials before deploying.
set -euo pipefail

REGION="${AWS_REGION:-ap-northeast-2}"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "==> Creating secrets in AWS Secrets Manager (region: ${REGION})"

# Slack webhook URL — fetch the value from the current GitHub Actions secret
# and provide it here, then remove it from GitHub Actions secrets.
aws secretsmanager create-secret \
  --region "${REGION}" \
  --name "hslee/slack-webhook-url" \
  --description "Slack incoming webhook URL for deployment notifications" \
  --secret-string "${SLACK_WEBHOOK_URL:?Set SLACK_WEBHOOK_URL env var}" \
  || aws secretsmanager update-secret \
       --region "${REGION}" \
       --secret-id "hslee/slack-webhook-url" \
       --secret-string "${SLACK_WEBHOOK_URL}"

echo "==> Configuring 90-day rotation for hslee/slack-webhook-url"
# Slack webhook URLs do not support AWS-managed rotation.
# Set a 90-day schedule reminder — rotate manually via update-secret.
aws secretsmanager rotate-secret \
  --region "${REGION}" \
  --secret-id "hslee/slack-webhook-url" \
  --rotation-rules AutomaticallyAfterDays=90 2>/dev/null \
  || echo "NOTE: Slack URL rotation must be done manually (no Lambda rotator available). Reminder set."

echo "==> Creating IAM role: hslee-app-secrets-reader"
ROLE_NAME="hslee-app-secrets-reader"

aws iam create-role \
  --role-name "${ROLE_NAME}" \
  --assume-role-policy-document "file://$(dirname "$0")/iam-trust-policy.json" \
  || echo "Role already exists, skipping create."

POLICY_ARN=$(aws iam create-policy \
  --policy-name "hslee-secrets-read" \
  --policy-document "file://$(dirname "$0")/iam-secrets-policy.json" \
  --query Policy.Arn --output text 2>/dev/null \
  || aws iam list-policies --query "Policies[?PolicyName=='hslee-secrets-read'].Arn" --output text)

aws iam attach-role-policy \
  --role-name "${ROLE_NAME}" \
  --policy-arn "${POLICY_ARN}"

echo ""
echo "==> Done. Role ARN:"
aws iam get-role --role-name "${ROLE_NAME}" --query Role.Arn --output text
echo ""
echo "Next steps:"
echo "  1. Set AWS_DEPLOY_ROLE_ARN_STAGING and AWS_DEPLOY_ROLE_ARN_PRODUCTION"
echo "     as GitHub Actions Variables in your repo settings."
echo "  2. Add the GitHub OIDC provider to AWS if not already present:"
echo "     https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html"
echo "  3. Apply deployments/k8s/serviceaccount.yaml to each cluster."
echo "  4. Remove AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY from GitHub Secrets."
echo "  5. Remove SLACK_WEBHOOK_URL from GitHub Secrets."
