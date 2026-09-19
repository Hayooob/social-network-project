import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  // Without this the React plugin is imported but never registered, so Fast
  // Refresh does not run and edits reload the whole page.
  plugins: [react()],

  server: {
    proxy: {
      // Same paths nginx proxies in production, so dev and prod behave alike.
      '/api': 'http://localhost:8080',
      '/uploads': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
})
