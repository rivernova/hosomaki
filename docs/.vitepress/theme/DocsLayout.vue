<template>
  <div class="docs-wrap">
    <!-- Our shared nav — identical to the landing page -->
    <SiteNav />
    <!-- VitePress default layout, with its own nav hidden via CSS -->
    <DefaultLayout>
      <template v-for="(_, name) in $slots" #[name]="slotData">
        <slot :name="name" v-bind="slotData" />
      </template>
    </DefaultLayout>
  </div>
</template>

<script setup>
import DefaultTheme from 'vitepress/theme'
import SiteNav from './SiteNav.vue'

const DefaultLayout = DefaultTheme.Layout
</script>

<style>
/* Hide VitePress's own nav — we render SiteNav instead */
.VPNav        { display: none !important; }
.VPNavBar     { display: none !important; }
.VPNavScreen  { display: none !important; }

/* VitePress offsets the sidebar and content by --vp-nav-height.
   Our nav is 60px tall — keep that var in sync. */
:root { --vp-nav-height: 60px; }
</style>