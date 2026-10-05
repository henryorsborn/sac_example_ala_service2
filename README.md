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

Local by default (`docker compose up`). For cloud deploys, regenerate the
scaffold with one of:

```bash
servicectl init demo --template=go-webapi --deploy=azure
servicectl init demo --template=go-webapi --deploy=gcp-cloud-run
servicectl init demo --template=go-webapi --deploy=gcp-gke-autopilot
```

The cloud overlays (Bicep for Azure, Terraform for GCP) ship passwordless
CD via Workload Identity Federation — no JSON keys in CI.

## API

A short list of endpoints this scaffold implements. The same shape is what
the upstream `servicectl` `go-webapi` template produces for any service.

| Method | Path             | Purpose                                                              |
|--------|------------------|----------------------------------------------------------------------|
| GET    | `/healthz`       | Liveness probe — 200 if the process is up.                           |
| POST   | `/v1/aliases`    | Create a short alias. Body: `{"alias_url": "...", "redirect_uri": "..."}`. Returns 201 on success, 409 on duplicate `alias_url`, 400 on validation error. |
| GET    | `/v1/aliases`    | List all aliases. Returns `{"count": N, "values": [...]}` (always a JSON object, never null). |
| GET    | `/:alias_url`    | Follow an alias. 302 to its `redirect_uri`, 404 if unknown.          |

Example session against a freshly-scaffolded service:

```bash
# Create an alias
curl -X POST http://localhost:3000/v1/aliases \
  -H 'Content-Type: application/json' \
  -d '{"alias_url":"github","redirect_uri":"https://github.com/henryorsborn"}'

# Follow it
curl -i http://localhost:3000/github
# HTTP/1.1 302 Found
# Location: https://github.com/henryorsborn

# List all aliases
curl http://localhost:3000/v1/aliases
# {"count":1,"values":[{"alias_id":1,"alias_url":"github", ...}]}
```

## Pre-commit hook

A `.githooks/pre-commit` script is bundled that runs the same smoke tests
CI does — locally, before each commit. To enable it once per checkout:

```bash
git config core.hooksPath .githooks
```

To bypass in an emergency (`--no-verify`).

## Secrets

Never commit secrets. `.env.example` shows the required variables. CI runs
[gitleaks](https://github.com/gitleaks/gitleaks) so accidental commits get
caught before they hit `main`.

## License

MIT.