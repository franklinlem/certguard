import { useId } from "react";
import type { Check, Summary } from "./api";

const when = (value?: string | null) =>
  value ? new Date(value).toLocaleString("pt-BR") : "Não disponível";
const validation = (value?: boolean, good = "Válido", bad = "Inválido") =>
  value === undefined ? "Não avaliado" : value ? good : bad;

export function SummaryCards({
  summary,
  generatedAt,
}: {
  summary: Summary;
  generatedAt: string;
}) {
  const cards: [keyof Summary, string][] = [
    ["total", "Endpoints"],
    ["enabled", "Ativos"],
    ["paused", "Pausados"],
    ["healthy", "✓ Saudáveis"],
    ["warning", "! Atenção"],
    ["critical", "!! Críticos"],
    ["errors", "× Erros"],
    ["unchecked", "Sem resultado"],
    ["expiring_30_days", "Vencem em até 30 dias"],
  ];
  return (
    <section aria-label="Visão geral" className="overview">
      <div className="summary-grid">
        {cards.map(([key, label]) => (
          <div className={`summary-card ${key}`} key={key}>
            <span>{label}</span>
            <strong>{summary[key]}</strong>
          </div>
        ))}
      </div>
      <p className="hint">
        Todos os endpoints não arquivados, incluindo pausados. Status do último
        check; não representa uma verificação em tempo real. Atualizado em{" "}
        {when(generatedAt)}.
      </p>
    </section>
  );
}

export function CertificateDetails({
  result,
}: {
  result?: Check["result"] | null;
}) {
  if (!result) return <p>Ainda não há resultado para o destino atual.</p>;
  const cert = result.certificate;
  return (
    <section aria-label="Detalhes do último resultado">
      <h3>Último resultado do destino atual</h3>
      <dl>
        <dt>Destino observado</dt>
        <dd>
          {result.hostname}:{result.port ?? "Não disponível"}
        </dd>
        <dt>Verificado em</dt>
        <dd>{when(result.checked_at)}</dd>
        <dt>IP resolvido</dt>
        <dd>{result.resolved_ip || "Não disponível"}</dd>
        <dt>Protocolo / cifra</dt>
        <dd>
          {result.tls_version || "Não disponível"} /{" "}
          {result.cipher_suite || "Não disponível"}
        </dd>
        <dt>Hostname / cadeia</dt>
        <dd>
          {validation(result.hostname_valid)} /{" "}
          {validation(result.chain_valid, "Confiável", "Não confiável")}
        </dd>
        {cert && (
          <>
            <dt>Emissor</dt>
            <dd>{cert.issuer || "Não disponível"}</dd>
            <dt>Titular (subject)</dt>
            <dd>{cert.subject || "Não disponível"}</dd>
            <dt>Número de série</dt>
            <dd>{cert.serial_number || "Não disponível"}</dd>
            <dt>Início da validade</dt>
            <dd>{when(cert.not_before)}</dd>
            <dt>Fim da validade</dt>
            <dd>{when(cert.not_after)}</dd>
            <dt>Dias restantes no check</dt>
            <dd>{result.days_remaining ?? "Não disponível"}</dd>
            <dt>SANs DNS</dt>
            <dd>{cert.dns_names?.join(", ") || "Nenhum"}</dd>
            <dt>SANs IP</dt>
            <dd>{cert.ip_addresses?.join(", ") || "Nenhum"}</dd>
            <dt>SANs e-mail</dt>
            <dd>{cert.email_addresses?.join(", ") || "Nenhum"}</dd>
            <dt>SANs URI</dt>
            <dd>{cert.uris?.join(", ") || "Nenhum"}</dd>
            <dt>SHA-256</dt>
            <dd className="fingerprint">{cert.fingerprint_sha256}</dd>
          </>
        )}
      </dl>
      {!cert && <p>Esta inspeção não retornou certificado.</p>}
    </section>
  );
}

export function HistoryChart({ checks }: { checks: Check[] }) {
  const title = useId();
  const rows = [...checks].sort(
    (a, b) =>
      Date.parse(a.result.checked_at) - Date.parse(b.result.checked_at) ||
      a.id.localeCompare(b.id),
  );
  const valid = (c: Check) =>
    Number.isFinite(c.result.days_remaining) &&
    Number.isFinite(Date.parse(c.result.checked_at));
  const points = rows.filter(valid);
  const times = rows
    .map((c) => Date.parse(c.result.checked_at))
    .filter(Number.isFinite);
  const minT = Math.min(...times),
    maxT = Math.max(...times);
  const low = Math.min(0, ...points.map((c) => c.result.days_remaining!));
  const high = Math.max(
    low + 1,
    ...points.map((c) => c.result.days_remaining!),
  );
  const x = (c: Check) =>
    maxT === minT
      ? 335
      : 55 + ((Date.parse(c.result.checked_at) - minT) / (maxT - minT)) * 555;
  const y = (c: Check) =>
    195 - ((c.result.days_remaining! - low) / (high - low)) * 155;
  return (
    <section className="history-chart" aria-label="Evolução da validade">
      <h3>Dias restantes no momento de cada check</h3>
      <p className="hint">
        Até 20 amostras desta página do histórico. Falhas sem validade e
        mudanças de destino interrompem a linha. Consulte as datas; estes
        valores não são uma contagem atual.
      </p>
      {points.length === 0 ? (
        <p>Sem dados de validade nesta página para desenhar o gráfico.</p>
      ) : (
        <svg
          viewBox="0 0 640 250"
          width="640"
          height="250"
          role="img"
          aria-labelledby={title}
        >
          <title id={title}>
            Evolução dos dias restantes, {points.length} amostras com validade.
            Dados na tabela abaixo.
          </title>
          <line x1="55" y1="40" x2="55" y2="195" className="chart-axis" />
          <line x1="55" y1="195" x2="610" y2="195" className="chart-axis" />
          <text x="5" y="45">
            {high} d
          </text>
          <text x="5" y="197">
            {low} d
          </text>
          <text x="55" y="225">
            {new Date(minT).toLocaleDateString("pt-BR")}
          </text>
          <text x="610" y="245" textAnchor="end">
            {new Date(maxT).toLocaleDateString("pt-BR")}
          </text>
          {rows.map((c, i) => {
            if (!valid(c)) return null;
            const prev = rows[i - 1];
            const connect =
              prev &&
              valid(prev) &&
              prev.result.hostname === c.result.hostname &&
              prev.result.port === c.result.port;
            return (
              <g key={c.id}>
                {connect && (
                  <line
                    data-testid="history-segment"
                    x1={x(prev)}
                    y1={y(prev)}
                    x2={x(c)}
                    y2={y(c)}
                    className="chart-line"
                  />
                )}
                <circle cx={x(c)} cy={y(c)} r="4" className="chart-point">
                  <title>
                    {when(c.result.checked_at)}: {c.result.days_remaining} dias
                    · {c.result.hostname}
                  </title>
                </circle>
              </g>
            );
          })}
        </svg>
      )}
      <details>
        <summary>Consultar dados do gráfico em tabela</summary>
        <div className="table-scroll">
          <table>
            <caption>Amostras da página em ordem cronológica</caption>
            <thead>
              <tr>
                <th>Data do check</th>
                <th>Destino</th>
                <th>Dias restantes</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((c) => (
                <tr key={c.id}>
                  <td>{when(c.result.checked_at)}</td>
                  <td>
                    {c.result.hostname}:{c.result.port ?? "—"}
                  </td>
                  <td>{c.result.days_remaining ?? "Não disponível"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </details>
    </section>
  );
}
