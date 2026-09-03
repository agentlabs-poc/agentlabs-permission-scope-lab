import { resolve } from 'node:path'
import { defineConfig } from 'vite'

export default defineConfig({
  build: {
    rollupOptions: {
      input: {
        playground: resolve(__dirname, 'index.html'),
        concept: resolve(__dirname, 'concept.html'),
        projectsExplorer: resolve(__dirname, 'projects-explorer.html'),
      },
    },
  },
})
