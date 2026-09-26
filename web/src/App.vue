<template>
  <div v-if="route.name === 'login'" class="min-h-screen bg-[#0b1120]">
    <router-view />
  </div>
  <div v-else class="flex min-h-screen bg-[#0b1120] text-slate-100">
    <!-- Left fixed Sidebar (aaPanel style) -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0">
      <AppHeader />
      <main class="flex-1 p-8 overflow-y-auto">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'
import AppSidebar from './components/AppSidebar.vue'
import AppHeader from './components/AppHeader.vue'

const route = useRoute()
</script>
