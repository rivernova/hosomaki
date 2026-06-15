import DefaultTheme from 'vitepress/theme'
import { h } from 'vue'
import DocsLayout from './DocsLayout.vue'
import LandingPage from './LandingPage.vue'
import SiteNav from './SiteNav.vue'
import './custom.css'

export default {
    ...DefaultTheme,

    Layout: DocsLayout,

    enhanceApp({ app }) {
        app.component('LandingPage', LandingPage)
        app.component('SiteNav', SiteNav)
    },
}