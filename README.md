# sre-reference-service

A small Go service wrapped with the SRE basics: CI/CD with rollback, metrics/logs/traces, SLOs with error budgets, alert runbooks, and load tests.

작은 Go HTTP 서비스 하나를 두고, 운영에 필요한 기본기(CI/CD, 관측성, SLO, 알람과 런북, 부하 테스트)를 한 저장소에 끝까지 붙여 본 레퍼런스 프로젝트다.

## 왜 만들었나

회사에서는 이미 있는 시스템 위에 기능을 하나씩 얹는 일이 많아서, 서비스 하나에 운영 기본기를 처음부터 끝까지 일관되게 붙여 본 적이 드물었다. 그래서 "서비스 하나를 운영 가능한 상태로 만든다는 것"의 최소 단위를 직접 정리해 보려고 만들었다.

진행은 AI 코딩 에이전트 팀(Paperclip)에게 이슈 단위로 맡기고, 나는 이슈를 쪼개고 리뷰하고 결과를 검증하는 역할을 했다. 이슈 번호(`HSLA-n`)는 커밋 메시지와 [ROADMAP](ROADMAP.md)에 남아 있다.

## 구성

| 영역 | 내용 | 위치 |
|---|---|---|
| 서비스 | Go `net/http` 서버. `/health`, `/ready`, `/version`, `/metrics`(라우트 패턴 기준 RED 지표), 구조화 로그(slog), OpenTelemetry 트레이싱. 종료 신호를 받으면 `/ready`를 먼저 503으로 바꾸고 drain 시간(`SHUTDOWN_DRAIN_DELAY`, 기본 5s) 뒤에 서버를 닫는다 | `cmd/hslee`, `pkg/httpmetrics`, `pkg/lifecycle`, `pkg/config` |
| CI | golangci-lint v2, `go test -race` + 커버리지, govulncheck, 바이너리 빌드 | `.github/workflows/ci.yml` |
| CD | GHCR 푸시 → EKS 스테이징 배포 → 스모크 테스트 → 실패 시 이전 이미지로 자동 롤백. AWS 인증은 GitHub OIDC 역할(정적 키 없음) | `.github/workflows/cd.yml`, `cmd/smoketest` |
| 관측성 | Prometheus, Grafana, Loki + Promtail, Jaeger, Alertmanager(Slack), blackbox exporter를 Docker Compose로 | `observability/` |
| SLO | 가용성 99.9%, p99 지연 등 SLI 정의, 기록 규칙, 에러 버짓 대시보드 | `docs/SLO.md`, `observability/prometheus/rules/slo.yml`, `grafana/` |
| 알람/런북 | 알람 규칙 10개, 규칙마다 대응 런북 1개 | `observability/prometheus/rules/alerting.yml`, `docs/runbooks/` |
| 부하 테스트 | k6 baseline / stress / 알람 검증 시나리오와 기준 결과 | `tests/load/` |

## 빠른 시작

```bash
# 서비스
make test        # go test -race
make build
go run ./cmd/hslee          # :8080

# 관측성 스택 (Grafana 관리자 비밀번호는 환경변수로 지정)
cd observability
GF_ADMIN_PASSWORD=change-me docker compose up -d
# Prometheus :9090 · Grafana :3000 · Alertmanager :9093 · Jaeger :16686 · Loki :3100

# 부하 테스트
k6 run tests/load/baseline.js
```

CD 워크플로는 기본으로 꺼져 있다. 저장소 변수 `DEPLOY_ENABLED=true`, `AWS_DEPLOY_ROLE_ARN`, `AWS_REGION`, `EKS_CLUSTER_NAME_STAGING`을 설정하면 동작한다. 자세한 내용은 [docs/CI-CD.md](docs/CI-CD.md).

## 문서

- [ARCHITECTURE](docs/ARCHITECTURE.md): 전체 구조
- [OBSERVABILITY](docs/OBSERVABILITY.md): 지표·로그·트레이스 수집 흐름과 대시보드
- [SLO](docs/SLO.md): SLI/SLO와 에러 버짓
- [runbooks](docs/runbooks/README.md): 알람별 대응 절차
- [CI-CD](docs/CI-CD.md): 파이프라인, 롤백 전략, 브랜치 보호
- [부하 테스트 결과](tests/load/results/baseline-results.md): 로컬 기준치 (macOS, OTel 끔)

## 한계

- 부하 테스트 수치는 로컬 머신 기준이라 운영 환경 성능을 뜻하지 않는다.
- 서비스 자체는 의도적으로 단순하다. 이 저장소의 대상은 비즈니스 로직이 아니라 그 주변의 운영 장치다.

## 관련

- 블로그: https://thereisnotherway.tistory.com/
- 같은 패턴을 쿠버네티스 쪽으로 확장한 실습: [platform-lab](https://github.com/Mason0920/platform-lab)

## License

MIT
