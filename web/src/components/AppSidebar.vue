<template>
  <aside class="w-64 bg-[#0b1120] border-r border-slate-800 flex flex-col h-screen sticky top-0 select-none">
    <!-- Brand -->
    <div class="h-16 flex items-center px-6 border-b border-slate-800 gap-3">
      <div class="w-8 h-8 rounded bg-blue-600 flex items-center justify-center font-black text-white text-lg tracking-wider shadow">
        W
      </div>
      <div>
        <div class="font-bold text-slate-100 tracking-wide leading-none">WrenPanel</div>
        <div class="text-[10px] text-slate-400 font-mono mt-1">VPS & HOSTING CONTROL</div>
      </div>
    </div>

    <!-- Navigation Menu (aaPanel / 1Panel style) -->
    <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
      <div class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider px-3 mb-2">Web Services</div>

      <router-link
        to="/vhosts"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/vhosts') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
        </svg>
        Website & Vhosts
      </router-link>

      <router-link
        to="/ssl"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/ssl') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
        SSL & ACME (Lego)
      </router-link>

      <div class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider px-3 mt-6 mb-2">App & Runtimes</div>

      <router-link
        to="/apps"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/apps') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        App Manager (Systemd)
      </router-link>

      <router-link
        to="/php"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/php') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
        </svg>
        PHP Manager
      </router-link>

      <router-link
        to="/node"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/node') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        Node.js Runtime
      </router-link>

      <router-link
        to="/python"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/python') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
        Python Runtime
      </router-link>

      <div class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider px-3 mt-6 mb-2">Storage & Data</div>

      <router-link
        to="/databases"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/databases') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
        </svg>
        Databases (MySQL/PG)
      </router-link>

      <router-link
        to="/ftp"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/ftp') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
        </svg>
        FTP / SFTP Accounts
      </router-link>

      <router-link
        to="/backups"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/backups') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
        </svg>
        Backups (tar.zst)
      </router-link>

      <div class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider px-3 mt-6 mb-2">Management</div>

      <router-link
        to="/accounts"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/accounts') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
        Hosting Accounts
      </router-link>

      <router-link
        to="/system"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition"
        :class="$route.path.startsWith('/system') ? 'bg-blue-600/20 text-blue-400 font-medium border border-blue-500/30' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        Audit Log & Status
      </router-link>
    </nav>

    <!-- Bottom Status -->
    <div class="p-4 border-t border-slate-800 text-xs text-slate-400 flex items-center justify-between">
      <span class="flex items-center gap-1.5">
        <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
        API Online
      </span>
      <span class="font-mono text-[11px]">v0.1.0</span>
    </div>
  </aside>
</template>
