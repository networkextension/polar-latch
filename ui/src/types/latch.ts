// Latch plugin type surface. Mirrors the relevant slice of dock's
// /ui/src/types/dashboard.ts so the plugin owns its own type defs
// once it's no longer bundled with the dock.

export type ErrorResponse = {
  error?: string;
  message?: string;
};

export type LatchProxyType =
  | "ss"
  | "ss3"
  | "kcp_over_http"
  | "kcp_over_ss"
  | "kcp_over_ss3"
  | "wireguard";

export type LatchProxy = {
  id: string;
  group_id: string;
  name: string;
  type: LatchProxyType;
  config: Record<string, unknown>;
  sha1: string;
  version: number;
  created_at: string;
};

export type LatchRule = {
  id: string;
  group_id: string;
  name: string;
  content: string;
  sha1: string;
  version: number;
  created_at: string;
};

export type LatchProfile = {
  id: string;
  name: string;
  description: string;
  proxy_group_ids: string[];
  rule_group_id: string;
  enabled: boolean;
  shareable: boolean;
  created_at: string;
  updated_at: string;
};

export type LatchServiceNode = {
  id: string;
  name: string;
  ip: string;
  port: number;
  proxy_type: LatchProxyType;
  config: Record<string, unknown>;
  status: string;
  last_updated_at: string;
  created_at: string;
  updated_at: string;
  is_deleted: boolean;
};

export type LatchProxyListResponse = ErrorResponse & {
  proxies?: LatchProxy[];
  proxy?: LatchProxy;
  versions?: LatchProxy[];
};

export type LatchRuleListResponse = ErrorResponse & {
  rules?: LatchRule[];
  rule?: LatchRule;
  versions?: LatchRule[];
};

export type LatchProfileDetail = LatchProfile & {
  proxies: LatchProxy[];
  rule?: LatchRule;
};

export type LatchProfileListResponse = ErrorResponse & {
  profiles?: LatchProfile[] | LatchProfileDetail[];
  profile?: LatchProfile;
};

export type LatchServiceNodeListResponse = ErrorResponse & {
  nodes?: LatchServiceNode[];
  node?: LatchServiceNode;
  token?: string;
  meta?: {
    id: string;
    node_id: string;
    created_by: string;
    created_at: string;
    revoked: boolean;
  };
};
