# Security Scan False-Positive Suppression

## Overview

This document describes how to suppress false-positive findings from `gosec` (Go SAST) and `govulncheck` (dependency vulnerability scanner) in the CI pipeline.

## gosec False Positives

### Inline Suppression

Suppress a specific rule on a single line using a `#nosec` comment:

```go
password := os.Getenv("SECRET") // #nosec G101 -- not a hardcoded credential, reading from env
```

Suppress multiple rules on one line:

```go
result, _ := exec.Command(cmd).Output() // #nosec G204,G204 -- cmd is validated upstream
```

### nosec Comment Format

```
// #nosec <RuleID> -- <mandatory reason explaining why this is a false positive>
```

The reason field is **required**. PRs with bare `// #nosec` annotations (no rule ID, no reason) will be rejected in code review.

### Common gosec Rule IDs

| Rule   | Description                                     | Typical FP Scenario                         |
|--------|-------------------------------------------------|---------------------------------------------|
| G101   | Hardcoded credentials                           | Variable named `*token*` but value from env |
| G104   | Errors unhandled                                | `defer f.Close()` where error is irrelevant |
| G204   | Subprocess launched with variable               | Controlled internal tooling                 |
| G304   | File path provided as taint input               | Path constructed from validated config      |
| G307   | Deferring a method which returns an error       | `defer rows.Close()` — error not actionable |
| G401   | Use of weak cryptographic primitive             | Non-security hash (checksums, etc.)         |
| G501   | Import of blocklisted package `crypto/md5`      | Non-security use (e.g., ETag generation)    |

Full rule list: https://github.com/securego/gosec#available-rules

### Bulk Suppression (Use Sparingly)

To suppress gosec for an entire file (e.g., a test helper with intentional insecure patterns):

```go
//go:build ignore
```

Or add a `nosec` directive at the top of the file — this is **strongly discouraged** for production code.

### Project-Level Configuration

For recurring false positives across the codebase, add exclusions to `.gosec.yaml`:

```yaml
# .gosec.yaml
exclude:
  - G104  # Only if justified and documented below
```

Any project-level exclusion MUST have a corresponding entry in this document explaining the justification.

#### Current Project-Level Exclusions

_None at this time._

---

## govulncheck False Positives

govulncheck only reports vulnerabilities in code paths that are **actually called** (not just imported). This means most false positives from simple imports are already filtered out.

### Situations Requiring Suppression

- A vulnerable function is imported but called only in test code that is not shipped
- A CVE affects a code path that is unreachable given this application's configuration

### How to Handle

govulncheck does not support inline suppression annotations. Options:

1. **Upgrade the dependency** — preferred
2. **Replace the dependency** — if upgrade is not available
3. **Document and accept risk** — add an entry to the table below with justification, approval, and review date

#### Accepted Vulnerability Exceptions

| CVE | Package | Affected Function | Reason Not Fixed | Approved By | Review Date |
|-----|---------|-------------------|------------------|-------------|-------------|
| _none_ | | | | | |

---

## Process for Adding a New Suppression

1. Confirm the finding is genuinely a false positive (reproduce the scenario manually)
2. Add the inline `// #nosec <ID> -- <reason>` comment or the govulncheck exception table entry
3. Include the suppression in the same PR as the code change
4. Request review from a second engineer
5. If a project-level `.gosec.yaml` exclusion is needed, update this document in the same PR

---

## GitHub Security Tab

SARIF results from both tools are uploaded to the **Security → Code scanning** tab on every PR and push to `main`. Findings visible there can be dismissed via the GitHub UI with a required reason. Dismissed findings are recorded in the repository's security audit log.

**Note:** SARIF upload requires the repository to be public or GitHub Advanced Security to be enabled. Contact your GitHub org admin if the Security tab is not visible.
