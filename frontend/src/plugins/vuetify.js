import Vue from "vue"
import Vuetify from "vuetify/lib"
import tailwind from "../../tailwind.config"
import { prefersDarkMode } from "@/utils/general_utils"

Vue.use(Vuetify)

export default new Vuetify({
  theme: {
    dark: prefersDarkMode(),
    themes: {
      light: {
        primary: tailwind.theme.colors.green,
        error: tailwind.theme.colors.red,
      },
      dark: {
        primary: tailwind.theme.colors["light-green"],
        error: tailwind.theme.colors.red,
        // background (the page, behind .v-application) is darker than
        // surface (cards/dialogs/menus) on purpose, so elevated surfaces are
        // visibly distinct from the page instead of blending into it -
        // matches --color-page-bg/--color-white in index.css.
        background: "#121212",
        surface: "#1e1e1e",
      },
    },
  },
  breakpoint: {
    thresholds: {
      xs: 640,
      sm: 768,
      md: 1024,
      lg: 1280,
    },
    scrollBarWidth: 0,
  },
})
