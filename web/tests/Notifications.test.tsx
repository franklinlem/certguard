import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { Notifications } from "../src/Notifications";
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
const response = (items: unknown[]) => new Response(JSON.stringify({ items }));
test("sem chave informa indisponibilidade e mantém histórico legível", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockImplementation(() => Promise.resolve(response([]))),
  );
  render(
    <Notifications
      csrf="t"
      configured={false}
      enabled={false}
      revision={0}
      onError={vi.fn()}
    />,
  );
  expect(await screen.findByText("Nenhum canal cadastrado.")).toBeTruthy();
  expect(screen.getByText(/Cadastro de canais indisponível/)).toBeTruthy();
  expect(
    screen.queryByRole("button", { name: "Salvar canal pausado" }),
  ).toBeNull();
});
test("cadastro usa CSRF, apaga segredo do formulário e não ativa canal", async () => {
  const mock = vi
    .fn()
    .mockImplementation((_path: string, options: RequestInit) =>
      Promise.resolve(
        options.method === "POST"
          ? new Response(JSON.stringify({ id: "c", enabled: false }), {
              status: 201,
            })
          : response([]),
      ),
    );
  vi.stubGlobal("fetch", mock);
  render(
    <Notifications
      csrf="csrf"
      configured
      enabled={false}
      revision={0}
      onError={vi.fn()}
    />,
  );
  await screen.findByText("Nenhum canal cadastrado.");
  fireEvent.click(screen.getByText("Cadastrar canal de notificação"));
  fireEvent.change(screen.getByLabelText("Nome do canal"), {
    target: { value: "Webhook autorizado" },
  });
  fireEvent.change(screen.getByLabelText("URL HTTPS do webhook"), {
    target: { value: "https://notify.example.com/hook" },
  });
  fireEvent.change(screen.getByLabelText(/Segredo de assinatura/), {
    target: { value: "s".repeat(32) },
  });
  fireEvent.click(screen.getByRole("button", { name: "Salvar canal pausado" }));
  expect(await screen.findByText(/Canal salvo pausado/)).toBeTruthy();
  const sent = mock.mock.calls.find(([, o]) => o.method === "POST");
  expect(sent?.[0]).toBe("/api/v1/alert-channels");
  expect(sent?.[1].headers["X-CSRF-Token"]).toBe("csrf");
  expect(JSON.parse(sent?.[1].body).enabled).toBeUndefined();
  expect(
    (screen.getByLabelText(/Segredo de assinatura/) as HTMLInputElement).value,
  ).toBe("");
});
test("ativação exige confirmação e só altera enabled", async () => {
  const mock = vi
    .fn()
    .mockImplementation((path: string, options: RequestInit) =>
      Promise.resolve(
        options.method === "PATCH"
          ? new Response(JSON.stringify({ id: "c" }))
          : response(
              path.includes("alert-channels")
                ? [
                    {
                      id: "c",
                      name: "Canal teste",
                      type: "smtp",
                      enabled: false,
                    },
                  ]
                : [],
            ),
      ),
    );
  vi.stubGlobal("fetch", mock);
  render(
    <Notifications
      csrf="csrf"
      configured
      enabled
      revision={0}
      onError={vi.fn()}
    />,
  );
  fireEvent.click(await screen.findByRole("button", { name: "Ativar canal" }));
  expect(screen.getByRole("alertdialog")).toBeTruthy();
  expect(mock.mock.calls.some(([, o]) => o.method === "PATCH")).toBe(false);
  fireEvent.click(screen.getByRole("button", { name: "Confirmar ativação" }));
  await waitFor(() =>
    expect(
      mock.mock.calls.some(
        ([, o]) => o.method === "PATCH" && o.body === '{"enabled":true}',
      ),
    ).toBe(true),
  );
});
