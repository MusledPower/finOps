# Task Decomposer Backend (Go + PostgreSQL)

Backend starter implementing the MVP HTTP boundary from the supplied specification.

## Requirements
- Go 1.24+
- PostgreSQL 16+
- golang-migrate CLI (for migrations)

## Run locally
1. Copy `.env.example` to `.env` and configure `API_BEARER_TOKEN`, `LLM_API_KEY`, and `LLM_MODEL`.
2. Start PostgreSQL: `docker compose up -d postgres`
3. Export environment variables from `.env` (or use your IDE run configuration).
4. Apply schema: `make migrate-up`
5. Start API: `go run ./cmd/api`

The API uses an OpenAI-compatible Chat Completions endpoint. Configure `LLM_BASE_URL` for another compatible provider. The provider API key is server-side only.

## Endpoints
- `GET /health` (public)
- `GET /v1/capabilities` (Bearer auth)
- `POST /v1/plans` (Bearer auth)
- `GET /v1/plans/{id}` (Bearer auth)

## Notes / remaining MVP work
This is an executable foundation, not yet a complete production implementation of every acceptance criterion in the specification. In particular:
- request idempotency currently persists completed responses but does not yet short-circuit an in-flight duplicate or reject same requestId with a different payload hash;
- `/v1/recommendations`, structured retry/repair call, cancellation propagation to all provider implementations, registry-backed policy enforcement, and pilot event ingestion are not implemented yet;
- provider structured JSON mode depends on provider compatibility;
- database migrations are applied separately by the deployment process;
- auth currently uses one configured pilot bearer token; use a proper token store/rotation mechanism for production.

Never log request bodies, prompts, source code, or provider credentials.
