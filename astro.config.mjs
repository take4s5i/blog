import { defineConfig } from "astro/config";
import preact from "@astrojs/preact";

// https://astro.build/config
export default defineConfig({
  integrations: [preact()],
  site: "https://blog.take4s5i.dev",
});
