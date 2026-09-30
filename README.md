# ala_service

> Go 1.22+ + net/http + PostgreSQL (distroless runtime) — scaffolded by [servicectl](https://github.com/henryorsborn/servicectl).

## Quick start

```bash
# Local dev with Postgres in Docker
docker compose -f docker-compose.dev.yml up

# Or run without containers
cp .env.example .env
go mod download
go run ./cmd/server
```

The service listens on `:3000`. Health check: `GET /healthz`. Readiness: `GET /readyz`.

## Scripts

| Command | What it does |
|---|---|
| `go test ./...` | Run tests with coverage; fails below 80% |
| `go vet ./...` | Static analysis |
| `go run ./cmd/server` | Dev server |
| `docker build -t ala_service .` | Multi-stage production build (distroless runtime) |
| `docker compose -f docker-compose.dev.yml up` | Local stack: this service + Postgres |

## Layout

```
.
├── cmd/server/                       # binary entrypoint
├── internal/ala_service/  # HTTP handlers + tests
├── Dockerfile                        # multi-stage build, distroless runtime, nonroot
├── docker-compose.dev.yml            # local dev: this service + Postgres
├── .devcontainer/                    # VS Code remote dev
├── .github/workflows/ci.yml          # GitHub Actions: vet, test, build, scan
├── azure-pipelines.yml               # Azure DevOps equivalent
├── go.mod                            # module declaration
├── .env.example                      # all env vars, no plaintext secrets
└── .gitleaks.toml                    # secrets scanning config
```

The Go layout follows the standard `cmd/` + `internal/` convention: `cmd/server/main.go` is the binary, `internal/ala_service/` holds packages that should not be imported by other modules.

## CI

Default CI runs on [github-actions](https://github.com/features/actions). Each PR runs:

1. **vet** — `go vet ./...`
2. **test** — `go test` with `-race` and a 80% coverage threshold
3. **build** — multi-stage Docker build (distroless static runtime)
4. **scan** — Trivy container scan; HIGH/CRITICAL = fail

5. **publish** — on push to `main`, push image to ghcr


## Deployment

Local only for now (`docker compose up`). Cloud deploy flags coming next:
`--deploy=azure`, `--deploy=aws`, `--deploy=gcp-cloud-run`.

## Secrets

Never commit secrets. `.env.example` shows the required variables. CI runs
[gitleaks](https://github.com/gitleaks/gitleaks) so accidental commits get
caught before they hit `main`.

## License

MIT.