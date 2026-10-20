import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: { '/api': 'http://127.0.0.1:8080' },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
    clearMocks: true,
    restoreMocks: true,
    coverage: {
      provider: 'v8',
      include: ['src/App.tsx', 'src/api.ts', 'src/operations.ts'],
      reporter: ['text', 'html', 'json-summary', 'lcov'],
      reportsDirectory: '../reports/frontend',
      thresholds: { statements: 90, branches: 85, functions: 90, lines: 90 },
    },
  },
});
