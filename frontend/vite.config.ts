import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Served locally through Laravel Valet at https://expense-tracker.test, which
// proxies TLS traffic to this dev server. The dev server in turn proxies /api
// to the Go backend, so the SPA stays same-origin with the API and the HttpOnly
// refresh cookie works without CORS friction.
export default defineConfig({
  plugins: [react()],
  server: {
    host: '127.0.0.1',
    port: 5173,
    // Accept the Valet host header (Vite blocks unknown hosts by default).
    allowedHosts: ['expense-tracker.test', 'localhost', '127.0.0.1'],
    // Run HMR over Valet's TLS (wss on 443) so hot reload works on the domain.
    hmr: {
      host: 'expense-tracker.test',
      protocol: 'wss',
      clientPort: 443,
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
