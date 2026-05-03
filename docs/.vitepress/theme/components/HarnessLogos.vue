<script setup lang="ts">
const harnesses = [
  {
    name: "Claude Code",
    href: "https://docs.claude.com/en/docs/claude-code",
    logo: "claude",
  },
  {
    name: "Codex",
    href: "https://github.com/openai/codex",
    logo: "codex",
  },
];
</script>

<template>
  <div class="hiw-harness">
    <a
      v-for="h in harnesses"
      :key="h.name"
      class="hiw-harness__item"
      :href="h.href"
      target="_blank"
      rel="noopener"
      :aria-label="h.name"
    >
      <!-- Claude — stylised concentric/sunburst glyph -->
      <svg
        v-if="h.logo === 'claude'"
        viewBox="0 0 24 24"
        fill="currentColor"
        aria-hidden="true"
      >
        <path
          d="M12 2c.4 0 .73.31.78.7l.4 3.06a4.5 4.5 0 0 0 3.86 3.85l3.05.4a.79.79 0 0 1 0 1.57l-3.05.4a4.5 4.5 0 0 0-3.86 3.86l-.4 3.05a.79.79 0 0 1-1.56 0l-.4-3.05A4.5 4.5 0 0 0 6.96 11.99l-3.05-.4a.79.79 0 0 1 0-1.57l3.05-.4a4.5 4.5 0 0 0 3.86-3.85l.4-3.06A.79.79 0 0 1 12 2z"
        />
      </svg>
      <!-- Codex — wordmark-y monogram -->
      <svg
        v-else-if="h.logo === 'codex'"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <circle cx="12" cy="12" r="9" />
        <path d="M9 9.5l-2.5 2.5L9 14.5" />
        <path d="M15 9.5l2.5 2.5L15 14.5" />
      </svg>
    </a>
  </div>
</template>
