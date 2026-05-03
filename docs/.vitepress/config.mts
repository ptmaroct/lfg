import { defineConfig } from "vitepress";

export default defineConfig({
  title: "lfg",
  description:
    "Open-source TUI bootstrap CLI — make a new dev machine feel like home in minutes.",
  cleanUrls: true,
  lastUpdated: true,
  head: [
    ["link", { rel: "icon", type: "image/svg+xml", href: "/logo.svg" }],
  ],
  themeConfig: {
    logo: "/logo.svg",
    siteTitle: "lfg",
    nav: [
      { text: "Install", link: "/install" },
      { text: "Quick start", link: "/quick-start" },
      { text: "Guides", link: "/guides/bundles" },
      { text: "Reference", link: "/reference/cli" },
    ],
    sidebar: [
      {
        text: "Start here",
        items: [
          { text: "Introduction", link: "/introduction" },
          { text: "Install", link: "/install" },
          { text: "Quick start", link: "/quick-start" },
        ],
      },
      {
        text: "Guides",
        items: [
          { text: "Bundles & tools", link: "/guides/bundles" },
          { text: "Themes", link: "/guides/themes" },
          { text: "Custom presets", link: "/guides/presets" },
          { text: "Backup & restore", link: "/guides/backup" },
          { text: "Doctor", link: "/guides/doctor" },
          { text: "Self-update", link: "/guides/update" },
          { text: "Key bindings", link: "/guides/keys" },
        ],
      },
      {
        text: "Reference",
        items: [
          { text: "CLI commands", link: "/reference/cli" },
          { text: "Preset TOML schema", link: "/reference/preset-schema" },
          { text: "Environment variables", link: "/reference/env" },
          { text: "State files", link: "/reference/state" },
        ],
      },
      {
        text: "Internals",
        items: [
          { text: "Architecture", link: "/internals/architecture" },
          { text: "Installer & PATH", link: "/internals/installer" },
          { text: "Version resolver", link: "/internals/resolver" },
          { text: "Snapshot tests", link: "/internals/snapshots" },
        ],
      },
      {
        text: "Project",
        items: [
          { text: "Roadmap", link: "/project/roadmap" },
          { text: "Contributing", link: "/project/contributing" },
          { text: "Design principles", link: "/project/principles" },
        ],
      },
    ],
    socialLinks: [
      { icon: "github", link: "https://github.com/ptmaroct/lfg" },
    ],
    editLink: {
      pattern: "https://github.com/ptmaroct/lfg/edit/main/docs/:path",
      text: "Edit this page on GitHub",
    },
    search: {
      provider: "local",
    },
    footer: {
      message: "Released under the MIT License.",
      copyright: "© ptmaroct",
    },
  },
});
