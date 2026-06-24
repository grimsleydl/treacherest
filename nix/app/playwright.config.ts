import { defineConfig, devices } from '@playwright/test';

const baseURL = process.env.BASE_URL ?? 'http://localhost:8080';
const parsedBaseURL = new URL(baseURL);
const webServerPort = Number(parsedBaseURL.port || (parsedBaseURL.protocol === 'https:' ? 443 : 80));
const repoRoot = process.env.PRJ_ROOT ?? '../..';

export default defineConfig({
  testDir: './tests/playwright',
  timeout: 30 * 1000,
  expect: {
    timeout: 5000
  },
  fullyParallel: false, // Run tests sequentially for SSE tests
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // Single worker for SSE connection tests
  reporter: 'html',
  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: {
    command: process.env.PLAYWRIGHT_WEB_SERVER_COMMAND ?? `cd ${repoRoot} && just _serve-test ${webServerPort}`,
    port: webServerPort,
    timeout: 120 * 1000,
    reuseExistingServer: !process.env.CI,
  },
});
