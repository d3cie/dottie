import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    lib: {
      entry: 'src/index.ts',
      name: 'Dottie',
      formats: ['iife'],
      fileName: () => 'tracker.js',
    },
    minify: 'esbuild',
  },
});

