import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, new URL('.', import.meta.url).pathname, '');

  return {
    base: env.ORCASTACK_WEB_BASE || '/',
    plugins: [react()],
    server: {
      host: '0.0.0.0',
      port: 5050,
    },
  };
});
