CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE plans (
  id UUID PRIMARY KEY,
  request_id UUID NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  current_version INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE plan_versions (
  id BIGSERIAL PRIMARY KEY,
  plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  version INTEGER NOT NULL,
  response_json JSONB NOT NULL,
  context_hash TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(plan_id, version)
);

CREATE TABLE plan_requests (
  request_id UUID PRIMARY KEY,
  payload_hash TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('processing','completed','failed')),
  response_json JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + interval '10 minutes'
);
CREATE INDEX idx_plan_requests_expires ON plan_requests(expires_at);

CREATE TABLE model_registry (
  model_id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  task_classes JSONB NOT NULL DEFAULT '[]'::jsonb,
  supported_params JSONB NOT NULL DEFAULT '[]'::jsonb,
  evidence TEXT NOT NULL DEFAULT '',
  verified_at TIMESTAMPTZ,
  enabled BOOLEAN NOT NULL DEFAULT false,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pilot_events (
  id BIGSERIAL PRIMARY KEY,
  experiment_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  plan_version INTEGER,
  task_id TEXT,
  catalog_source TEXT,
  policy_version TEXT,
  latency_ms BIGINT,
  provider_calls INTEGER,
  usage JSONB,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
