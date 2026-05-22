# polar-latch

Proxy / rule / profile management plugin for the [Polar](https://github.com/networkextension/Polar) platform.

Owns `latch_proxies`, `latch_rules`, `latch_profiles`, plus `latch_service_nodes` and the lightweight node-agent runtime (`/api/latch/agent/register` + `/heartbeat`). Latch profiles compose 0-N proxies (ss / ss3 / kcp_over_* / wireguard) with an optional rule file, with SHA1-tracked versioning and per-group rollback.

## Status

W2 handler migration combined with extraction at 2026-05-22. Tables are workspace-agnostic (single shared catalog per deployment); no LLM coupling.

## Install

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /tmp/latch-svc ./cmd/latch-svc
rsync -avz /tmp/latch-svc local@<deploy-box>:/Users/local/.local/bin/
```

Environment:
- `POLAR_DOCK_URL` (default `http://127.0.0.1:8080`)
- `POLAR_PLUGIN_TOKEN` (from `/admin-plugins.html`)
- `POLAR_LATCH_DB_DSN` (Postgres for `polar_latch`)
- `POLAR_LATCH_LISTEN` (default `127.0.0.1:8098`)
- `POLAR_LATCH_BLOB_DIR` (reserved for future blob storage; unused in W2)
- `POLAR_LATCH_METRICS_TOKEN` (optional — enables `/metrics`)

## Endpoints

Admin (Bearer + role=admin):

- Proxies: `GET / POST /api/latch/proxies`, `GET / PUT / DELETE /api/latch/proxies/:group_id`, `GET /api/latch/proxies/:group_id/versions`, `PUT /api/latch/proxies/:group_id/rollback/:version`
- Rules: `GET / POST /api/latch/rules`, `POST /api/latch/rules/upload`, `GET / PUT / DELETE /api/latch/rules/:group_id`, `GET /api/latch/rules/:group_id/{content,versions}`, `POST /api/latch/rules/:group_id/upload`, `PUT /api/latch/rules/:group_id/rollback/:version`
- Profiles: `GET / POST /api/latch/admin/profiles`, `GET / PUT / DELETE /api/latch/admin/profiles/:id`
- Service nodes: `GET / POST /api/latch/admin/service-nodes`, `PUT / DELETE /api/latch/admin/service-nodes/:id`, `POST /api/latch/admin/service-nodes/:id/agent-token`

User (Bearer):

- `GET /api/latch/profiles` — enabled+shareable profiles with resolved proxies + rules

Agent (Bearer = agent token from `latch_service_node_agent_tokens`):

- `POST /api/latch/agent/register`
- `POST /api/latch/agent/heartbeat`

## Schema

`scripts/migrate/latch-schema.sql` — six tables. Apply against a fresh `polar_latch` DB.

## Data migration from monorepo dock

```bash
SRC_DSN=postgres://ideamesh:test123456@127.0.0.1:5432/ideamesh \
DST_DSN=postgres://ideamesh:test123456@127.0.0.1:5432/polar_latch \
  ./scripts/migrate/latch-data.sh           # dry-run
  ./scripts/migrate/latch-data.sh --apply   # actually copy
```

## Related

- [Polar dock](https://github.com/networkextension/Polar)
- [polar-sdk](https://github.com/networkextension/polar-sdk)

## License

MIT
