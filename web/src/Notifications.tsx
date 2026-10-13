import { useCallback, useEffect, useId, useState } from "react";
import type { FormEvent } from "react";
import { api } from "./api";
import type { Channel, Delivery } from "./api";

const when = (v: string | null) =>
  v ? new Date(v).toLocaleString("pt-BR") : "—";
const states: Record<string, string> = {
  pending: "Pendente",
  sent: "Aceita pelo destino",
  failed: "Falha definitiva",
  cancelled: "Cancelada",
};
export function Notifications({
  csrf,
  configured,
  enabled,
  revision,
  onError,
}: {
  csrf: string;
  configured: boolean;
  enabled: boolean;
  revision: number;
  onError: (e: unknown) => void;
}) {
  const formId = useId();
  const [kind, setKind] = useState("webhook");
  const [channels, setChannels] = useState<Channel[]>([]);
  const [deliveries, setDeliveries] = useState<Delivery[]>([]);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [reload, setReload] = useState(0);
  const [notice, setNotice] = useState("");
  const [confirm, setConfirm] = useState<{
    channel: Channel;
    action: "activate" | "pause" | "archive";
  } | null>(null);
  const refresh = useCallback(() => setReload((v) => v + 1), []);
  useEffect(() => {
    let active = true;
    setLoading(true);
    setFailed(false);
    Promise.all([
      api<{ items: Channel[] }>("/alert-channels"),
      api<{ items: Delivery[] }>(`/alert-deliveries?limit=20&offset=${offset}`),
    ])
      .then(([c, d]) => {
        if (active) {
          setChannels(c.items);
          setDeliveries(d.items);
        }
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
  }, [offset, reload, revision, onError]);
  async function perform(fn: () => Promise<void>) {
    setBusy(true);
    setNotice("");
    try {
      await fn();
      refresh();
    } catch (e) {
      onError(e);
    } finally {
      setBusy(false);
    }
  }
  async function create(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const element = e.currentTarget;
    const data = new FormData(element);
    const config =
      kind === "webhook"
        ? { url: data.get("url"), secret: data.get("secret") }
        : {
            host: data.get("host"),
            port: Number(data.get("port")),
            username: data.get("username"),
            password: data.get("password"),
            from: data.get("from"),
            to: data.get("to"),
          };
    await perform(async () => {
      await api(
        "/alert-channels",
        "POST",
        { name: data.get("name"), type: kind, config },
        csrf,
      );
      element.reset();
      setNotice("Canal salvo pausado. Nenhuma mensagem foi enviada.");
    });
  }
  const input = (name: string, label: string, type = "text", max = 254) => (
    <label htmlFor={`${formId}-${name}`} key={name}>
      {label}
      <input
        id={`${formId}-${name}`}
        name={name}
        type={type}
        required
        maxLength={max}
        autoComplete={type === "password" ? "new-password" : "off"}
      />
    </label>
  );
  return (
    <section className="panel" aria-labelledby="notifications-title">
      <div className="section-header">
        <h2 id="notifications-title">Canais e entregas</h2>
        <button disabled={loading || busy} onClick={refresh}>
          Atualizar entregas
        </button>
      </div>
      <p className="message">
        {enabled
          ? "Executor de entregas habilitado."
          : "Executor de entregas desabilitado. Pendências serão preservadas até a ativação."}{" "}
        Sucesso indica aceite do webhook ou servidor SMTP, não chegada à caixa
        de e-mail.
      </p>
      {!configured ? (
        <p>
          Cadastro de canais indisponível neste ambiente. Consulte a
          configuração na documentação.
        </p>
      ) : (
        <details>
          <summary>Cadastrar canal de notificação</summary>
          <label htmlFor={`${formId}-type`}>Tipo de canal</label>
          <select
            id={`${formId}-type`}
            value={kind}
            disabled={busy}
            onChange={(e) => setKind(e.target.value)}
          >
            <option value="webhook">Webhook HTTPS</option>
            <option value="smtp">E-mail SMTP</option>
          </select>
          <form key={kind} onSubmit={create} className="endpoint-form">
            {input("name", "Nome do canal", "text", 200)}
            {kind === "webhook" ? (
              <>
                {input("url", "URL HTTPS do webhook", "url", 2048)}
                <label htmlFor={`${formId}-secret`}>
                  Segredo de assinatura (mínimo 32 caracteres)
                  <input
                    id={`${formId}-secret`}
                    name="secret"
                    type="password"
                    autoComplete="new-password"
                    required
                    minLength={32}
                    maxLength={1024}
                  />
                </label>
              </>
            ) : (
              <>
                {input("host", "Servidor SMTP")}
                <div>
                  <label htmlFor={`${formId}-port`}>Porta SMTP</label>
                  <select id={`${formId}-port`} name="port" defaultValue="587">
                    <option value="587">587 — STARTTLS obrigatório</option>
                    <option value="465">465 — TLS implícito</option>
                  </select>
                </div>
                {input("username", "Usuário SMTP")}
                {input("password", "Senha SMTP", "password", 1024)}
                {input("from", "E-mail remetente", "email")}
                {input("to", "E-mail destinatário", "email")}
              </>
            )}
            <p className="hint">
              Use somente destinos e destinatários autorizados. Configuração
              fica cifrada e não é exibida novamente. Para trocar destino ou
              credenciais, crie um substituto e arquive o anterior.
            </p>
            <button
              className="primary"
              disabled={busy || channels.length >= 20}
            >
              Salvar canal pausado
            </button>
          </form>
        </details>
      )}
      {notice && <p role="status">{notice}</p>}
      {loading ? (
        <p role="status">Carregando canais e entregas…</p>
      ) : failed ? (
        <p>Canais e entregas indisponíveis. Tente atualizar.</p>
      ) : (
        <>
          <h3>Canais cadastrados</h3>
          {channels.length === 0 ? (
            <p>Nenhum canal cadastrado.</p>
          ) : (
            channels.map((c) => (
              <article className="check" key={c.id}>
                <h4>{c.name}</h4>
                <p>
                  {c.type === "smtp" ? "E-mail SMTP" : "Webhook HTTPS"} ·{" "}
                  {c.enabled ? "Ativo" : "Pausado"}
                </p>
                <div className="actions">
                  <button
                    disabled={busy || (!c.enabled && !configured)}
                    onClick={() =>
                      setConfirm({
                        channel: c,
                        action: c.enabled ? "pause" : "activate",
                      })
                    }
                  >
                    {c.enabled ? "Pausar canal" : "Ativar canal"}
                  </button>
                  <button
                    disabled={busy}
                    onClick={() =>
                      setConfirm({ channel: c, action: "archive" })
                    }
                  >
                    Arquivar canal
                  </button>
                </div>
              </article>
            ))
          )}
          <h3>Histórico de entregas</h3>
          {deliveries.length === 0 ? (
            <p>Nenhuma entrega nesta página.</p>
          ) : (
            deliveries.map((d) => (
              <article className="check" key={d.id}>
                <h4>
                  {states[d.status] || d.status} ·{" "}
                  {d.channel_type === "smtp" ? "E-mail" : "Webhook"}
                </h4>
                <p className="hint">
                  Entrega {d.id} · Evento {d.alert_event_id}
                </p>
                <p>
                  {d.attempt_count} tentativa(s) · Última:{" "}
                  {when(d.last_attempt_at)} · Próxima: {when(d.next_attempt_at)}{" "}
                  · Aceita: {when(d.accepted_at)}
                </p>
                {d.last_error && <p>Código: {d.last_error}</p>}
              </article>
            ))
          )}
          <nav className="actions pagination" aria-label="Páginas de entregas">
            <button
              disabled={offset === 0 || busy}
              onClick={() => setOffset(Math.max(0, offset - 20))}
            >
              Entregas mais recentes
            </button>
            <span>Página {offset / 20 + 1}</span>
            <button
              disabled={deliveries.length < 20 || busy}
              onClick={() => setOffset(offset + 20)}
            >
              Entregas mais antigas
            </button>
          </nav>
        </>
      )}
      {confirm && (
        <section
          role="alertdialog"
          aria-modal="false"
          aria-labelledby="channel-confirm-title"
        >
          <h3 id="channel-confirm-title">
            {confirm.action === "activate"
              ? "Ativar"
              : confirm.action === "pause"
                ? "Pausar"
                : "Arquivar"}{" "}
            {confirm.channel.name}?
          </h3>
          <p>
            {confirm.action === "activate"
              ? "Novos eventos serão enfileirados para este canal. Se o executor estiver habilitado, poderão ser enviados automaticamente."
              : "As pendências deste canal serão canceladas. Um envio já em andamento precisa terminar antes da alteração."}
          </p>
          <div className="actions">
            <button
              disabled={busy}
              onClick={() =>
                perform(async () => {
                  await api(
                    `/alert-channels/${confirm.channel.id}`,
                    confirm.action === "archive" ? "DELETE" : "PATCH",
                    confirm.action === "archive"
                      ? undefined
                      : { enabled: confirm.action === "activate" },
                    csrf,
                  );
                  setConfirm(null);
                  setNotice("Canal atualizado.");
                })
              }
            >
              Confirmar{" "}
              {confirm.action === "activate"
                ? "ativação"
                : confirm.action === "pause"
                  ? "pausa"
                  : "arquivamento de canal"}
            </button>
            <button disabled={busy} onClick={() => setConfirm(null)}>
              Cancelar alteração de canal
            </button>
          </div>
        </section>
      )}
    </section>
  );
}
