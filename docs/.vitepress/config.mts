import { defineConfig } from "vitepress";

const SITE_URL = "https://lfg-docs.netlify.app";
const SITE_NAME = "LFG";
const SITE_DESCRIPTION =
  "Open-source TUI bootstrap CLI — make a new dev machine feel like home in minutes.";
const OG_IMAGE = `${SITE_URL}/screens/welcome.png`;
const TWITTER_HANDLE = "@waahbete";

const HERO_CSS = `
:root {
  --vp-c-brand-1: #a78bfa;
  --vp-c-brand-2: #c4b5fd;
  --vp-c-brand-3: #a78bfa;
  --vp-c-brand-soft: rgba(167, 139, 250, 0.16);

  --vp-home-hero-name-color: transparent;
  --vp-home-hero-name-background: linear-gradient(120deg, #f472b6 0%, #a78bfa 50%, #34d399 100%);

  --vp-home-hero-image-background-image: linear-gradient(120deg, rgba(244,114,182,0.45), rgba(52,211,153,0.45));
  --vp-home-hero-image-filter: blur(80px);
}

.dark {
  --vp-c-brand-1: #c4b5fd;
  --vp-c-brand-2: #a78bfa;
  --vp-c-brand-3: #c4b5fd;
}

/* Hero typography — terminal-bold gradient on the headline (text), mono tagline */
.VPHero .text {
  font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1.05;
  background: linear-gradient(120deg, #f472b6 0%, #a78bfa 50%, #34d399 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  font-size: clamp(2rem, 6.5vw, 3.5rem);
  max-width: 16ch;
}
.VPHero .tagline {
  font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
  font-weight: 400;
  font-size: 0.95rem;
  letter-spacing: 0;
  line-height: 1.55;
  color: var(--vp-c-text-2);
  max-width: 38em;
}

/* Feature cards — mono titles, tighter look */
.VPFeature .title {
  font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
  font-weight: 700;
  letter-spacing: -0.02em;
  font-size: 1.02rem;
}
.VPFeature .details {
  font-size: 0.9rem;
  line-height: 1.55;
}

/* Body sections under the home hero — gradient bar above each H2 */
.VPHome > .vp-doc h2,
.VPHome .container .vp-doc h2 {
  font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
  font-weight: 700;
  letter-spacing: -0.02em;
  border-top: none;
  padding-top: 2.75rem;
  margin-top: 3rem;
  position: relative;
}
.VPHome > .vp-doc h2::before,
.VPHome .container .vp-doc h2::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 2.5rem;
  height: 2px;
  background: linear-gradient(90deg, #f472b6, #a78bfa, #34d399);
  border-radius: 2px;
}

/* Constrain + center the home body markdown */
.VPHome > .vp-doc,
.VPHome .container .vp-doc {
  max-width: 720px;
  margin: 0 auto;
  padding: 0 24px 4rem;
}
.VPHome img[src*='/screens/'] {
  border-radius: 10px;
  box-shadow: 0 1px 0 rgba(255,255,255,0.04) inset, 0 24px 60px -20px rgba(167,139,250,0.25);
  margin-top: 1.25rem;
}

/* Nav title — logo IS the wordmark, so trim the gap */
.VPNavBar .VPNavBarTitle .title { gap: 0.5rem; }

/* Footer — mono, brand link color */
.VPFooter .message,
.VPFooter .copyright {
  font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.82rem;
  letter-spacing: 0;
}
.VPFooter a {
  color: var(--vp-c-brand-1);
  text-decoration: none;
  transition: opacity 0.15s ease;
}
.VPFooter a:hover { opacity: 0.7; }
`;

export default defineConfig({
  title: SITE_NAME,
  description: SITE_DESCRIPTION,
  cleanUrls: true,
  lastUpdated: true,
  sitemap: { hostname: SITE_URL },
  head: [
    // Favicons
    ["link", { rel: "icon", type: "image/svg+xml", href: "/favicon.svg" }],
    ["link", { rel: "alternate icon", href: "/favicon.svg" }],
    ["link", { rel: "apple-touch-icon", href: "/favicon.svg" }],
    ["link", { rel: "mask-icon", href: "/favicon.svg", color: "#a78bfa" }],
    ["meta", { name: "theme-color", content: "#0b0b12", media: "(prefers-color-scheme: dark)" }],
    ["meta", { name: "theme-color", content: "#ffffff", media: "(prefers-color-scheme: light)" }],
    ["meta", { name: "color-scheme", content: "light dark" }],

    // Generic SEO
    ["meta", { name: "author", content: "Anuj Sharma" }],
    ["meta", { name: "keywords", content: "lfg, cli, bootstrap, dev machine, dotfiles, brew, mise, claude code, codex, tui, charm" }],
    ["meta", { name: "robots", content: "index, follow" }],
    ["meta", { name: "application-name", content: SITE_NAME }],
    ["meta", { name: "apple-mobile-web-app-title", content: SITE_NAME }],

    // Open Graph
    ["meta", { property: "og:type", content: "website" }],
    ["meta", { property: "og:site_name", content: SITE_NAME }],
    ["meta", { property: "og:locale", content: "en_US" }],
    ["meta", { property: "og:title", content: `${SITE_NAME} — feel back at home, in minutes` }],
    ["meta", { property: "og:description", content: SITE_DESCRIPTION }],
    ["meta", { property: "og:url", content: SITE_URL }],
    ["meta", { property: "og:image", content: OG_IMAGE }],
    ["meta", { property: "og:image:width", content: "1558" }],
    ["meta", { property: "og:image:height", content: "1008" }],
    ["meta", { property: "og:image:alt", content: "LFG TUI welcome screen" }],

    // Twitter / X
    ["meta", { name: "twitter:card", content: "summary_large_image" }],
    ["meta", { name: "twitter:site", content: TWITTER_HANDLE }],
    ["meta", { name: "twitter:creator", content: TWITTER_HANDLE }],
    ["meta", { name: "twitter:title", content: `${SITE_NAME} — feel back at home, in minutes` }],
    ["meta", { name: "twitter:description", content: SITE_DESCRIPTION }],
    ["meta", { name: "twitter:image", content: OG_IMAGE }],

    // Fonts
    ["link", { rel: "preconnect", href: "https://fonts.googleapis.com" }],
    ["link", { rel: "preconnect", href: "https://fonts.gstatic.com", crossorigin: "" }],
    [
      "link",
      {
        href: "https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700;800&display=swap",
        rel: "stylesheet",
      },
    ],
    ["style", {}, HERO_CSS],
  ],

  // Per-page canonical + og:title/description overrides.
  // VitePress emits frontmatter title/description as <title> and <meta name="description">,
  // but the og:* and canonical tags still need to be derived per page.
  transformPageData(pageData) {
    const canonicalUrl = `${SITE_URL}/${pageData.relativePath}`
      .replace(/index\.md$/, "")
      .replace(/\.md$/, "");
    const pageTitle = pageData.title
      ? `${pageData.title} | ${SITE_NAME}`
      : `${SITE_NAME} — feel back at home, in minutes`;
    const pageDesc = pageData.description || SITE_DESCRIPTION;

    pageData.frontmatter.head ??= [];
    pageData.frontmatter.head.push(
      ["link", { rel: "canonical", href: canonicalUrl }],
      ["meta", { property: "og:title", content: pageTitle }],
      ["meta", { property: "og:description", content: pageDesc }],
      ["meta", { property: "og:url", content: canonicalUrl }],
      ["meta", { name: "twitter:title", content: pageTitle }],
      ["meta", { name: "twitter:description", content: pageDesc }],
    );
  },
  themeConfig: {
    logo: "/logo.svg",
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
      { icon: "x", link: "https://x.com/waahbete" },
    ],
    editLink: {
      pattern: "https://github.com/ptmaroct/lfg/edit/main/docs/:path",
      text: "Edit this page on GitHub",
    },
    search: {
      provider: "local",
    },
    footer: {
      message:
        'Built by <a href="https://anujsh.com" target="_blank" rel="noopener">Anuj Sharma</a> · <a href="https://x.com/waahbete" target="_blank" rel="noopener">@waahbete</a>',
      copyright: "Released under the MIT License.",
    },
  },
});
