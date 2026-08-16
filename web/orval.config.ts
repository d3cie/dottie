import { defineConfig } from 'orval';

export default defineConfig({
  dottie: {
    input: './openapi.yaml',
    output: {
      target: './src/lib/api-client/generated/dottie.ts',
      schemas: './src/lib/api-client/generated/models',
      client: 'fetch',
      mode: 'single',
      clean: true,
      override: {
        mutator: {
          path: './src/lib/api-client/fetcher.ts',
          name: 'customFetch',
        },
        fetch: {
          includeHttpResponseReturnType: false,
        },
      },
    },
  },
});
