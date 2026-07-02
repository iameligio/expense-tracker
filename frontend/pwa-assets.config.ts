import {
  defineConfig,
  minimal2023Preset,
} from '@vite-pwa/assets-generator/config'

// Generates the PWA icon set (pwa-192/512, maskable-512, apple-touch-180) into
// public/ from the app logo. Run with `npm run generate-pwa-assets`.
export default defineConfig({
  headLinkOptions: { preset: '2023' },
  preset: minimal2023Preset,
  images: ['public/favicon.svg'],
})
