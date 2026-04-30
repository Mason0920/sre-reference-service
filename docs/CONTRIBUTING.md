# Contributing Guide

## Quick Reference

| What                 | Rule                                              |
|----------------------|---------------------------------------------------|
| Branch name          | `hsla-<N>/<short-description>`                   |
| Commit format        | Conventional Commits (`feat:`, `fix:`, `docs:` …) |
| PR required          | Yes — for all logic, config, infra changes        |
| Direct-to-main OK    | Typos, comment-only, minor doc fixes only         |
| Merge owner          | Engineer who opened the PR                        |
| Merge method         | `gh pr merge <number> --merge` (no squash/rebase) |

## Branch Naming

```
hsla-<N>/<short-description>
```

- `N` = Paperclip issue number (e.g. `HSLA-6`)
- Description: lowercase, hyphens, max 5 words

Examples:
```
hsla-6/initial-project-docs
hsla-12/add-lint-config
hsla-7/setup-github-actions
```

## Commit Conventions

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <short description>
```

Types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `perf`

Rules:
- Lowercase after colon
- No period at end
- Under 72 characters
- Reference issue ID in commit body when applicable

Examples:
```
feat: add GitHub Actions CI workflow
fix: resolve nil pointer in health check handler
docs: add architecture overview
chore: add golangci-lint config
```

**Agent commits** must include:
```
Co-Authored-By: Paperclip <noreply@paperclip.ing>
```

## PR Workflow

### Opening a PR

1. Create branch: `git checkout -b hsla-<N>/<description>`
2. Make changes, commit with conventional commit messages
3. Push: `git push -u origin hsla-<N>/<description>`
4. Open PR: `gh pr create --title "<type>: <description>" --body "$(cat pr-body.md)"`

### PR Body Template

```markdown
## What changed
<Brief description of the changes>

## Why
<Motivation and context — link to Paperclip issue>

## How to test
<Steps to verify the changes locally>

## Related
Closes HSLA-N
```

### Review Process

1. Engineer opens PR, sets originating Paperclip issue to `in_review`
2. Engineer @-mentions Code Reviewer and Product Owner on the issue with PR link
3. **Code Reviewer** reviews: correctness, security, code style, simplicity
4. **Product Owner** reviews: intent match, scope discipline, acceptance criteria
5. Both post verdict on the Paperclip issue
6. Engineer merges when both approve: `gh pr merge <number> --merge`
7. Engineer sets Paperclip issue to `done`

### What Requires a PR

- Code logic changes
- CI/CD workflow changes (`.github/workflows/`)
- Configuration changes (`Makefile`, linter config, etc.)
- Infrastructure / Terraform changes
- Agent config changes

### Direct-to-Main (No PR)

- Typo fixes in documentation
- Comment-only code changes
- Minor doc formatting

Must reference issue ID in commit body.

## Local Development Checklist

Before pushing:

- [ ] `make lint` passes locally
- [ ] `make test` passes locally
- [ ] `make build` succeeds
- [ ] No secrets or credentials committed
- [ ] Commit message follows Conventional Commits format

## Labels

Apply one primary label to each PR:

| Label     | When to use                              |
|-----------|------------------------------------------|
| `feature` | New functionality                        |
| `bug`     | Defect fix                               |
| `docs`    | Documentation only                       |
| `chore`   | Tooling, config, maintenance             |
| `infra`   | Infrastructure / CI-CD changes           |
| `agent`   | Agent config or Paperclip workflow change |

## CI Requirements

All PRs must pass:
1. **Lint** — zero lint errors
2. **Test** — zero test failures
3. **Build** — successful artifact build

CI failures block merge. Fix CI before requesting review.

## Merge Rules

- Code Reviewer and Product Owner must both approve
- CI must be green
- No force pushes to `main`
- No squash or rebase — use merge commits to preserve history
- Engineer is always the merge owner; reviewers do not merge

## Getting Help

- Check [ARCHITECTURE.md](ARCHITECTURE.md) for system context
- Open a Paperclip issue for questions or blockers
- Tag the relevant agent on the Paperclip issue thread
