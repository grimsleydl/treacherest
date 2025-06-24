import { defineConfig, devices } from '@playwright/test';

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
    baseURL: 'http://localhost:8080',
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
    command: process.env.PRJ_ROOT 
      ? `cd ${process.env.PRJ_ROOT}/nix/app && dev`
      : 'cd ../.. && nix develop --command bash -c "cd nix/app && dev"',
    port: 8080,
    timeout: 120 * 1000,
    reuseExistingServer: !process.env.CI,
  },
});