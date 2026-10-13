import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { App } from "../src/App";
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
function response(data: unknown, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
test("mostra login sem sessão e comunica credenciais inválidas", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(response({ error: "Sessão ausente" }, 401))
    .mockResolvedValueOnce(
      response({ error: "E-mail ou senha inválidos." }, 401),
    );
  vi.stubGlobal("fetch", fetchMock);
  render(<App />);
  await screen.findByRole("button", { name: "Entrar" });
  fireEvent.change(screen.getByLabelText("E-mail"), {
    target: { value: "admin@example.com" },
  });
  fireEvent.change(screen.getByLabelText("Senha"), {
    target: { value: "errada" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Entrar" }));
  expect((await screen.findByRole("alert")).textContent).toContain(
    "E-mail ou senha inválidos",
  );
});
test("carrega endpoints reais e envia check autenticado com CSRF", async () => {
  const endpoint = {
    id: "123",
    name: "Exemplo",
    hostname: "example.com",
    enabled: true,
    last_check_at: null,
  };
  const fetchMock = vi.fn().mockImplementation((path: string) => {
    if (path.endsWith("/auth/me"))
      return Promise.resolve(
        response({
          user: { id: "u", email: "admin@example.com" },
          csrf_token: "csrf-test",
        }),
      );
    if (path.includes("/checks?"))
      return Promise.resolve(
        response({
          items: [
            {
              id: "c",
              result: {
                hostname: "example.com",
                status: "error",
                code: "blocked_target",
                checked_at: "2026-10-12T12:00:00Z",
              },
            },
          ],
        }),
      );
    if (path.endsWith("/check"))
      return Promise.resolve(response({ id: "c" }, 201));
    return Promise.resolve(response({ items: [endpoint] }));
  });
  vi.stubGlobal("fetch", fetchMock);
  render(<App />);
  fireEvent.click(
    await screen.findByRole("button", { name: "Verificar agora" }),
  );
  expect(
    (await screen.findByText("Destino bloqueado pela proteção SSRF"))
      .textContent,
  ).toContain("SSRF");
  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/endpoints/123/check",
      expect.objectContaining({
        method: "POST",
        credentials: "same-origin",
        headers: expect.objectContaining({ "X-CSRF-Token": "csrf-test" }),
      }),
    ),
  );
  expect(
    screen.getByText(/Agendamento desabilitado neste ambiente/),
  ).toBeTruthy();
});

test("mostra agenda habilitada, próxima execução e retenção", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockImplementation((path: string) =>
        Promise.resolve(
          response(
            path.endsWith("/auth/me")
              ? {
                  user: { id: "u", email: "admin@example.com" },
                  csrf_token: "t",
                  background: {
                    scheduler_enabled: true,
                    retention_enabled: true,
                    retention_days: 180,
                  },
                }
              : {
                  items: [
                    {
                      id: "e",
                      name: "Site",
                      hostname: "example.com",
                      enabled: true,
                      last_check_at: null,
                      next_check_at: "2026-10-13T12:00:00Z",
                    },
                  ],
                },
          ),
        ),
      ),
  );
  render(<App />);
  expect(await screen.findByText(/Agendamento habilitado/)).toBeTruthy();
  expect(
    await screen.findByText(/Próxima execução prevista:.*2026/),
  ).toBeTruthy();
  expect(screen.getByText(/Resultados com mais de 180 dias/)).toBeTruthy();
});
