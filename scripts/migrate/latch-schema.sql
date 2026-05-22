-- ============================================================
-- polar_latch schema — end-state.
--
-- Apply:
--   CREATE DATABASE polar_latch OWNER ideamesh;
--   psql -d polar_latch -f scripts/migrate/latch-schema.sql
--
-- Tables are workspace-agnostic (single shared proxy / rule / profile
-- catalog per deployment). Agent-issued tokens authenticate the
-- lightweight node agents that report heartbeats.
-- ============================================================

CREATE TABLE IF NOT EXISTS latch_proxies (
    id         TEXT NOT NULL,
    group_id   TEXT NOT NULL,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL,
    config     JSONB NOT NULL DEFAULT '{}'::jsonb,
    sha1       TEXT NOT NULL DEFAULT '',
    version    INT  NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_latch_proxies_group_version
    ON latch_proxies(group_id, version DESC);

CREATE TABLE IF NOT EXISTS latch_service_nodes (
    id              TEXT NOT NULL PRIMARY KEY,
    name            TEXT NOT NULL,
    ip              TEXT NOT NULL,
    port            INT NOT NULL,
    proxy_type      TEXT NOT NULL,
    config          JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          TEXT NOT NULL DEFAULT 'unknown',
    last_updated_at TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL,
    is_deleted      BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_latch_service_nodes_active_updated
    ON latch_service_nodes(is_deleted, updated_at DESC);

CREATE TABLE IF NOT EXISTS latch_service_node_agent_tokens (
    id           TEXT NOT NULL PRIMARY KEY,
    node_id      TEXT NOT NULL REFERENCES latch_service_nodes(id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL UNIQUE,
    created_by   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ,
    revoked      BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_latch_service_node_agent_tokens_node_active
    ON latch_service_node_agent_tokens(node_id, revoked, created_at DESC);

CREATE TABLE IF NOT EXISTS latch_service_node_heartbeats (
    id              TEXT NOT NULL PRIMARY KEY,
    node_id         TEXT NOT NULL REFERENCES latch_service_nodes(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'unknown',
    connected_peers INT NOT NULL DEFAULT 0,
    rx_bytes        BIGINT NOT NULL DEFAULT 0,
    tx_bytes        BIGINT NOT NULL DEFAULT 0,
    agent_version   TEXT NOT NULL DEFAULT '',
    hostname        TEXT NOT NULL DEFAULT '',
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    reported_at     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_latch_service_node_heartbeats_node_time
    ON latch_service_node_heartbeats(node_id, reported_at DESC);

CREATE TABLE IF NOT EXISTS latch_rules (
    id         TEXT NOT NULL,
    group_id   TEXT NOT NULL,
    name       TEXT NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    sha1       TEXT NOT NULL DEFAULT '',
    version    INT  NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_latch_rules_group_version
    ON latch_rules(group_id, version DESC);

CREATE TABLE IF NOT EXISTS latch_profiles (
    id              TEXT NOT NULL PRIMARY KEY,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    proxy_group_ids TEXT[] NOT NULL DEFAULT '{}',
    rule_group_id   TEXT,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    shareable       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_latch_profiles_enabled_shareable
    ON latch_profiles(enabled, shareable, created_at DESC);
