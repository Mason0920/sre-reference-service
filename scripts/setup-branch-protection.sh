#!/usr/bin/env bash
# Configure GitHub branch protection rules for the main branch.
# Usage: ./scripts/setup-branch-protection.sh [OWNER/REPO]
#
# Requires: gh CLI authenticated with repo admin scope.
# CI status check names match job `name` fields in .github/workflows/ci.yml.

set -euo pipefail

REPO="${1:-Mason0920/hslee}"
BRANCH="main"

echo "Configuring branch protection for ${REPO}@${BRANCH} ..."

gh api "repos/${REPO}/branches/${BRANCH}/protection" \
  --method PUT \
  --header 'Accept: application/vnd.github+json' \
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "Lint",
      "Test",
      "Build",
      "Smoke Test"
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true
}
EOF

echo "Branch protection configured successfully."
echo ""
echo "Rules applied to ${REPO}@${BRANCH}:"
echo "  - Required status checks: Lint, Test, Build, Smoke Test (strict mode)"
echo "  - Required approvals: 1"
echo "  - Dismiss stale reviews: yes"
echo "  - Force pushes: disabled"
echo "  - Direct deletions: disabled"
echo "  - Enforce for admins: yes"
echo "  - Required conversation resolution: yes"
