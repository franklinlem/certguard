import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test } from "vitest";
import {
  CertificateDetails,
  HistoryChart,
  SummaryCards,
} from "../src/Dashboard";
import type { Check } from "../src/api";
afterEach(cleanup);

const check = (
  id: string,
  day: number,
  days?: number,
  hostname = "example.com",
): Check => ({
  id,
  result: {
    hostname,
    port: 443,
    checked_at: `2026-10-${String(day).padStart(2, "0")}T12:00:00Z`,
    status: "healthy",
    code: "success",
    days_remaining: days,
  },
});

test("contadores globais mostram zero real e não dependem da lista paginada", () => {
  render(
    <SummaryCards
      summary={{
        total: 80,
        enabled: 75,
        paused: 5,
        healthy: 70,
        warning: 3,
        critical: 2,
        errors: 0,
        unchecked: 5,
        expiring_30_days: 4,
      }}
      generatedAt="2026-10-12T12:00:00Z"
    />,
  );
  expect(screen.getByText("80")).toBeTruthy();
  expect(screen.getByText("0")).toBeTruthy();
  expect(screen.getByText(/incluindo pausados/)).toBeTruthy();
});
test("detalhes exibem certificado e diferenciam validação desconhecida", () => {
  render(
    <CertificateDetails
      result={{
        ...check("a", 1, 20).result,
        resolved_ip: "93.184.216.34",
        tls_version: "TLS 1.3",
        cipher_suite: "TLS_AES_128_GCM_SHA256",
        certificate: {
          issuer: "CA de teste",
          subject: "CN=example.com",
          serial_number: "abcd",
          not_before: "2026-09-01T00:00:00Z",
          not_after: "2026-11-01T00:00:00Z",
          dns_names: ["example.com", "*.example.com"],
          fingerprint_sha256: "fingerprint",
        },
      }}
    />,
  );
  expect(screen.getByText("CA de teste")).toBeTruthy();
  expect(screen.getByText("abcd")).toBeTruthy();
  expect(screen.getByText("Não avaliado / Não avaliado")).toBeTruthy();
  expect(screen.getByText("example.com, *.example.com")).toBeTruthy();
});
test("gráfico interrompe falhas e mudança de destino, preservando ordem temporal e zero", () => {
  const { container } = render(
    <HistoryChart
      checks={[
        check("e", 7, 20, "other.com"),
        check("d", 6, 0),
        check("c", 5),
        check("b", 2, 9),
        check("a", 1, 10),
      ]}
    />,
  );
  expect(screen.getAllByTestId("history-segment")).toHaveLength(1);
  expect(container.querySelectorAll("circle")).toHaveLength(4);
  expect(screen.getByRole("img").getAttribute("aria-labelledby")).toBeTruthy();
  const cells = Array.from(container.querySelectorAll("tbody tr"));
  expect(cells[0].textContent).toContain("01/10/2026");
  expect(cells[2].textContent).toContain("Não disponível");
  expect(cells[3].textContent).toContain("0");
});
test("sem certificado não desenha curva fictícia", () => {
  render(<HistoryChart checks={[check("a", 1)]} />);
  expect(screen.queryByRole("img")).toBeNull();
  expect(screen.getByText(/Sem dados de validade/)).toBeTruthy();
});
