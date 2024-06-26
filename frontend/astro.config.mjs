import { defineConfig } from "astro/config";
import react from "@astrojs/react";
import astroI18next from "astro-i18next";
import tailwind from "@astrojs/tailwind";

// https://astro.build/config
export default defineConfig({
  output: "static",
  site: import.meta.env.PUBLIC_BASE_URL,
  integrations: [react(), astroI18next(), tailwind()],
  server: { port: 3000 },
  outDir: "build",
  vite: {
    server: {
      proxy: {
        "/api": {
          target: "http://server:8084",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ""),
        },
      },
    },
  },
});
