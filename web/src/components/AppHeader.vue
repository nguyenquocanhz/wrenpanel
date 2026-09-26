<template>
  <header class="h-16 bg-[#0f172a] border-b border-slate-800 flex items-center justify-between px-8 sticky top-0 z-30">
    <!-- Breadcrumb or Title -->
    <div class="flex items-center gap-3">
      <h2 class="text-base font-semibold text-slate-200">
        {{ currentTitle }}
      </h2>
    </div>

    <!-- Right Controls -->
    <div class="flex items-center gap-4">
      <!-- Active Hosting Account Selector -->
      <div class="flex items-center gap-2">
        <label class="text-xs text-slate-400 font-medium">Hosting User:</label>
        <select
          v-model="accountStore.selectedAccountId"
          class="bg-[#1e293b] border border-slate-700 text-slate-200 text-xs rounded px-2.5 py-1.5 focus:outline-none focus:border-blue-500 font-medium"
        >
          <option v-for="acc in accountStore.accounts" :key="acc.id" :value="acc.id">
            {{ acc.username }} ({{ acc.home_dir }})
          </option>
        </select>
      </div>

      <!-- Root Worker Indicator -->
      <div class="flex items-center gap-1.5 px-2.5 py-1 rounded bg-slate-800 border border-slate-700 text-xs">
        <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
        <span class="text-slate-300 font-medium">Worker: Allowlist Mode</span>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAccountStore } from '../stores/account'

const route = useRoute()
const accountStore = useAccountStore()

const currentTitle = computed(() => {
  if (route.path.startsWith('/vhosts')) return 'Website & Vhosts Management'
  if (route.path.startsWith('/apps')) return 'App Manager (Systemd Isolated Units)'
  if (route.path.startsWith('/php')) return 'PHP Manager (Multi-version Side-by-side)'
  if (route.path.startsWith('/node')) return 'Node.js Runtime Manager'
  if (route.path.startsWith('/python')) return 'Python Runtime Manager'
  if (route.path.startsWith('/databases')) return 'Database Services (MySQL / Postgres / Mongo)'
  if (route.path.startsWith('/ftp')) return 'FTP / SFTP Accounts Management'
  if (route.path.startsWith('/ssl')) return 'SSL / TLS Manager (Let\'s Encrypt ACME)'
  if (route.path.startsWith('/backups')) return 'Backup & Disaster Recovery (tar.zst)'
  if (route.path.startsWith('/accounts')) return 'Linux Hosting Accounts'
  if (route.path.startsWith('/system')) return 'Audit Logs & System Status'
  return 'WrenPanel'
})
</script>
