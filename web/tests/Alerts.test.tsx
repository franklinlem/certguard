import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { Alerts } from "../src/Alerts";
import type { AlertEvent } from "../src/api";
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
const event: AlertEvent = {
  id: "a",
  endpoint_id: "e",
  check_id: null,
  endpoint_name: "Portal",
  hostname: "example.com",
  port: 443,
  event_type: "certificate_expiry_warning",
  severity: "warning",
  threshold_days: 30,
  created_at: "2026-10-13T12:00:00Z",
  resolved_at: null,
  resolution: null,
  snapshot: {
    hostname: "example.com",
    port: 443,
    status: "warning",
    code: "success",
    checked_at: "2026-10-13T12:00:00Z",
    days_remaining: 20,
  },
};
const page = (items: AlertEvent[]) => ({
  items,
  summary: { active: 1, critical: 0, resolved: 0, info: 0 },
  generated_at: "2026-10-13T12:00:00Z",
});
test("lista problemas ativos, filtra eventos e preserva observação sem check", async () => {
  const mock = vi
    .fn()
    .mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify(page([event])))),
    );
  vi.stubGlobal("fetch", mock);
  render(<Alerts revision={0} onError={vi.fn()} />);
  expect(
    await screen.findByText("Certificado próximo do vencimento"),
  ).toBeTruthy();
  expect(screen.getByText(/Problema ativo no último resultado/)).toBeTruthy();
  fireEvent.click(screen.getByText("Ver observação que originou o evento"));
  expect(
    screen.getByText(/Removido pela retenção; observação preservada/),
  ).toBeTruthy();
  fireEvent.change(
    screen.getByRole("combobox", { name: "Exibir eventos" }),
    {
      target: { value: "all" },
    },
  );
  await waitFor(() =>
    expect(mock).toHaveBeenCalledWith(
      "/api/v1/alerts?state=all&limit=20&offset=0",
      expect.any(Object),
    ),
  );
  expect(
    screen.getByText(/Entrega por e-mail e webhook ainda não implementada/),
  ).toBeTruthy();
});
test("erro não aparece como lista vazia e permite nova tentativa", async () => {
  const onError = vi.fn();
  const mock = vi
    .fn()
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ error: "Indisponível" }), { status: 503 }),
    )
    .mockResolvedValue(new Response(JSON.stringify(page([]))));
  vi.stubGlobal("fetch", mock);
  render(<Alerts revision={0} onError={onError} />);
  expect(await screen.findByText(/Alertas indisponíveis/)).toBeTruthy();
  expect(screen.queryByText(/Nenhum evento/)).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Atualizar alertas" }));
  expect(await screen.findByText(/Nenhum evento neste filtro/)).toBeTruthy();
  expect(onError).toHaveBeenCalledTimes(1);
});
test("resolução administrativa não é descrita como recuperação TLS", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify(
          page([
            {
              ...event,
              resolved_at: "2026-10-13T13:00:00Z",
              resolution: "archived",
            },
          ]),
        ),
      ),
    ),
  );
  render(<Alerts revision={0} onError={vi.fn()} />);
  expect(await screen.findByText(/Endpoint arquivado/)).toBeTruthy();
  expect(screen.queryByText(/Recuperação confirmada/)).toBeNull();
});
