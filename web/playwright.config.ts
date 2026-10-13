import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "./e2e",
  use: { baseURL: "http://localhost:8088" },
  projects: [
    { name: "desktop", use: { viewport: { width: 1366, height: 900 } } },
    {
      name: "mobile",
      use: { viewport: { width: 390, height: 844 }, colorScheme: "dark" },
    },
  ],
  retries: 0,
  workers: 1,
});
