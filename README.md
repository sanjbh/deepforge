# deepforge

A production-grade, framework-free multi-agent deep research pipeline written in Go. Given a research query, deepforge autonomously plans targeted searches, executes them in parallel, synthesizes the results into a structured report, and delivers it via email — all observable end-to-end through OpenTelemetry traces and structured logs.

Built without LangChain, LlamaIndex, or any AI orchestration framework. Every agent is a plain Go struct. Every design decision has an explicit rationale.

---

## How It Works

```
┌──────────────────────────────────────────────────────────────────┐
│                           deepforge                              │
│                                                                  │
│  ┌─────────────┐                                                 │
│  │   Planner   │  LLM → generates N targeted search queries      │
│  └──────┬──────┘                                                 │
│         │ WebSearchPlan                                          │
│         ▼                                                        │
│  ┌──────────────────────────────────┐                            │
│  │     SearchAgent  (errgroup)      │  parallel execution        │
│  │  ┌──────────┐  ┌──────────┐      │                            │
│  │  │ Search 1 │  │ Search 2 │ ...  │  SearXNG (self-hosted)     │
│  │  └──────────┘  └──────────┘      │                            │
│  └───────────────────┬──────────────┘                            │
│                      │ []SearchResult                            │
│                      ▼                                           │
│  ┌───────────────────────┐                                       │
│  │      WriterAgent      │  LLM → structured Markdown report     │
│  └───────────┬───────────┘                                       │
│              │ ReportData                                        │
│              ▼                                                   │
│  ┌───────────────────────┐                                       │
│  │      EmailAgent       │  SendGrid / Mailhog / File            │
│  └───────────────────────┘                                       │
│                                                                  │
│  LLM:           Gemini 2.0 Flash → Ollama fallback               │
│  Observability: zap + OpenTelemetry → Grafana Tempo              │
└──────────────────────────────────────────────────────────────────┘
```

---

## Features

- **Framework-free multi-agent pipeline** — four agents (Planner, Searcher, Writer, Emailer) with clean separation of concerns; no AI framework dependency
- **Parallel search execution** — `golang.org/x/sync/errgroup` drives concurrent searches with automatic context cancellation and race-free result collection
- **LLM provider abstraction** — a `Provider` interface decouples agents from LLM implementations; Gemini and Ollama are wired at startup, invisible to agents
- **Transparent LLM failover** — `FallbackProvider` wraps primary and secondary providers; both errors are surfaced if both fail
- **Retry with exponential backoff** — LLM calls retry up to three times with 1s/2s/4s delays; context cancellation is respected during both active calls and retry sleeps
- **Structured JSON output** — `jsonschema` reflection drives Gemini's structured output mode for reliable, typed responses from the Planner and Writer agents
- **Self-hosted search** — SearXNG aggregates Google, Bing, and DuckDuckGo; no search API key required
- **Pluggable email delivery** — SendGrid for production, Mailhog for dev/testing, file output for local runs; selection is automatic based on config
- **Full observability** — structured JSON logs via `go.uber.org/zap` and distributed traces via OpenTelemetry exported to Grafana Tempo
- **Graceful shutdown** — `signal.NotifyContext` ties the pipeline context to `SIGTERM`/`SIGINT`; a 5-minute `context.WithTimeout` budget wraps the entire run
- **Deployable anywhere** — Docker Compose for local dev, Kubernetes Job and CronJob manifests for production; tested on k0s

---

## Why Go

The Go ecosystem has no equivalent of LangChain or LlamaIndex — no dominant agentic AI framework to reach for. That's not a gap; it's an opportunity to build the pipeline the way Go actually works.

Multi-agent pipelines have a natural concurrency shape: the Planner produces a set of independent searches, all of which can run simultaneously, with results fanning back in before the Writer proceeds. Go's `errgroup` models this exactly — parallel goroutines, automatic context cancellation on the first failure, and race-free result collection via pre-indexed slices. There's no async/await friction, no thread pool configuration, no promise chaining. The concurrency is structural.

The same applies to cancellation. A single context flows from `signal.NotifyContext` through the pipeline, the `errgroup`, each agent, the `Provider`, and down to every HTTP call. SIGTERM at the OS level unwinds the entire chain in milliseconds. This isn't something bolted on — it's how Go programs are supposed to be written.

The result is a pipeline that's fast because it's concurrent where it should be, not because of runtime magic. And because there's no framework mediating between components, every agent, interface, and retry path is directly readable and directly testable.

---

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| LLM (primary) | Gemini 2.0 Flash via OpenAI-compatible API |
| LLM (fallback) | Ollama — `qwen3:8b-q4_K_M` or any local model |
| Search | SearXNG (self-hosted meta-search) |
| Email (prod) | SendGrid |
| Email (dev) | Mailhog |
| Logging | `go.uber.org/zap` (structured JSON) |
| Tracing | OpenTelemetry → Grafana Tempo |
| Dashboards | Grafana |
| Container | Docker / Docker Compose |
| Orchestration | Kubernetes (tested on k0s) |

---

## Project Structure

```
deepforge/
├── cmd/deepforge/main.go         # entrypoint — signal handling, wiring, startup
├── internal/
│   ├── agents/
│   │   ├── planner.go            # PlannerAgent — generates search plan via LLM
│   │   ├── searcher.go           # SearchAgent  — executes search + summarizes
│   │   ├── writer.go             # WriterAgent  — synthesizes report via LLM
│   │   └── emailer.go            # EmailAgent   — renders HTML + delivers
│   ├── llm/
│   │   ├── provider.go           # Provider interface + OpenAICompatibleProvider
│   │   ├── client.go             # Generate / GenerateStructured + retry logic
│   │   └── fallback.go           # FallbackProvider — primary → secondary
│   ├── models/types.go           # shared domain types with jsonschema tags
│   ├── pipeline/research.go      # ResearcherPipeline — orchestrates all agents
│   ├── search/searxng.go         # SearXNG HTTP client
│   └── tools/
│       ├── sender.go             # EmailSender interface
│       ├── sendgrid.go           # SendGrid implementation
│       ├── mailhog.go            # Mailhog SMTP implementation
│       └── file_sender.go        # File-based implementation (local dev)
├── config/config.go              # viper + go-playground/validator config loading
├── observability/
│   ├── logger.go                 # zap logger factory
│   ├── tracer.go                 # OpenTelemetry tracer + OTLP exporter
│   └── observability.go          # Observability struct — logger + tracer bundle
├── deploy/
│   ├── Dockerfile                # two-stage build (golang:alpine → alpine)
│   ├── docker-compose.yml        # full local stack
│   ├── .env.docker.example       # Docker env template
│   ├── searxng/settings.yml      # SearXNG configuration
│   ├── tempo.yaml                # Tempo configuration
│   └── k8s/
│       ├── namespace.yaml
│       ├── configmap.yaml
│       ├── secret.yaml
│       ├── job.yaml              # one-shot research run
│       ├── cronjob.yaml          # scheduled research
│       └── infra/
│           ├── searxng/          # Deployment, Service, ConfigMap
│           ├── mailhog/          # Deployment, Service
│           ├── tempo/            # Deployment, Service, PVC, ConfigMap
│           └── grafana/          # Deployment, Service, PVC, ConfigMap
├── .env.example                  # local dev env template
└── Makefile                      # all workflows automated
```

---

## Design Decisions

### No AI framework

Agents are plain Go structs. There is no LangChain, no LlamaIndex, no orchestration SDK. The entire pipeline is ~1,000 lines of idiomatic Go that any engineer can read, debug, and extend without framework knowledge.

### Provider interface

```go
type Provider interface {
    Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
    GenerateStructured(ctx context.Context, systemPrompt, userPrompt string, schema *jsonschema.Schema) (string, error)
    Name() string
}
```

All four agents program against this interface. The concrete implementations — `OpenAICompatibleProvider` for Gemini, the same type configured for Ollama — are wired once in `main.go` and never referenced by agents. Adding a new LLM provider requires zero changes to agent code.

### FallbackProvider

`FallbackProvider` wraps two `Provider` values and implements the same interface. The primary is tried first; on any error the secondary is attempted. If both fail, a combined error is returned with both failure messages. The wrapping is transparent to the pipeline.

### Retry with exponential backoff

```go
delay := min(baseDelay*(1<<attempt), maxDelay)

select {
case <-time.After(delay):
case <-ctx.Done():
    return ctx.Err()
}
```

Delays are computed via bit-shifting (1s → 2s → 4s, capped at 10s). A `select` in the retry sleep means context cancellation — from a signal or pipeline timeout — unwinds the delay immediately rather than blocking for the full sleep duration.

### errgroup for parallel search

```go
g, gCtx := errgroup.WithContext(ctx)
results := make([]models.SearchResult, len(planner.Searches))

for i, item := range planner.Searches {
    g.Go(func() error {  // Go 1.22+: loop vars are per-iteration, no capture needed
        result, err := p.searcher.Search(gCtx, item)
        if err != nil {
            return err
        }
        results[i] = *result
        return nil
    })
}

if err := g.Wait(); err != nil { ... }
```

Results are pre-allocated by index — no mutex, no append races. If any search goroutine returns an error, `errgroup` cancels the shared context, unwinding all other in-flight searches immediately via `gCtx`.

### Structured output with jsonschema reflection

`PlannerAgent` and `WriterAgent` use `GenerateStructured`, which passes a JSON Schema derived from the response struct to Gemini's `response_format` parameter. This produces structured, typed JSON rather than free-form text that requires parsing heuristics, and works consistently across Gemini and Ollama.

### EmailSender interface

```go
type EmailSender interface {
    Send(subject string, htmlBody string) error
}
```

`EmailAgent` holds this interface. The concrete implementation — `SendGridEmailSender`, `MailHogEmailSender`, or `FileEmailSender` — is selected in `main.go` based on config at startup. No agent code changes are needed to swap delivery backends.

### Context propagation

A single context originates from `signal.NotifyContext` in `main.go`. It flows through `context.WithTimeout` (5-minute budget), into the pipeline, into `errgroup.WithContext`, into each agent, into the `Provider`, and down to every HTTP call. Cancellation at any level — SIGTERM, timeout, or search failure — propagates immediately through the entire chain.

### Graceful shutdown

```go
ctx, stop := signal.NotifyContext(context.TODO(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
defer cancel()
```

SIGTERM or SIGINT cancels the context. The 5-minute timeout is a hard budget for the full pipeline run. All in-flight LLM calls and HTTP requests respect context cancellation and terminate cleanly.

### OpenTelemetry instrumentation pattern

Every agent method follows the same three-line pattern:

```go
ctx, span := e.obs.Tracer.Start(ctx, "AgentName.Method")
defer span.End()
// ... on error:
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
```

The root span `ResearchPipeline.Run` provides a single trace covering the full execution. Child spans per agent give precise timing for each stage.

### Two-stage Docker build

Stage 1 compiles the binary in `golang:1.25-alpine` with `-ldflags="-s -w"` to strip debug symbols. Stage 2 copies only the binary into `alpine:latest` with `ca-certificates` added. The result is a minimal production image with no Go toolchain included.

### Optional `.env` loading

```go
if _, err := os.Stat(".env"); err == nil {
    viper.SetConfigFile(".env")
    viper.ReadInConfig()
}
```

A missing `.env` is not an error. The same binary works in local dev (reading `.env`), Docker Compose (environment variables injected by the runtime), and Kubernetes (ConfigMap + Secret) without modification.

### Viper `BindEnv` for reliable env var mapping

`viper.AutomaticEnv()` alone does not reliably populate struct fields through `mapstructure` tags during `Unmarshal`. Each config field has an explicit `viper.BindEnv()` call to guarantee correct mapping regardless of environment.

### OTLP endpoint hardcoded in Compose

`OTLP_ENDPOINT` is set to `tempo:4318` (the Docker service name) directly in `docker-compose.yml` rather than read from `.env`. This prevents the `localhost` value from a local `.env` from leaking into the container and silently breaking trace export.

---

## Quick Start

### Prerequisites

- Go 1.25+
- Docker + Docker Compose
- A [Gemini API key](https://aistudio.google.com/app/apikey)

### Everything in Docker

```bash
git clone https://github.com/sanjbh/deepforge
cd deepforge

cp deploy/.env.docker.example deploy/.env.docker
# edit deploy/.env.docker — set GEMINI_API_KEY at minimum

make up
make logs
```

Once the pipeline finishes, open these in your browser:

- **http://localhost:8025** — Mailhog web console. The generated research report will be sitting there as a fully formatted HTML email, ready to read.
- **http://localhost:3000** — Grafana. Go to Explore → Tempo → Search, filter by service name `deepforge`, and you'll see the full trace for the pipeline run: every agent, every span, wall-clock timing, and any errors recorded exactly where they happened.

### App local, infra in Docker

```bash
# start infra only
docker compose --env-file deploy/.env.docker \
  -f deploy/docker-compose.yml \
  up searxng tempo grafana mailhog

# configure local env
cp .env.example .env
# edit .env — set GEMINI_API_KEY

go run ./cmd/deepforge -query "Latest developments in Rust async programming"
```

---

## Configuration

All configuration is via environment variables. Copy `.env.example` to `.env` for local development.

| Variable | Description | Required |
|---|---|---|
| `GEMINI_API_KEY` | Gemini API key | Yes |
| `GEMINI_MODEL` | Gemini model name (e.g. `gemini-2.0-flash`) | Yes |
| `GEMINI_BASE_URL` | Gemini OpenAI-compatible base URL | Yes |
| `OLLAMA_BASE_URL` | Ollama base URL | Yes |
| `OLLAMA_MODEL` | Ollama model name | Yes |
| `HOW_MANY_SEARCHES` | Number of parallel searches (1–10) | Yes |
| `RESULTS_PER_SEARCH` | Results returned per query (1–20) | Yes |
| `SEARXNG_BASE_URL` | SearXNG instance URL | Yes |
| `SENDGRID_API_KEY` | SendGrid API key — enables SendGrid delivery | No |
| `FROM_EMAIL` | Sender email address | No |
| `TO_EMAIL` | Recipient email address | No |
| `MAILHOG_HOST` | Mailhog SMTP host — enables Mailhog delivery | No |
| `MAILHOG_PORT` | Mailhog SMTP port | No |
| `SERVICE_NAME` | Service name for observability | Yes |
| `SERVICE_VERSION` | Service version | Yes |
| `LOG_LEVEL` | Log level (`debug`/`info`/`warn`/error`) | Yes |
| `OTLP_ENDPOINT` | OpenTelemetry collector endpoint | Yes |
| `DEEPFORGE_QUERY` | Default research query (overridden by `-query` flag) | Yes |

**Email sender selection is automatic:**

- `SENDGRID_API_KEY` present → SendGrid
- `MAILHOG_HOST` + `MAILHOG_PORT` present → Mailhog SMTP
- Neither set → writes timestamped HTML reports to `emails/`

---

## Observability

deepforge emits structured JSON logs via `zap` and distributed traces via OpenTelemetry. Traces are exported to Grafana Tempo over OTLP/HTTP.

Each pipeline run produces a single `ResearchPipeline.Run` root span with child spans per agent:

```
ResearchPipeline.Run        (~4m total)
├── PlannerAgent.Plan        (12s)
├── SearchAgent.Search       (20s) ─┐
├── SearchAgent.Search       (39s)  ├─ parallel via errgroup
├── SearchAgent.Search       (55s) ─┘
├── WriterAgent.Write        (1m 14s)
└── EmailAgent.Send          (1m 29s)
```

Access Grafana at `http://localhost:3000` → Explore → Tempo → Search by service name `deepforge`.

---

## Kubernetes Deployment

### Prerequisites

- `kubectl` installed and configured (see cluster setup below)
- `local-path-provisioner` for PVC support on single-node clusters

### Setting up a cluster with k0s

If you don't have a cluster, k0s is the fastest path to a single-node Kubernetes cluster on any Linux machine:

```bash
curl -sSLf https://get.k0s.sh | sudo sh
sudo k0s install controller --single
sudo k0s start
```

k0s generates its kubeconfig at `/var/lib/k0s/pki/admin.conf`. Export it so `kubectl` can find it:

```bash
mkdir -p ~/.kube
sudo cp /var/lib/k0s/pki/admin.conf ~/.kube/config
sudo chown $USER ~/.kube/config
export KUBECONFIG=~/.kube/config
```

Add the `export` to your `~/.bashrc` or `~/.zshrc` so it persists across sessions. Verify with `kubectl get nodes` — your node should show `Ready`.

Then install the local-path-provisioner so PVCs work on a single node:

```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```

### Deploying deepforge

```bash
# build and import image into containerd
make k8s-load

# deploy everything — namespace, infra, config, job — in order
make k8s-deploy
```

### View results

```bash
# expose Mailhog (8025) and Grafana (3000) to localhost
make port-forward
```

Same endpoints as Docker:

- **http://localhost:8025** — Mailhog web console. The research report delivered as a formatted HTML email.
- **http://localhost:3000** — Grafana with Tempo pre-wired. Full distributed trace for the pipeline run, showing exactly how long each agent took and where any errors occurred.

```bash
# check pod status
make k8s-status

# follow logs
make k8s-logs
```

### Scheduled research (CronJob)

Edit `deploy/k8s/cronjob.yaml` to set your query and schedule:

```yaml
schedule: "0 7 * * *"   # every day at 7am
args:
- "--query"
- "Your scheduled research topic"
```

Apply:

```bash
kubectl apply -f deploy/k8s/cronjob.yaml
```

The CronJob uses `concurrencyPolicy: Forbid` to prevent overlapping runs and `ttlSecondsAfterFinished: 3600` to clean up completed Jobs automatically.

### Cleanup

```bash
make k8s-clean   # remove resources, keep PVCs and namespace
make k8s-nuke    # full wipe including PVCs and namespace
```

---

## Makefile Reference

| Target | Description |
|---|---|
| `make up` | Start full Docker Compose stack |
| `make down` | Stop Docker Compose stack |
| `make logs` | Follow deepforge container logs |
| `make k8s-load` | Build image and import into k0s containerd |
| `make k8s-deploy` | Deploy namespace + infra + config + job |
| `make k8s-redeploy` | Rebuild image and redeploy everything |
| `make k8s-job` | Delete and rerun the research job |
| `make k8s-status` | Show all resources in deepforge namespace |
| `make k8s-logs` | Follow deepforge pod logs |
| `make port-forward` | Expose Mailhog (8025) and Grafana (3000) |
| `make k8s-clean` | Remove K8s resources, keep PVCs |
| `make k8s-nuke` | Full K8s wipe including PVCs and namespace |

---

## Infrastructure Components (Kubernetes)

The `deploy/k8s/infra/` directory contains production-ready manifests for every dependency:

**SearXNG** — Deployment with liveness/readiness probes, ConfigMap-mounted `settings.yml` (JSON format enabled, Google/Bing/DuckDuckGo engines enabled), ClusterIP Service, resource limits (256Mi/512Mi memory, 100m/500m CPU).

**Grafana Tempo** — Deployment with OTLP/HTTP receiver on port 4318 and query API on port 3200, ConfigMap-mounted configuration, PersistentVolumeClaim (10Gi) for trace storage, liveness/readiness probes, ClusterIP Service exposing both ports.

**Grafana** — Deployment with Tempo auto-provisioned as a datasource via ConfigMap, anonymous admin access for local use, PersistentVolumeClaim (5Gi), ClusterIP Service.

**Mailhog** — Deployment exposing SMTP (1025) and web UI (8025), ClusterIP Service with named ports.

All components run in the `deepforge` namespace. Resource requests and limits are defined on every container.

---

## Extending deepforge

**Adding an LLM provider** — implement the `Provider` interface (three methods) and wire it in `main.go`. No agent code changes required.

**Adding an email backend** — implement the `EmailSender` interface (one method) and add a case to the sender switch in `main.go`. No agent code changes required.

**Adding an agent** — create a struct in `internal/agents/`, inject `llm.Provider` and `*observability.Observability`, follow the span pattern. Add it to `ResearcherPipeline` in `pipeline/research.go`.

**Swapping the search backend** — `SearchAgent` calls `searchClient.Search(ctx, query) (string, error)`. Extract an interface from `SearXNGClient` and provide an alternative implementation.

