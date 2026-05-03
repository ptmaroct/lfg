import DefaultTheme from "vitepress/theme";
import type { Theme } from "vitepress";
import HowItWorks from "./components/HowItWorks.vue";
import Step from "./components/Step.vue";
import HarnessLogos from "./components/HarnessLogos.vue";
import ScreenFlow from "./components/ScreenFlow.vue";
import InstallTabs from "./components/InstallTabs.vue";
import "./style.css";

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component("HowItWorks", HowItWorks);
    app.component("Step", Step);
    app.component("HarnessLogos", HarnessLogos);
    app.component("ScreenFlow", ScreenFlow);
    app.component("InstallTabs", InstallTabs);
  },
} satisfies Theme;
