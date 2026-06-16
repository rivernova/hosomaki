<template>
  <nav class="sitenav" :class="{ scrolled: isScrolled }">

    <a class="sn-brand" :href="withBase('/')">
      <img src="/hosomaki_logo.png" alt="Hosomaki" />
      <span>Hosomaki</span>
    </a>

    <div class="sn-right" :class="{ open: menuOpen }">
      <a :href="withBase('/guide/introduction')" @click="menuOpen = false">Docs</a>
      <a :href="withBase('/reference/commands')" @click="menuOpen = false">Reference</a>
      <a href="https://github.com/rivernova/hosomaki" target="_blank" rel="noopener" @click="menuOpen = false">GitHub</a>

      <button
        class="sn-search"
        type="button"
        aria-label="Search documentation (Ctrl+K)"
        @click="openSearch"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
          fill="none" stroke="currentColor" stroke-width="2"
          stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
        </svg>
        <span>Search</span>
        <kbd aria-hidden="true">⌘K</kbd>
      </button>

      <button
        class="sn-theme-toggle"
        type="button"
        :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
        :title="isDark ? 'Light mode' : 'Dark mode'"
        @click="toggleDark"
      >

        <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
          viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <circle cx="12" cy="12" r="4"/>
          <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/>
        </svg>

        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16"
          viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
      </button>

      <a :href="withBase('/guide/installation')" class="sn-pill" @click="menuOpen = false">Get started</a>
    </div>

    <div class="sn-mobile-actions">
      <button
        class="sn-icon-btn"
        type="button"
        :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
        @click="toggleDark"
      >
        <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="18" height="18"
          viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <circle cx="12" cy="12" r="4"/>
          <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/>
        </svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="18" height="18"
          viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
      </button>

      <button class="sn-icon-btn" type="button" aria-label="Search documentation" @click="openSearch">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
          fill="none" stroke="currentColor" stroke-width="2"
          stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
        </svg>
      </button>
      <button
        class="sn-burger"
        type="button"
        :aria-expanded="menuOpen"
        aria-label="Toggle navigation menu"
        @click="menuOpen = !menuOpen"
      >
        <span :class="{ top: menuOpen }" />
        <span :class="{ mid: menuOpen }" />
        <span :class="{ bot: menuOpen }" />
      </button>
    </div>
  </nav>

  <Teleport to="body">
    <VPLocalSearchBox v-if="showSearch" @close="showSearch = false" />
  </Teleport>
</template>

<script setup>
import { ref, onMounted, onUnmounted, defineAsyncComponent } from 'vue'
import { useData, withBase } from 'vitepress'

const VPLocalSearchBox = defineAsyncComponent(() =>
  import('vitepress/dist/client/theme-default/components/VPLocalSearchBox.vue')
)

const menuOpen   = ref(false)
const isScrolled = ref(false)
const showSearch = ref(false)

const { isDark } = useData()
function toggleDark() {
  isDark.value = !isDark.value
}

function onScroll() {
  isScrolled.value = window.scrollY > 8
}

function openSearch() {
  showSearch.value = true
  menuOpen.value = false
}

function onKeyDown(e) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    showSearch.value = true
  }
  if (e.key === '/' && !['INPUT', 'TEXTAREA', 'SELECT'].includes(e.target.tagName) && !e.target.isContentEditable) {
    e.preventDefault()
    showSearch.value = true
  }
  if (e.key === 'Escape') {
    showSearch.value = false
    menuOpen.value = false
  }
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('keydown', onKeyDown)
  onScroll()
})
onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('keydown', onKeyDown)
})
</script>

<style scoped>
.sitenav {
  position: fixed; top: 0; left: 0; right: 0; z-index: 9999;
  height: 60px;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 clamp(1.25rem, 5vw, 3rem);
  background: rgba(242, 239, 233, .82);
  backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid transparent;
  transition: border-color .25s, background .25s;
  box-sizing: border-box;
}
.sitenav.scrolled {
  border-bottom-color: rgba(46, 43, 39, .1);
  background: rgba(242, 239, 233, .96);
}

.sn-brand {
  display: flex; align-items: center; gap: .55rem;
  text-decoration: none; color: #2e2b27;
  font-family: 'Inter', system-ui, sans-serif;
  font-weight: 600; font-size: .9rem; letter-spacing: -.01em;
  flex-shrink: 0;
}
.sn-brand img { width: 24px; height: 24px; object-fit: contain; border-radius: 4px; }

.sn-right { display: flex; align-items: center; gap: 1.75rem; }
.sn-right a {
  text-decoration: none; color: #6b6660;
  font-family: 'Inter', system-ui, sans-serif;
  font-size: .84rem; font-weight: 450; transition: color .15s;
}
.sn-right a:hover { color: #2e2b27; }

.sn-search {
  display: inline-flex; align-items: center; gap: .4rem;
  padding: .3rem .65rem;
  background: rgba(46, 43, 39, .06);
  border: 1px solid rgba(46, 43, 39, .12);
  border-radius: 7px;
  color: #a09b95;
  font-family: 'Inter', system-ui, sans-serif;
  font-size: .8rem; cursor: pointer;
  transition: background .15s, border-color .15s, color .15s;
  white-space: nowrap;
}
.sn-search:hover { background: rgba(46, 43, 39, .1); border-color: rgba(46, 43, 39, .2); color: #2e2b27; }
.sn-search kbd {
  font-family: 'JetBrains Mono', monospace; font-size: .67rem; color: #a09b95;
  background: rgba(46, 43, 39, .07); border: 1px solid rgba(46, 43, 39, .12);
  border-radius: 4px; padding: .05em .3em; margin-left: .1rem;
}

.sn-pill {
  background: #2e2b27 !important; color: #f2efe9 !important;
  padding: .36rem 1rem !important; border-radius: 100px;
  font-weight: 500 !important; font-size: .8rem !important;
  transition: background .2s !important;
}
.sn-pill:hover { background: #e06b50 !important; }

.sn-mobile-actions { display: none; align-items: center; gap: .35rem; }
.sn-icon-btn {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px;
  background: none; border: none; cursor: pointer; color: #6b6660;
  transition: color .15s;
}
.sn-icon-btn:hover { color: #2e2b27; }

.sn-burger {
  display: flex; flex-direction: column; justify-content: center; gap: 5px;
  width: 36px; height: 36px; background: none; border: none; cursor: pointer; padding: 6px;
}
.sn-burger span { display: block; height: 1.5px; background: #2e2b27; border-radius: 2px; transition: all .22s; }
.sn-burger .top { transform: translateY(6.5px) rotate(45deg); }
.sn-burger .mid { opacity: 0; transform: scaleX(0); }
.sn-burger .bot { transform: translateY(-6.5px) rotate(-45deg); }

@media (max-width: 720px) {
  .sn-mobile-actions { display: flex; }
  .sn-search { display: none; }

  .sn-right {
    position: fixed; inset: 60px 0 0 0;
    flex-direction: column; align-items: flex-start; gap: 0;
    background: #f2efe9; border-top: 1px solid rgba(46, 43, 39, .1);
    padding: 1.25rem clamp(1.25rem, 5vw, 3rem) 2rem;
    transform: translateY(-110%); opacity: 0;
    transition: transform .26s ease, opacity .26s ease;
    pointer-events: none;
  }
  .sn-right.open { transform: none; opacity: 1; pointer-events: auto; }
  .sn-right a {
    font-size: 1.05rem; padding: .8rem 0;
    border-bottom: 1px solid rgba(46, 43, 39, .1); width: 100%;
  }
  .sn-right a:last-child { border-bottom: none; }
  .sn-pill {
    background: none !important; color: #e06b50 !important;
    padding: .8rem 0 !important; border-radius: 0 !important; font-size: 1.05rem !important;
  }
}

.sn-theme-toggle {
  display: flex; align-items: center; justify-content: center;
  width: 30px; height: 30px;
  background: none; border: none; cursor: pointer;
  color: #6b6660; border-radius: 6px;
  transition: color .15s, background .15s;
}
.sn-theme-toggle:hover { color: #2e2b27; background: rgba(46,43,39,.06); }
</style>
