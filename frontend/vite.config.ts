import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    chunkSizeWarningLimit: 3000,
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          markdown: ['@kangc/v-md-editor', 'highlight.js', 'katex'],
        },
      },
    },
  },
  server: {
    port: 5173,
    // 开发时把 /api 与 /uploads 代理到本地后端，免跨域
    proxy: {
      // timeout/proxyTimeout：后端不可用时请求及时失败，而不是无限挂起
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        timeout: 15000,
        proxyTimeout: 15000,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        timeout: 15000,
        proxyTimeout: 15000,
      },
    },
  },
})
