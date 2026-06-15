<template>
  <div class="docs-wrap">
    <SiteNav />
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
.VPNav        { display: none !important; }
.VPNavBar     { display: none !important; }
.VPNavScreen  { display: none !important; }

:root { --vp-nav-height: 60px; }
</style>