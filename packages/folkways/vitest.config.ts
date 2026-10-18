import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['contract/**/*.test.ts', 'test/**/*.test.ts'],
    setupFiles: ['contract/setup.ts'],
    testTimeout: 20_000,
  },
});
