import { defineConfig } from 'orval';

export default defineConfig({
  dottie: {
    input: './openapi.yaml',
    output: {
      target: './packages/api-client/src/generated/dottie.ts',
      schemas: './packages/api-client/src/generated/models',
      client: 'fetch',
      mode: 'single',
      clean: true,
      override: {
        mutator: {
          path: './packages/api-client/src/fetcher.ts',
          name: 'customFetch',
        },
        fetch: {
          includeHttpResponseReturnType: false,
        },
      },
    },
  },
});
