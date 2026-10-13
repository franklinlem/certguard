import { expect, test } from "@playwright/test";

test("painéis abertos ficam visíveis e recebem foco sem aguardar histórico", async ({ page }) => {
  const endpoints = Array.from({ length: 20 }, (_, index) => ({
    id: `endpoint-${index}`,
    name: `Site ${index + 1}`,
    hostname: "example.com",
    enabled: true,
    last_check_at: null,
  }));
  await page.route("**/api/v1/**", async (route) => {
    const path = new URL(route.request().url()).pathname;
    const data = path.endsWith("/auth/me")
      ? { user: { id: "test", email: "admin@example.com" }, csrf_token: "test" }
      : path.endsWith("/dashboard")
        ? {
            items: endpoints,
            summary: { total: 20, enabled: 20, paused: 0, healthy: 0, warning: 0, critical: 0, errors: 0, unchecked: 20, expiring_30_days: 0 },
            generated_at: "2026-10-13T12:00:00Z",
          }
        : path.endsWith("/alerts")
          ? { items: [], summary: { active: 0, critical: 0, resolved: 0, info: 0 }, generated_at: "2026-10-13T12:00:00Z" }
          : { items: [] };
    // Keep history pending: navigation must work before the request finishes.
    if (path.endsWith("/checks")) return;
    await route.fulfill({ json: data });
  });
  await page.goto("/");
  const add = page.getByRole("button", { name: "Adicionar endpoint", exact: true });
  await add.click();
  const name = page.getByLabel("Nome", { exact: true });
  await expect(name).toBeFocused();
  await expect(name).toBeInViewport();
  await expect(page.getByRole("heading", { name: "Adicionar endpoint", exact: true })).toBeInViewport();
  await name.fill("Rascunho");
  await add.click();
  await expect(name).toBeFocused();
  await expect(name).toHaveValue("Rascunho");
  const details = page.getByRole("button", { name: "Detalhes e histórico", exact: true });
  await details.first().click();
  let title = page.getByRole("heading", { name: "Detalhes e histórico · Site 1", exact: true });
  await expect(title).toBeFocused();
  await expect(title).toBeInViewport();
  await expect(page.getByText("Carregando histórico…")).toBeVisible();
  await details.first().click();
  await expect(title).toBeFocused();
  await expect(title).toBeInViewport();
  await details.last().click();
  title = page.getByRole("heading", { name: "Detalhes e histórico · Site 20", exact: true });
  await expect(title).toBeFocused();
  await expect(title).toBeInViewport();
  await page.getByRole("button", { name: "Editar", exact: true }).last().click();
  await expect(name).toBeFocused();
  await expect(name).toHaveValue("Site 20");
  await expect(name).toBeInViewport();
});
