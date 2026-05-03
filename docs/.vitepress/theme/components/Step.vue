<script setup lang="ts">
import { onMounted, ref } from "vue";

defineProps<{
  num: string;
  title: string;
}>();

const root = ref<HTMLElement | null>(null);

onMounted(() => {
  const el = root.value;
  if (!el) return;
  const reduced = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;
  if (reduced || typeof IntersectionObserver === "undefined") return;
  el.classList.add("hiw-step--prime");
  const io = new IntersectionObserver(
    (entries) => {
      for (const e of entries) {
        if (e.isIntersecting) {
          el.classList.add("is-visible");
          io.disconnect();
          break;
        }
      }
    },
    { threshold: 0.05 },
  );
  io.observe(el);
  setTimeout(() => el.classList.add("is-visible"), 50);
});
</script>

<template>
  <section
    ref="root"
    class="hiw-step"
    :aria-labelledby="`step-${num}-title`"
    :style="{ transitionDelay: `${(parseInt(num) - 1) * 60}ms` }"
  >
    <h3 :id="`step-${num}-title`" class="hiw-step__title">{{ title }}</h3>
    <div class="hiw-step__body">
      <slot />
    </div>
  </section>
</template>
