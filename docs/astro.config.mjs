// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import starlightThemeTerminal from "starlight-theme-terminal";

export default defineConfig({
  site: "https://ptmaroct.github.io",
  base: "/lfg",
  integrations: [
    starlight({
      plugins: [starlightThemeTerminal()],
      title: "lfg",
      description:
        "Open-source TUI bootstrap CLI — make a new dev machine feel like home in minutes.",
      logo: { src: "./src/assets/logo.svg", replacesTitle: false },
      social: {
        github: "https://github.com/ptmaroct/lfg",
      },
      editLink: {
        baseUrl: "https://github.com/ptmaroct/lfg/edit/main/docs/",
      },
      lastUpdated: true,
      customCss: ["./src/styles/custom.css"],
      sidebar: [
        {
          label: "Start here",
          items: [
            { label: "Introduction", slug: "introduction" },
            { label: "Install", slug: "install" },
            { label: "Quick start", slug: "quick-start" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Bundles & tools", slug: "guides/bundles" },
            { label: "Themes", slug: "guides/themes" },
            { label: "Custom presets", slug: "guides/presets" },
            { label: "Backup & restore", slug: "guides/backup" },
            { label: "Doctor", slug: "guides/doctor" },
            { label: "Self-update", slug: "guides/update" },
            { label: "Key bindings", slug: "guides/keys" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "CLI commands", slug: "reference/cli" },
            { label: "Preset TOML schema", slug: "reference/preset-schema" },
            { label: "Environment variables", slug: "reference/env" },
            { label: "State files", slug: "reference/state" },
          ],
        },
        {
          label: "Internals",
          items: [
            { label: "Architecture", slug: "internals/architecture" },
            { label: "Installer & PATH", slug: "internals/installer" },
            { label: "Version resolver", slug: "internals/resolver" },
            { label: "Snapshot tests", slug: "internals/snapshots" },
          ],
        },
        {
          label: "Project",
          items: [
            { label: "Roadmap", slug: "project/roadmap" },
            { label: "Contributing", slug: "project/contributing" },
            { label: "Design principles", slug: "project/principles" },
          ],
        },
      ],
    }),
  ],
});
