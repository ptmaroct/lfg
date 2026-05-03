<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from "vue";

const slides = [
  {
    src: "/screens/welcome.png",
    label: "01 — Welcome",
    caption: "Live host probe + theme cycle.",
  },
  {
    src: "/screens/tree.png",
    label: "02 — Pick the essentials",
    caption: "Bundles + tools, pre-checked if installed.",
  },
  {
    src: "/screens/tree-skills.png",
    label: "03 — Add the skills",
    caption: "Cross-harness skill packs install via npx.",
  },
  {
    src: "/screens/confirm.png",
    label: "04 — Confirm before it runs",
    caption: "See every command before a single one fires.",
  },
  {
    src: "/screens/install.png",
    label: "05 — Watch it install",
    caption: "PATH re-augmented between every step.",
  },
];

const idx = ref(0);
const isPaused = ref(false);
let timer: ReturnType<typeof setInterval> | null = null;
let io: IntersectionObserver | null = null;
const root = ref<HTMLElement | null>(null);

const current = computed(() => slides[idx.value]);

const goto = (i: number) => {
  idx.value = (i + slides.length) % slides.length;
  resetTimer();
};
const next = () => goto(idx.value + 1);
const prev = () => goto(idx.value - 1);

const startTimer = () => {
  if (timer) return;
  timer = setInterval(() => {
    if (!isPaused.value) idx.value = (idx.value + 1) % slides.length;
  }, 3800);
};
const stopTimer = () => {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
};
const resetTimer = () => {
  stopTimer();
  startTimer();
};

const onKey = (e: KeyboardEvent) => {
  if (e.key === "ArrowLeft") prev();
  else if (e.key === "ArrowRight") next();
};

onMounted(() => {
  if (!root.value) return;
  io = new IntersectionObserver(
    (entries) => {
      for (const e of entries) {
        if (e.isIntersecting) startTimer();
        else stopTimer();
      }
    },
    { threshold: 0.2 },
  );
  io.observe(root.value);
});
onBeforeUnmount(() => {
  stopTimer();
  io?.disconnect();
});
</script>

<template>
  <figure
    ref="root"
    class="sflow"
    @mouseenter="isPaused = true"
    @mouseleave="isPaused = false"
    @focusin="isPaused = true"
    @focusout="isPaused = false"
    @keydown="onKey"
    tabindex="0"
    role="region"
    aria-roledescription="carousel"
    aria-label="lfg journey"
  >
    <div class="sflow__stage">
      <img
        v-for="(s, i) in slides"
        :key="s.src"
        :src="s.src"
        :alt="s.label"
        class="sflow__img"
        :class="{ 'is-active': i === idx }"
        loading="eager"
        decoding="async"
      />
    </div>

    <figcaption class="sflow__caption">
      <span class="sflow__label">{{ current.label }}</span>
      <span class="sflow__sep" aria-hidden="true">·</span>
      <span class="sflow__text">{{ current.caption }}</span>
    </figcaption>

    <div class="sflow__dots" role="tablist" aria-label="Choose screen">
      <button
        v-for="(s, i) in slides"
        :key="s.src"
        type="button"
        role="tab"
        :aria-selected="i === idx"
        :aria-label="s.label"
        class="sflow__dot"
        :class="{ 'is-active': i === idx }"
        @click="goto(i)"
      />
    </div>
  </figure>
</template>
