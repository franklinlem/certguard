import { useCallback, useEffect, useId, useState } from "react";
import { api } from "./api";
import type { AlertPage } from "./api";
import { CertificateDetails } from "./Dashboard";

const names: Record<string, string> = {
  certificate_expiry_warning: "Certificado próximo do vencimento",
  certificate_expiry_critical: "Certificado em faixa crítica",
  certificate_expiry_reminder: "Lembrete: até três dias para vencer",
  certificate_expired: "Certificado expirado",
  certificate_not_yet_valid: "Certificado ainda não válido",
  hostname_mismatch: "Hostname incompatível",
  untrusted_chain: "Cadeia não confiável",
  blocked_target: "Destino bloqueado pela proteção SSRF",
  dns_error: "Falha de DNS",
  connection_error: "Falha de conexão",
  connection_timeout: "Tempo de conexão esgotado",
  connection_refused: "Conexão recusada",
  tls_error: "Falha no handshake TLS",
  internal_error: "Falha interna de inspeção",
  invalid_validation: "Validação incompleta",
  invalid_validity: "Validade não avaliada",
  invalid_result: "Resultado de inspeção inconsistente",
  endpoint_recovered: "Endpoint recuperado",
};
const reasons: Record<string, string> = {
  recovered: "Recuperação confirmada",
  changed: "Condição ou certificado alterado",
  target_changed: "Destino alterado",
  archived: "Endpoint arquivado",
};
const severity: Record<string, string> = {
  critical: "!! Crítico",
  warning: "! Atenção",
  info: "ⓘ Informativo",
};
const when = (value: string) => new Date(value).toLocaleString("pt-BR");

export function Alerts({
  revision,
  onError,
}: {
  revision: number;
  onError: (e: unknown) => void;
}) {
  const filterId = useId();
  const [state, setState] = useState("active");
  const [offset, setOffset] = useState(0);
  const [page, setPage] = useState<AlertPage | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [refresh, setRefresh] = useState(0);
  const reload = useCallback(() => setRefresh((n) => n + 1), []);
  useEffect(() => {
    let active = true;
    setLoading(true);
    setFailed(false);
    setPage(null);
    api<AlertPage>(`/alerts?state=${state}&limit=20&offset=${offset}`)
      .then((data) => {
        if (active) setPage(data);
      })
      .catch((e) => {
        if (active) {
          setFailed(true);
          onError(e);
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [state, offset, revision, refresh, onError]);
  return (
    <section className="panel" aria-labelledby="alerts-title">
      <div className="section-header">
        <h2 id="alerts-title">Alertas e eventos</h2>
        <button disabled={loading} onClick={reload}>
          Atualizar alertas
        </button>
      </div>
      <p className="hint">
        Eventos gerados por novas verificações. Entrega por e-mail e webhook
        ainda não implementada. Pausar o monitoramento não confirma recuperação.
      </p>
      <label htmlFor={filterId}>Exibir eventos</label>
      <select
        id={filterId}
        value={state}
        disabled={loading}
        onChange={(e) => {
          setOffset(0);
          setState(e.target.value);
        }}
      >
        <option value="active">Problemas ativos</option>
        <option value="all">Todos os eventos</option>
        <option value="resolved">Resolvidos e recuperações</option>
      </select>
      {loading ? (
        <p role="status">Carregando alertas…</p>
      ) : failed ? (
        <p>Alertas indisponíveis. Tente atualizar novamente.</p>
      ) : (
        page && (
          <>
            <p className="hint" aria-label="Resumo dos alertas">
              {page.summary.active} ativos · {page.summary.critical} críticos
              ativos · {page.summary.resolved} problemas resolvidos ·{" "}
              {page.summary.info} recuperações. Atualizado em{" "}
              {when(page.generated_at)}.
            </p>
            {page.items.length === 0 ? (
              <p>Nenhum evento neste filtro e página.</p>
            ) : (
              page.items.map((event) => (
                <article className="check" key={event.id}>
                  <span className={`badge ${event.severity}`}>
                    {severity[event.severity] || event.severity}
                  </span>
                  <h3>{names[event.event_type] || event.event_type}</h3>
                  <p>
                    {event.endpoint_name} · {event.hostname}:{event.port}
                  </p>
                  <p className="hint">
                    Registrado em {when(event.created_at)} ·{" "}
                    {event.resolved_at
                      ? `Resolvido em ${when(event.resolved_at)} — ${reasons[event.resolution!] || event.resolution}`
                      : "Problema ativo no último resultado observado"}
                  </p>
                  {event.threshold_days > 0 && (
                    <p className="hint">
                      Limiar desta regra: {event.threshold_days} dias.
                    </p>
                  )}
                  <details>
                    <summary>Ver observação que originou o evento</summary>
                    <CertificateDetails
                      result={event.snapshot}
                      title="Observação preservada no evento"
                    />
                    <p className="hint">
                      Referência ao check:{" "}
                      {event.check_id ||
                        "Removido pela retenção; observação preservada"}
                    </p>
                  </details>
                </article>
              ))
            )}
            <nav className="actions pagination" aria-label="Páginas de alertas">
              <button
                disabled={offset === 0}
                onClick={() => setOffset(Math.max(0, offset - 20))}
              >
                Eventos mais recentes
              </button>
              <span>Página {offset / 20 + 1}</span>
              <button
                disabled={page.items.length < 20}
                onClick={() => setOffset(offset + 20)}
              >
                Eventos mais antigos
              </button>
            </nav>
          </>
        )
      )}
    </section>
  );
}
