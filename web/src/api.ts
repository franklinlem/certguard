export type Session = {
  user: { id: string; email: string };
  csrf_token: string;
  background: {
    scheduler_enabled: boolean;
    check_workers: number;
    retention_enabled: boolean;
    retention_days: number;
  };
};
export type Endpoint = {
  id: string;
  name: string;
  hostname: string;
  port: number;
  enabled: boolean;
  interval_minutes: number;
  warning_days: number;
  critical_days: number;
  last_check_at: string | null;
  next_check_at: string | null;
  latest?: Check["result"] | null;
};
export type Check = {
  id: string;
  result: {
    hostname: string;
    port?: number;
    resolved_ip?: string;
    tls_version?: string;
    cipher_suite?: string;
    duration_ms?: number;
    status: string;
    code: string;
    checked_at: string;
    days_remaining?: number;
    hostname_valid?: boolean;
    chain_valid?: boolean;
    certificate?: {
      issuer: string;
      subject?: string;
      serial_number?: string;
      dns_names?: string[];
      ip_addresses?: string[];
      email_addresses?: string[];
      uris?: string[];
      not_before?: string;
      fingerprint_sha256: string;
      not_after: string;
    };
  };
};
export class APIError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}
export async function api<T>(
  path: string,
  method = "GET",
  body?: unknown,
  csrf?: string,
): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    method,
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      ...(csrf ? { "X-CSRF-Token": csrf } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new APIError(
      res.status,
      data.error || "Não foi possível concluir a operação.",
    );
  }
  return res.status === 204 ? (undefined as T) : res.json();
}

export type Summary = {
  total: number;
  enabled: number;
  paused: number;
  healthy: number;
  warning: number;
  critical: number;
  errors: number;
  unchecked: number;
  expiring_30_days: number;
};
export type Dashboard = {
  summary: Summary;
  items: Endpoint[];
  generated_at: string;
  limit: number;
  offset: number;
};

export type AlertEvent = {
  id: string;
  endpoint_id: string;
  check_id: string | null;
  endpoint_name: string;
  hostname: string;
  port: number;
  event_type: string;
  severity: string;
  threshold_days: number;
  snapshot: Check["result"];
  created_at: string;
  resolved_at: string | null;
  resolution: string | null;
};
export type AlertPage = {
  items: AlertEvent[];
  summary: { active: number; critical: number; resolved: number; info: number };
  limit: number;
  offset: number;
  generated_at: string;
};
