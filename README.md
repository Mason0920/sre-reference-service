# hslee — SRE Foundation Project

This repository establishes the SRE foundation for the `hslee` company: automated CI/CD pipeline, code quality enforcement, testing infrastructure, and deployment automation.

## Goals

- Automated linting on every push and PR
- Automated test suite execution with clear pass/fail reporting
- Reproducible artifact builds in CI
- Automated deployment to target environments from `main`
- Post-deploy smoke tests and rollback strategy

## Prerequisites

- **Go** ≥ 1.21 (primary language) _or_ adjust for your stack
- **Docker** ≥ 24.0 (for containerized builds and local environment)
- **GitHub CLI** (`gh`) ≥ 2.40 — for PR management
- **make** — task runner (Makefile included)

> Adjust the prerequisites section to match the actual technology stack once it is confirmed.

## Setup

```bash
# 1. Clone the repository
git clone https://github.com/hslee/<repo>.git
cd <repo>

# 2. Install dependencies
make install

# 3. Run linting locally
make lint

# 4. Run tests
make test

# 5. Build artifacts
make build
```

### Environment Variables

Copy `.env.example` to `.env` and fill in the required values before running locally:

```bash
cp .env.example .env
```

| Variable        | Required | Description                          |
|-----------------|----------|--------------------------------------|
| `APP_ENV`       | Yes      | `development`, `staging`, `production` |
| `LOG_LEVEL`     | No       | `debug`, `info`, `warn`, `error` (default: `info`) |

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a full list of environment-specific settings.

## Common Tasks

| Command         | Description                          |
|-----------------|--------------------------------------|
| `make lint`     | Run linter (fail on errors)          |
| `make test`     | Run full test suite                  |
| `make build`    | Build production artifacts           |
| `make fmt`      | Auto-format source code              |
| `make clean`    | Remove build artifacts               |

## CI/CD

Every push to `main` and every PR runs the full CI pipeline:

1. **Lint** — fail on any linting error
2. **Test** — fail on any test failure
3. **Build** — produce deployable artifacts
4. **Deploy** — automatic deployment to target environment (on `main` only)
5. **Smoke test** — verify deployment health after each deploy

See [docs/CI-CD.md](docs/CI-CD.md) for pipeline configuration details.

## Project Structure

```
.
├── docs/                  # Project documentation
│   ├── ARCHITECTURE.md    # System architecture overview
│   ├── API.md             # API / tool reference
│   ├── CI-CD.md           # CI/CD pipeline documentation
│   └── CONTRIBUTING.md    # Contribution guide
├── .github/
│   └── workflows/         # GitHub Actions CI/CD workflows
├── Makefile               # Common development tasks
└── README.md              # This file
```

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for branch naming, commit conventions, PR workflow, and review process.

## License

Internal project — not for public distribution.
