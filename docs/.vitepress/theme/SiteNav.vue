<template>
  <nav class="sitenav" :class="{ scrolled: isScrolled }">

    <a class="sn-brand" href="/">
      <img src="/hosomaki_logo.png" alt="Hosomaki" />
      <span>Hosomaki</span>
    </a>

    <!-- Desktop links + search + pill -->
    <div class="sn-right" :class="{ open: menuOpen }">
      <a href="/guide/introduction" @click="menuOpen = false">Docs</a>
      <a href="/reference/commands" @click="menuOpen = false">Reference</a>
      <a
        href="https://github.com/rivernova/hosomaki"
        target="_blank"
        rel="noopener"
        @click="menuOpen = false"
      >GitHub</a>

      <!-- Search trigger — dispatches Ctrl+K which VitePress's own listener catches -->
      <button
        class="sn-search"
        type="button"
        aria-label="Search documentation (Ctrl+K)"
        @click="openSearch"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="15"
          height="15"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </svg>
        <span class="sn-search-label">Search</span>
        <kbd class="sn-search-kbd" aria-hidden="true">⌘K</kbd>
      </button>

      <a href="/guide/installation" class="sn-pill" @click="menuOpen = false">
        Get started
      </a>
    </div>

    <!-- Mobile hamburger -->
    <div class="sn-mobile-actions">
      <!-- Search icon only on mobile (no label/kbd) -->
      <button
        class="sn-search-icon-only"
        type="button"
        aria-label="Search documentation"
        @click="openSearch"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
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
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const menuOpen   = ref(false)
const isScrolled = ref(false)

function onScroll() {
  isScrolled.value = window.scrollY > 8
}

/**
 * Open VitePress's built-in local search modal.
 *
 * VitePress registers a keydown listener for Ctrl/Cmd+K that sets its internal
 * `showSearch` ref to true, which mounts VPLocalSearchBox. We trigger that
 * same listener by dispatching a synthetic keyboard event — no internal imports
 * required, and the modal stays fully owned by VitePress.
 */
function openSearch() {
  const event = new KeyboardEvent('keydown', {
    key: 'k',
    ctrlKey: true,
    metaKey: true, // covers macOS Cmd+K as well
    bubbles: true,
    cancelable: true,
  })
  window.dispatchEvent(event)
  // Close mobile menu if open
  menuOpen.value = false
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()
})
onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<style scoped>
/* ── Layout ─────────────────────────────────────────────────────────── */
.sitenav {
  position: fixed; top: 0; left: 0; right: 0; z-index: 9999;
  height: 60px;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 clamp(1.25rem, 5vw, 3rem);
  background: rgba(242, 239, 233, .82);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid transparent;
  transition: border-color .25s, background .25s;
  box-sizing: border-box;
}
.sitenav.scrolled {
  border-bottom-color: rgba(46, 43, 39, .1);
  background: rgba(242, 239, 233, .96);
}

/* ── Brand ──────────────────────────────────────────────────────────── */
.sn-brand {
  display: flex; align-items: center; gap: .55rem;
  text-decoration: none; color: #2e2b27;
  font-family: 'Inter', system-ui, sans-serif;
  font-weight: 600; font-size: .9rem; letter-spacing: -.01em;
  flex-shrink: 0;
}
.sn-brand img { width: 24px; height: 24px; object-fit: contain; border-radius: 4px; }

/* ── Right group (desktop) ──────────────────────────────────────────── */
.sn-right {
  display: flex; align-items: center; gap: 1.75rem;
}
.sn-right a {
  text-decoration: none; color: #6b6660;
  font-family: 'Inter', system-ui, sans-serif;
  font-size: .84rem; font-weight: 450;
  transition: color .15s;
}
.sn-right a:hover { color: #2e2b27; }

/* ── Search button (desktop) ────────────────────────────────────────── */
.sn-search {
  display: inline-flex; align-items: center; gap: .45rem;
  padding: .32rem .7rem;
  background: rgba(46, 43, 39, .06);
  border: 1px solid rgba(46, 43, 39, .12);
  border-radius: 7px;
  color: #a09b95;
  font-family: 'Inter', system-ui, sans-serif;
  font-size: .8rem;
  cursor: pointer;
  transition: background .15s, border-color .15s, color .15s;
  white-space: nowrap;
}
.sn-search:hover {
  background: rgba(46, 43, 39, .1);
  border-color: rgba(46, 43, 39, .2);
  color: #2e2b27;
}
.sn-search-label { font-size: .8rem; }
.sn-search-kbd {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: .68rem;
  color: #a09b95;
  background: rgba(46, 43, 39, .07);
  border: 1px solid rgba(46, 43, 39, .12);
  border-radius: 4px;
  padding: .05em .3em;
  margin-left: .1rem;
}

/* ── Pill ───────────────────────────────────────────────────────────── */
.sn-pill {
  background: #2e2b27 !important;
  color: #f2efe9 !important;
  padding: .36rem 1rem !important;
  border-radius: 100px;
  font-weight: 500 !important;
  font-size: .8rem !important;
  transition: background .2s !important;
}
.sn-pill:hover { background: #e06b50 !important; }

/* ── Mobile actions container ───────────────────────────────────────── */
.sn-mobile-actions { display: none; align-items: center; gap: .5rem; }

/* Search icon-only (mobile) */
.sn-search-icon-only {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px;
  background: none; border: none; cursor: pointer;
  color: #6b6660;
  transition: color .15s;
}
.sn-search-icon-only:hover { color: #2e2b27; }

/* ── Hamburger ──────────────────────────────────────────────────────── */
.sn-burger {
  display: flex; flex-direction: column; justify-content: center; gap: 5px;
  width: 36px; height: 36px;
  background: none; border: none; cursor: pointer; padding: 6px;
}
.sn-burger span {
  display: block; height: 1.5px;
  background: #2e2b27; border-radius: 2px;
  transition: all .22s;
}
.sn-burger .top { transform: translateY(6.5px) rotate(45deg); }
.sn-burger .mid { opacity: 0; transform: scaleX(0); }
.sn-burger .bot { transform: translateY(-6.5px) rotate(-45deg); }

/* ── Mobile breakpoint (≤ 720px) ────────────────────────────────────── */
@media (max-width: 720px) {
  .sn-mobile-actions { display: flex; }

  /* The desktop sn-right becomes a full-screen overlay */
  .sn-right {
    position: fixed;
    inset: 60px 0 0 0;
    flex-direction: column; align-items: flex-start; gap: 0;
    background: #f2efe9;
    border-top: 1px solid rgba(46, 43, 39, .1);
    padding: 1.25rem clamp(1.25rem, 5vw, 3rem) 2rem;
    transform: translateY(-110%); opacity: 0;
    transition: transform .26s ease, opacity .26s ease;
    pointer-events: none;
  }
  .sn-right.open { transform: none; opacity: 1; pointer-events: auto; }

  /* Links in the mobile overlay */
  .sn-right a {
    font-size: 1.05rem;
    padding: .8rem 0;
    border-bottom: 1px solid rgba(46, 43, 39, .1);
    width: 100%;
  }
  .sn-right a:last-child { border-bottom: none; }

  /* Pill resets to a plain salmon text link in the overlay */
  .sn-pill {
    background: none !important;
    color: #e06b50 !important;
    padding: .8rem 0 !important;
    border-radius: 0 !important;
    font-size: 1.05rem !important;
  }

  /* Hide the desktop search button — mobile uses the icon-only button */
  .sn-search { display: none; }
}
</style>