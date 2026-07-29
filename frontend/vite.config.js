import process from 'node:process'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const cafeName = env.VITE_CAFE_NAME || 'KafePlat'

  return {
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.svg', 'apple-touch-icon.png', 'icon-192x192.png', 'icon-512x512.png'],
      manifest: {
        name: cafeName,
        short_name: cafeName,
        description: cafeName + ' - boshqaruv tizimi va yetkazib berish',
        theme_color: '#f8f7f5',
        background_color: '#f8f7f5',
        display: 'standalone',
        icons: [
          {
            src: 'icon-192x192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'icon-512x512.png',
            sizes: '512x512',
            type: 'image/png'
          },
          {
            src: 'favicon.svg',
            sizes: 'any',
            type: 'image/svg+xml'
          }
        ]
      }
    })
  ],
  build: {
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // Core React libraries — always cached together
          if (id.includes('node_modules/react/') || id.includes('node_modules/react-dom/')) {
            return 'vendor-react';
          }
          // Router
          if (id.includes('node_modules/react-router-dom/') || id.includes('node_modules/react-router/')) {
            return 'vendor-router';
          }
          // Animation library (large, rarely changes)
          if (id.includes('node_modules/framer-motion/')) {
            return 'vendor-animations';
          }
          // Heaviest pages get their own chunk to reduce initial load
          if (id.includes('/src/pages/Admin')) return 'page-admin';
          if (id.includes('/src/pages/Waiter')) return 'page-waiter';
          if (id.includes('/src/pages/Cashier')) return 'page-cashier';
          if (id.includes('/src/pages/Kitchen')) return 'page-kitchen';
          // All other node_modules together
          if (id.includes('node_modules/')) return 'vendor-misc';
        }
      }
    },
    // Warn when chunks exceed 500 kB
    chunkSizeWarningLimit: 500,
  },
  server: {
    proxy: {
      '/api': {
        // Lokal backend ishlamasa — productionga bog'lanadi
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
  }
})
