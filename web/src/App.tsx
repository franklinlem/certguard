import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api, APIError } from "./api";
import type { Check, Endpoint, Session } from "./api";

const date = (value: string | null) =>
  value ? new Date(value).toLocaleString("pt-BR") : "Ainda não verificado";
const codes: Record<string, string> = {
  success: "Certificado verificado",
  blocked_target: "Destino bloqueado pela proteção SSRF",
  dns_error: "Falha de DNS",
  connection_error: "Falha de conexão",
  connection_timeout: "Tempo esgotado",
  connection_refused: "Conexão recusada",
  tls_error: "Falha no handshake TLS",
  certificate_expired: "Certificado expirado",
  certificate_not_yet_valid: "Certificado ainda não válido",
  hostname_mismatch: "Hostname incompatível",
  untrusted_chain: "Cadeia não confiável",
  internal_error: "Falha interna",
};
const statuses: Record<string, string> = {
  healthy: "✓ Saudável",
  warning: "! Atenção",
  critical: "!! Crítico",
  error: "× Erro",
};

export function App() {
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [items, setItems] = useState<Endpoint[]>([]);
  const [offset, setOffset] = useState(0);
  const [editing, setEditing] = useState<Endpoint | "new" | null>(null);
  const [selected, setSelected] = useState<Endpoint | null>(null);
  const [history, setHistory] = useState<Check[]>([]);
  const [historyOffset, setHistoryOffset] = useState(0);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [archive, setArchive] = useState<Endpoint | null>(null);
  const [notice, setNotice] = useState("");
  const handleError = useCallback((e: unknown) => {
    if (e instanceof APIError && e.status === 401) {
      setSession(null);
      setItems([]);
      setSelected(null);
    }
    setError(
      e instanceof Error ? e.message : "Falha inesperada. Tente novamente.",
    );
  }, []);
  useEffect(() => {
    let active = true;
    api<Session>("/auth/me")
      .then((s) => {
        if (active) setSession(s);
      })
      .catch((e) => {
        if (active && (!(e instanceof APIError) || e.status !== 401))
          handleError(e);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [handleError]);
  const refresh = useCallback(async () => {
    const data = await api<{ items: Endpoint[] }>(
      `/endpoints?limit=20&offset=${offset}`,
    );
    setItems(data.items);
  }, [offset]);
  useEffect(() => {
    if (!session) return;
    let active = true;
    api<{ items: Endpoint[] }>(`/endpoints?limit=20&offset=${offset}`)
      .then((d) => {
        if (active) setItems(d.items);
      })
      .catch((e) => {
        if (active) handleError(e);
      });
    return () => {
      active = false;
    };
  }, [session, offset, handleError]);
  useEffect(() => {
    if (!session || !selected) return;
    let active = true;
    setHistoryLoading(true);
    setHistory([]);
    api<{ items: Check[] }>(
      `/endpoints/${selected.id}/checks?limit=20&offset=${historyOffset}`,
    )
      .then((d) => {
        if (active) setHistory(d.items);
      })
      .catch((e) => {
        if (active) handleError(e);
      })
      .finally(() => {
        if (active) setHistoryLoading(false);
      });
    return () => {
      active = false;
    };
  }, [session, selected, historyOffset, handleError]);
  async function act(fn: () => Promise<void>) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await fn();
    } catch (e) {
      handleError(e);
    } finally {
      setBusy(false);
    }
  }
  async function login(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await act(async () => {
      const s = await api<Session>("/auth/login", "POST", {
        email: form.get("email"),
        password: form.get("password"),
      });
      setSession(s);
      setOffset(0);
    });
  }
  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const data = {
      name: form.get("name"),
      hostname: form.get("hostname"),
      port: 443,
      warning_days: Number(form.get("warning_days")),
      critical_days: Number(form.get("critical_days")),
      interval_minutes: Number(form.get("interval_minutes")),
    };
    await act(async () => {
      await api(
        editing === "new"
          ? "/endpoints"
          : `/endpoints/${(editing as Endpoint).id}`,
        editing === "new" ? "POST" : "PATCH",
        data,
        session!.csrf_token,
      );
      setEditing(null);
      await refresh();
      setNotice("Endpoint salvo.");
    });
  }
  return (
    <main>
      <header>
        <a className="brand" href="/" aria-label="CertGuard, início">
          <span aria-hidden="true">◈</span> CertGuard
        </a>
        <div className="header-actions">
          <span className="phase">Verificações manuais</span>
          {session && (
            <button
              disabled={busy}
              onClick={() =>
                act(async () => {
                  await api(
                    "/auth/logout",
                    "POST",
                    undefined,
                    session.csrf_token,
                  );
                  setSession(null);
                  setItems([]);
                  setSelected(null);
                  setEditing(null);
                  setArchive(null);
                })
              }
            >
              Sair
            </button>
          )}
        </div>
      </header>
      {error && (
        <p role="alert" className="message error">
          {error}
        </p>
      )}
      {notice && (
        <p role="status" className="message">
          {notice}
        </p>
      )}
      {loading ? (
        <p role="status">Carregando sessão…</p>
      ) : !session ? (
        <section className="login" aria-labelledby="login-title">
          <p className="eyebrow">ACESSO ADMINISTRATIVO</p>
          <h1 id="login-title">
            Seus certificados.
            <br />
            Sob controle.
          </h1>
          <p>
            Entre para cadastrar endpoints e verificar seus certificados TLS.
          </p>
          <form onSubmit={login}>
            <label>
              E-mail
              <input
                name="email"
                type="email"
                autoComplete="username"
                required
                maxLength={254}
              />
            </label>
            <label>
              Senha
              <input
                name="password"
                type="password"
                autoComplete="current-password"
                required
                maxLength={72}
              />
            </label>
            <button className="primary" disabled={busy}>
              {busy ? "Entrando…" : "Entrar"}
            </button>
          </form>
          <p className="hint">
            Primeiro acesso? Crie o administrador com o comando local descrito
            na documentação. Não há cadastro público.
          </p>
        </section>
      ) : (
        <>
          <section className="page-title">
            <div>
              <p className="eyebrow">ÁREA DE MONITORAMENTO</p>
              <h1>Endpoints TLS</h1>
              <p className="hint">{session.user.email}</p>
            </div>
            <button
              className="primary"
              disabled={busy}
              onClick={() => setEditing("new")}
            >
              Adicionar endpoint
            </button>
          </section>
          <p className="message">
            As verificações são manuais nesta etapa. O agendamento e os alertas
            ainda não estão ativos.
          </p>
          {editing && (
            <section className="panel" aria-labelledby="form-title">
              <h2 id="form-title">
                {editing === "new" ? "Adicionar endpoint" : "Editar endpoint"}
              </h2>
              <form
                key={editing === "new" ? "new" : editing.id}
                onSubmit={save}
                className="endpoint-form"
              >
                <label>
                  Nome
                  <input
                    name="name"
                    required
                    maxLength={200}
                    defaultValue={editing === "new" ? "" : editing.name}
                  />
                </label>
                <label>
                  Hostname
                  <input
                    name="hostname"
                    required
                    placeholder="exemplo.com"
                    maxLength={253}
                    defaultValue={editing === "new" ? "" : editing.hostname}
                  />
                </label>
                <label>
                  Alerta de atenção (dias)
                  <input
                    name="warning_days"
                    type="number"
                    required
                    min={1}
                    max={36500}
                    defaultValue={editing === "new" ? 30 : editing.warning_days}
                  />
                </label>
                <label>
                  Alerta crítico (dias)
                  <input
                    name="critical_days"
                    type="number"
                    required
                    min={0}
                    max={36499}
                    defaultValue={editing === "new" ? 7 : editing.critical_days}
                  />
                </label>
                <label>
                  Intervalo futuro (minutos)
                  <input
                    name="interval_minutes"
                    type="number"
                    required
                    min={5}
                    max={10080}
                    defaultValue={
                      editing === "new" ? 360 : editing.interval_minutes
                    }
                  />
                </label>
                <p className="hint">
                  Apenas porta 443. Use somente domínios autorizados, sem URL ou
                  caminho. O DNS é validado em cada verificação.
                </p>
                <div className="actions">
                  <button className="primary" disabled={busy}>
                    Salvar
                  </button>
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => setEditing(null)}
                  >
                    Cancelar
                  </button>
                </div>
              </form>
            </section>
          )}
          <section aria-label="Endpoints cadastrados" className="endpoint-list">
            {items.length === 0 ? (
              <p>
                Nenhum endpoint nesta página. Cadastre um domínio para começar.
              </p>
            ) : (
              items.map((ep) => (
                <article className="endpoint" key={ep.id}>
                  <div>
                    <h2>{ep.name}</h2>
                    <p className="host">
                      {ep.hostname}:443{" "}
                      <span className="phase">
                        {ep.enabled ? "Ativo" : "Pausado"}
                      </span>
                    </p>
                    <p className="hint">
                      Última verificação: {date(ep.last_check_at)}
                    </p>
                  </div>
                  <div className="actions">
                    <button
                      className="primary"
                      disabled={busy || !ep.enabled}
                      onClick={() =>
                        act(async () => {
                          await api(
                            `/endpoints/${ep.id}/check`,
                            "POST",
                            undefined,
                            session.csrf_token,
                          );
                          await refresh();
                          setHistoryOffset(0);
                          setSelected({ ...ep });
                          setNotice(
                            "Verificação concluída. Consulte o resultado no histórico.",
                          );
                        })
                      }
                    >
                      {busy ? "Aguarde…" : "Verificar agora"}
                    </button>
                    <button
                      disabled={busy}
                      onClick={() => {
                        setSelected(ep);
                        setHistoryOffset(0);
                      }}
                    >
                      Histórico
                    </button>
                    <button disabled={busy} onClick={() => setEditing(ep)}>
                      Editar
                    </button>
                    <button
                      disabled={busy}
                      onClick={() =>
                        act(async () => {
                          await api(
                            `/endpoints/${ep.id}`,
                            "PATCH",
                            { enabled: !ep.enabled },
                            session.csrf_token,
                          );
                          await refresh();
                        })
                      }
                    >
                      {ep.enabled ? "Pausar" : "Ativar"}
                    </button>
                    <button disabled={busy} onClick={() => setArchive(ep)}>
                      Arquivar
                    </button>
                  </div>
                </article>
              ))
            )}
          </section>
          <nav className="actions pagination" aria-label="Páginas de endpoints">
            <button
              disabled={offset === 0 || busy}
              onClick={() => setOffset(Math.max(0, offset - 20))}
            >
              Anterior
            </button>
            <span>Página {offset / 20 + 1}</span>
            <button
              disabled={items.length < 20 || busy}
              onClick={() => setOffset(offset + 20)}
            >
              Próxima
            </button>
          </nav>
          {archive && (
            <section
              className="panel"
              role="alertdialog"
              aria-modal="false"
              aria-labelledby="archive-title"
            >
              <h2 id="archive-title">Arquivar {archive.name}?</h2>
              <p>
                O endpoint sairá da lista ativa. Seu histórico será preservado.
              </p>
              <div className="actions">
                <button
                  disabled={busy}
                  onClick={() =>
                    act(async () => {
                      await api(
                        `/endpoints/${archive.id}`,
                        "DELETE",
                        undefined,
                        session.csrf_token,
                      );
                      setArchive(null);
                      await refresh();
                      setNotice("Endpoint arquivado; histórico preservado.");
                    })
                  }
                >
                  Confirmar arquivamento
                </button>
                <button disabled={busy} onClick={() => setArchive(null)}>
                  Cancelar
                </button>
              </div>
            </section>
          )}
          {selected && (
            <section className="panel" aria-labelledby="history-title">
              <div className="section-header">
                <h2 id="history-title">Histórico · {selected.name}</h2>
                <button onClick={() => setSelected(null)}>
                  Fechar histórico
                </button>
              </div>
              {historyLoading ? (
                <p role="status">Carregando histórico…</p>
              ) : history.length === 0 ? (
                <p>Nenhuma verificação nesta página.</p>
              ) : (
                history.map((check) => (
                  <article className="check" key={check.id}>
                    <span className={`badge ${check.result.status}`}>
                      {statuses[check.result.status] || check.result.status}
                    </span>
                    <h3>{codes[check.result.code] || check.result.code}</h3>
                    <p className="hint">
                      {check.result.hostname} · {date(check.result.checked_at)}
                    </p>
                    {check.result.certificate && (
                      <dl>
                        <dt>Emissor</dt>
                        <dd>{check.result.certificate.issuer}</dd>
                        <dt>Validade</dt>
                        <dd>
                          {date(check.result.certificate.not_after)} ·{" "}
                          {check.result.days_remaining} dias restantes
                        </dd>
                        <dt>Hostname / cadeia</dt>
                        <dd>
                          {check.result.hostname_valid ? "Válido" : "Inválido"}{" "}
                          /{" "}
                          {check.result.chain_valid
                            ? "Confiável"
                            : "Não confiável"}
                        </dd>
                        <dt>SHA-256</dt>
                        <dd className="fingerprint">
                          {check.result.certificate.fingerprint_sha256}
                        </dd>
                      </dl>
                    )}
                  </article>
                ))
              )}
              <nav
                className="actions pagination"
                aria-label="Páginas do histórico"
              >
                <button
                  disabled={historyOffset === 0 || historyLoading}
                  onClick={() =>
                    setHistoryOffset(Math.max(0, historyOffset - 20))
                  }
                >
                  Mais recentes
                </button>
                <button
                  disabled={history.length < 20 || historyLoading}
                  onClick={() => setHistoryOffset(historyOffset + 20)}
                >
                  Mais antigas
                </button>
              </nav>
            </section>
          )}
        </>
      )}
      <footer>
        <span>CertGuard</span>
        <span>Segurança antes de conveniência.</span>
      </footer>
    </main>
  );
}
