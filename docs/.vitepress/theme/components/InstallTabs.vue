<script setup lang="ts">
import { ref, onMounted, computed } from "vue";

type Method = {
  id: string;
  label: string;
  blurb: string;
  cmd: string;
  follow?: string;
};

const methods: Method[] = [
  {
    id: "brew",
    label: "brew",
    blurb: "macOS + Linux · via the official tap",
    cmd: "brew install ptmaroct/tap/lfg",
    follow: "lfg",
  },
  {
    id: "curl",
    label: "curl",
    blurb: "no Homebrew required · drops into /usr/local/bin or ~/.local/bin",
    cmd: "curl -fsSL https://raw.githubusercontent.com/ptmaroct/lfg/main/install.sh | sh",
    follow: "lfg",
  },
  {
    id: "go",
    label: "go",
    blurb: "Go 1.25+ · binary lands in $GOBIN",
    cmd: "go install github.com/ptmaroct/lfg/cmd/lfg@latest",
    follow: "lfg",
  },
];

const active = ref<string>("brew");
const current = computed(
  () => methods.find((m) => m.id === active.value) ?? methods[0],
);

const tablistEl = ref<HTMLElement | null>(null);
const copied = ref<string | null>(null);

const select = (id: string) => {
  active.value = id;
};

const onKeydown = (e: KeyboardEvent) => {
  const idx = methods.findIndex((m) => m.id === active.value);
  if (idx < 0) return;

  let next: number | null = null;
  if (e.key === "ArrowRight") next = (idx + 1) % methods.length;
  else if (e.key === "ArrowLeft") next = (idx - 1 + methods.length) % methods.length;
  else if (e.key === "Home") next = 0;
  else if (e.key === "End") next = methods.length - 1;

  if (next !== null) {
    e.preventDefault();
    active.value = methods[next].id;
    const tab = tablistEl.value?.querySelectorAll<HTMLButtonElement>(
      "[role=tab]",
    )[next];
    tab?.focus();
  }
};

const copy = async (text: string, id: string) => {
  try {
    await navigator.clipboard.writeText(text);
    copied.value = id;
    setTimeout(() => {
      if (copied.value === id) copied.value = null;
    }, 1400);
  } catch {
    // No clipboard permission — silently fail. Selection still works.
  }
};

const primed = ref(false);
onMounted(() => {
  // Small tick so the stagger fires after VitePress hydrates
  requestAnimationFrame(() => {
    primed.value = true;
  });
});
</script>

<template>
  <section
    class="itabs"
    :class="{ 'itabs--prime': primed }"
    aria-labelledby="itabs-heading"
  >
    <h3 id="itabs-heading" class="itabs__sr">Install lfg</h3>

    <div
      ref="tablistEl"
      class="itabs__tabs"
      role="tablist"
      aria-label="Installation method"
      @keydown="onKeydown"
    >
      <button
        v-for="(m, i) in methods"
        :key="m.id"
        :id="`itab-${m.id}`"
        :class="['itabs__tab', { 'is-active': active === m.id }]"
        :style="{ '--itabs-stagger': `${i * 60}ms` }"
        role="tab"
        type="button"
        :aria-selected="active === m.id"
        :aria-controls="`itab-panel-${m.id}`"
        :tabindex="active === m.id ? 0 : -1"
        @click="select(m.id)"
      >
        <span class="itabs__tab-glyph" aria-hidden="true">$</span>
        <span class="itabs__tab-label">{{ m.label }}</span>
      </button>

      <span class="itabs__rule" aria-hidden="true" />
    </div>

    <div
      :id="`itab-panel-${current.id}`"
      class="itabs__panel"
      role="tabpanel"
      :aria-labelledby="`itab-${current.id}`"
      :key="current.id"
    >
      <div class="itabs__chrome" aria-hidden="true">
        <span class="itabs__dot itabs__dot--r" />
        <span class="itabs__dot itabs__dot--y" />
        <span class="itabs__dot itabs__dot--g" />
        <span class="itabs__chrome-label">~/ {{ current.id }}</span>

        <button
          class="itabs__copy"
          type="button"
          :aria-label="`Copy ${current.label} install command`"
          @click="copy(current.cmd, current.id)"
        >
          <span v-if="copied === current.id" class="itabs__copy-on">copied</span>
          <span v-else class="itabs__copy-off">
            <svg
              viewBox="0 0 24 24"
              width="13"
              height="13"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <rect x="9" y="9" width="11" height="11" rx="2" />
              <path d="M5 15V5a2 2 0 0 1 2-2h10" />
            </svg>
            copy
          </span>
        </button>
      </div>

      <pre class="itabs__pre"><code><span class="itabs__prompt">$</span> <span class="itabs__cmd">{{ current.cmd }}</span><template v-if="current.follow">
<span class="itabs__prompt">$</span> <span class="itabs__cmd">{{ current.follow }}</span></template></code></pre>

      <p class="itabs__blurb">{{ current.blurb }}</p>
    </div>
  </section>
</template>
