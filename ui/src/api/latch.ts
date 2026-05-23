// Typed wrappers for /api/latch/*. Mirrors the latch slice of dock's
// /ui/src/api/dashboard.ts but consumes the shared http client from
// polar-ui-common so the plugin can ship its UI independently.

import { requestJson } from "@networkextension/polar-ui-common/api/http";
import type {
  LatchProfileListResponse,
  LatchProxyListResponse,
  LatchRuleListResponse,
  LatchServiceNodeListResponse,
} from "../types/latch.js";

// ---------------------------------------------------------------------------
// Latch — Proxies
// ---------------------------------------------------------------------------

export async function fetchLatchProxies() {
  return requestJson<LatchProxyListResponse>("/api/latch/proxies");
}

export async function createLatchProxy(payload: { name: string; type: string; config: unknown }) {
  return requestJson<LatchProxyListResponse>("/api/latch/proxies", { method: "POST", body: payload });
}

export async function updateLatchProxy(groupId: string, payload: { name: string; type: string; config: unknown }) {
  return requestJson<LatchProxyListResponse>(`/api/latch/proxies/${encodeURIComponent(groupId)}`, { method: "PUT", body: payload });
}

export async function removeLatchProxy(groupId: string) {
  return requestJson<LatchProxyListResponse>(`/api/latch/proxies/${encodeURIComponent(groupId)}`, { method: "DELETE" });
}

export async function fetchLatchProxyVersions(groupId: string) {
  return requestJson<LatchProxyListResponse>(`/api/latch/proxies/${encodeURIComponent(groupId)}/versions`);
}

export async function rollbackLatchProxy(groupId: string, version: number) {
  return requestJson<LatchProxyListResponse>(`/api/latch/proxies/${encodeURIComponent(groupId)}/rollback/${version}`, { method: "PUT" });
}

// ---------------------------------------------------------------------------
// Latch — Rules
// ---------------------------------------------------------------------------

export async function fetchLatchRules() {
  return requestJson<LatchRuleListResponse>("/api/latch/rules");
}

export async function createLatchRule(payload: { name: string; content: string }) {
  return requestJson<LatchRuleListResponse>("/api/latch/rules", { method: "POST", body: payload });
}

export async function createLatchRuleFromFile(formData: FormData) {
  return requestJson<LatchRuleListResponse>("/api/latch/rules/upload", { method: "POST", body: formData });
}

export async function updateLatchRule(groupId: string, payload: { name: string; content: string }) {
  return requestJson<LatchRuleListResponse>(`/api/latch/rules/${encodeURIComponent(groupId)}`, { method: "PUT", body: payload });
}

export async function uploadLatchRuleFile(groupId: string, formData: FormData) {
  return requestJson<LatchRuleListResponse>(`/api/latch/rules/${encodeURIComponent(groupId)}/upload`, { method: "POST", body: formData });
}

export async function removeLatchRule(groupId: string) {
  return requestJson<LatchRuleListResponse>(`/api/latch/rules/${encodeURIComponent(groupId)}`, { method: "DELETE" });
}

export async function fetchLatchRuleVersions(groupId: string) {
  return requestJson<LatchRuleListResponse>(`/api/latch/rules/${encodeURIComponent(groupId)}/versions`);
}

export async function rollbackLatchRule(groupId: string, version: number) {
  return requestJson<LatchRuleListResponse>(`/api/latch/rules/${encodeURIComponent(groupId)}/rollback/${version}`, { method: "PUT" });
}

// ---------------------------------------------------------------------------
// Latch — Profiles
// ---------------------------------------------------------------------------

export async function fetchLatchAdminProfiles() {
  return requestJson<LatchProfileListResponse>("/api/latch/admin/profiles");
}

export async function createLatchProfile(payload: { name: string; description: string; proxy_group_ids: string[]; rule_group_id: string; enabled: boolean; shareable: boolean }) {
  return requestJson<LatchProfileListResponse>("/api/latch/admin/profiles", { method: "POST", body: payload });
}

export async function updateLatchProfile(id: string, payload: { name: string; description: string; proxy_group_ids: string[]; rule_group_id: string; enabled: boolean; shareable: boolean }) {
  return requestJson<LatchProfileListResponse>(`/api/latch/admin/profiles/${encodeURIComponent(id)}`, { method: "PUT", body: payload });
}

export async function removeLatchProfile(id: string) {
  return requestJson<LatchProfileListResponse>(`/api/latch/admin/profiles/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export async function fetchLatchProfiles() {
  return requestJson<LatchProfileListResponse>("/api/latch/profiles");
}

// ---------------------------------------------------------------------------
// Latch — Service Nodes
// ---------------------------------------------------------------------------

export async function fetchLatchServiceNodes(includeDeleted = false) {
  return requestJson<LatchServiceNodeListResponse>(`/api/latch/admin/service-nodes?include_deleted=${includeDeleted ? "true" : "false"}`);
}

export async function createLatchServiceNode(payload: {
  name: string;
  ip: string;
  port: number;
  proxy_type: string;
  config: unknown;
  status: string;
}) {
  return requestJson<LatchServiceNodeListResponse>("/api/latch/admin/service-nodes", { method: "POST", body: payload });
}

export async function updateLatchServiceNode(id: string, payload: {
  name: string;
  ip: string;
  port: number;
  proxy_type: string;
  config: unknown;
  status: string;
}) {
  return requestJson<LatchServiceNodeListResponse>(`/api/latch/admin/service-nodes/${encodeURIComponent(id)}`, { method: "PUT", body: payload });
}

export async function removeLatchServiceNode(id: string) {
  return requestJson<LatchServiceNodeListResponse>(`/api/latch/admin/service-nodes/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export async function issueLatchServiceNodeAgentToken(id: string) {
  return requestJson<LatchServiceNodeListResponse>(`/api/latch/admin/service-nodes/${encodeURIComponent(id)}/agent-token`, { method: "POST" });
}
