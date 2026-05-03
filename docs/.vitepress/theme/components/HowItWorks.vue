<script setup lang="ts">
import Step from "./Step.vue";

const skills = [
  "agent-browser",
  "frontend-design",
  "claude-api",
  "skill-creator",
];
</script>

<template>
  <div class="hiw">
    <Step
      num="01"
      title="Pick the essentials"
      subtitle="We don't judge — Python 2 still lives in our hearts too."
    >
      <p>
        Brew, Node, Python, runtime managers. Anything already on your box is
        flagged with its current version so you don't reinstall it. Tick what
        you want, leave the rest.
      </p>
    </Step>

    <Step
      num="02"
      title="Pick your AI harness"
      subtitle="Bring your own brain."
    >
      <p>
        Claude Code, Codex, OpenCode, Droid — one picker, every coding agent.
        Click the boxes; lfg handles the install, the PATH, and the auth
        nudges.
      </p>
    </Step>

    <Step
      num="03"
      title="Add the skills you actually use"
      subtitle="Your daily essentials and more — covered."
    >
      <p>
        Cross-harness skills install via <code>npx skills add</code> and land in
        every agent that's on your machine.
      </p>
      <div class="hiw-chips" aria-label="Featured skills">
        <span v-for="s in skills" :key="s" class="hiw-chip">{{ s }}</span>
      </div>
    </Step>

    <Step
      num="04"
      title="Make it yours, then share it"
      subtitle="The whole point."
    >
      <p>
        Author a <code>preset.toml</code>, host it anywhere
        (gist, S3, your own repo), and your team bootstraps with
        <code>lfg --config &lt;url&gt;</code>. Onboarding goes from a day to a
        coffee.
      </p>
      <p>
        <a href="/guides/presets">Learn how to write a preset →</a>
      </p>
    </Step>
  </div>
</template>
