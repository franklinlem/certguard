import { expect, test } from "@playwright/test";
test("administrador cadastra, edita, pausa e arquiva endpoint sem acessar destinos externos", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByLabel("E-mail", { exact: true }).fill("admin@example.com");
  await page
    .getByLabel("Senha", { exact: true })
    .fill("Senha-E2E-descartavel-123");
  await page.getByRole("button", { name: "Entrar", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Endpoints TLS" }),
  ).toBeVisible();
  await expect(
    page.getByText(/Agendamento desabilitado neste ambiente/),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Adicionar endpoint", exact: true })
    .click();
  await expect(page.getByLabel("Nome", { exact: true })).toBeFocused();
  await expect(page.getByRole("heading", { name: "Adicionar endpoint", exact: true })).toBeInViewport();
  await page.getByLabel("Nome", { exact: true }).fill("Exemplo E2E");
  await page.getByLabel("Hostname", { exact: true }).fill("example.com");
  await page.getByRole("button", { name: "Salvar", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Exemplo E2E", exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("region", { name: "Visão geral" })).toBeVisible();
  await expect(
    page.getByText("○ Sem resultado", { exact: true }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Atualizar painel", exact: true })
    .click();
  await expect(
    page.getByText("Painel atualizado.", { exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Editar", exact: true }).click();
  await expect(page.getByLabel("Nome", { exact: true })).toBeFocused();
  await expect(page.getByRole("heading", { name: "Editar endpoint", exact: true })).toBeInViewport();
  await page.getByLabel("Nome", { exact: true }).fill("Exemplo atualizado");
  await page.getByRole("button", { name: "Salvar", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Exemplo atualizado", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Pausar", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Verificar agora", exact: true }),
  ).toBeDisabled();
  await page
    .getByRole("button", { name: "Detalhes e histórico", exact: true })
    .click();
  const historyTitle = page.getByRole("heading", { name: "Detalhes e histórico · Exemplo atualizado", exact: true });
  await expect(historyTitle).toBeFocused();
  await expect(historyTitle).toBeInViewport();
  await page.getByRole("button", { name: "Detalhes e histórico", exact: true }).click();
  await expect(historyTitle).toBeFocused();
  await expect(historyTitle).toBeInViewport();
  await expect(
    page.getByText("Nenhuma verificação nesta página.", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Ainda não há resultado para o destino atual."),
  ).toBeVisible();
  await page.getByRole("button", { name: "Arquivar", exact: true }).click();
  await page
    .getByRole("button", { name: "Confirmar arquivamento", exact: true })
    .click();
  await expect(
    page.getByText("Endpoint arquivado; histórico preservado.", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Alertas e eventos" }),
  ).toBeVisible();
  await expect(
    page.getByText("Nenhum evento neste filtro e página."),
  ).toBeVisible();
  await page.getByLabel("Exibir eventos", { exact: true }).selectOption("all");
  await expect(
    page.getByText("Nenhum evento neste filtro e página."),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Atualizar alertas", exact: true })
    .click();
  await expect(
    page.getByText("Nenhum evento neste filtro e página."),
  ).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Canais e entregas", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Nenhuma entrega nesta página.", { exact: true }),
  ).toBeVisible();
  await expect(page.getByText(/Cadastro de canais indisponível/)).toBeVisible();
  await page
    .getByRole("button", { name: "Atualizar entregas", exact: true })
    .click();
  await expect(
    page.getByText("Nenhuma entrega nesta página.", { exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  for (const path of ["/metrics", "/api/metrics", "/api/anything"]) {
    expect((await page.request.get(path)).status()).toBe(404);
  }
  await page.getByRole("button", { name: "Sair", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Entrar", exact: true }),
  ).toBeVisible();
});
