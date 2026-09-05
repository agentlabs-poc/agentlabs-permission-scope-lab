import { resolve } from 'node:path'
import { defineConfig } from 'vite'

export default defineConfig({
  // Retired page URLs must return 404, not fall back to the handbook homepage.
  appType: 'mpa',
  build: {
    rollupOptions: {
      input: {
        handbook: resolve(__dirname, 'index.html'),
        concept: resolve(__dirname, 'concept.html'),
      },
    },
  },
})
