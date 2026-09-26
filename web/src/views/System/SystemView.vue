<template>
  <div class="space-y-6">
    <!-- System Status Cards -->
    <div class="grid grid-cols-4 gap-4">
      <div class="bg-[#1e293b] border border-slate-800 rounded-lg p-4">
        <div class="text-xs text-slate-400">Phiên bản Panel</div>
        <div class="text-lg font-bold text-slate-100 mt-1">{{ status?.version || '0.1.0' }}</div>
        <div class="text-[11px] text-blue-400 mt-1">Single Binary Go</div>
      </div>

      <div class="bg-[#1e293b] border border-slate-800 rounded-lg p-4">
        <div class="text-xs text-slate-400">Hệ điều hành / Kiến trúc</div>
        <div class="text-lg font-bold text-slate-100 mt-1 uppercase">{{ status?.os }} / {{ status?.arch }}</div>
        <div class="text-[11px] text-slate-400 mt-1">{{ status?.num_cpu }} CPU Cores</div>
      </div>

      <div class="bg-[#1e293b] border border-slate-800 rounded-lg p-4">
        <div class="text-xs text-slate-400">Root Worker IPC</div>
        <div class="text-lg font-bold text-emerald-400 mt-1 flex items-center gap-1.5">
          <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
          {{ status?.worker_status || 'Connected' }}
        </div>
        <div class="text-[11px] text-slate-400 mt-1">Unix Socket 0660</div>
      </div>

      <div class="bg-[#1e293b] border border-slate-800 rounded-lg p-4">
        <div class="text-xs text-slate-400">Go Runtime</div>
        <div class="text-lg font-bold text-slate-100 mt-1 font-mono text-sm">{{ status?.go_version }}</div>
        <div class="text-[11px] text-emerald-400 mt-1">Memory safe & compiled</div>
      </div>
    </div>

    <!-- Audit Logs Table -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-base font-bold text-slate-100">Nhật ký Thao tác Quản trị (Audit Log)</h2>
        <span class="text-xs text-slate-400 font-mono">Đồng bộ cả SQLite và /var/log/wrenpanel/audit.log</span>
      </div>

      <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
            <tr>
              <th class="py-3 px-4">Thời gian</th>
              <th class="py-3 px-4">Người thực hiện</th>
              <th class="py-3 px-4">Hành động</th>
              <th class="py-3 px-4">Đối tượng</th>
              <th class="py-3 px-4">Kết quả</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-slate-200">
            <tr v-if="loading" class="text-center text-slate-400">
              <td colspan="5" class="py-8">Đang tải nhật ký audit...</td>
            </tr>
            <tr v-else-if="auditLogs.length === 0" class="text-center text-slate-400">
              <td colspan="5" class="py-8">Chưa có bản ghi audit nào.</td>
            </tr>
            <tr v-for="log in auditLogs" :key="log.id" class="hover:bg-slate-800/40 transition">
              <td class="py-3 px-4 font-mono text-[11px] text-slate-400">
                {{ log.created_at }}
              </td>
              <td class="py-3 px-4 font-semibold text-slate-200">
                {{ log.actor }}
              </td>
              <td class="py-3 px-4 font-mono text-[11px] text-blue-400">
                {{ log.action }}
              </td>
              <td class="py-3 px-4 font-mono text-[11px] text-slate-300">
                {{ log.target || '-' }}
              </td>
              <td class="py-3 px-4">
                <span
                  class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase"
                  :class="log.result === 'success' ? 'bg-emerald-950/60 text-emerald-400 border border-emerald-800' : 'bg-red-950/60 text-red-400 border border-red-800'"
                >
                  {{ log.result }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { systemApi } from '../../api/system'

const status = ref(null)
const auditLogs = ref([])
const loading = ref(false)

async function loadData() {
  loading.value = true
  try {
    const [sData, aData] = await Promise.all([
      systemApi.getStatus(),
      systemApi.listAuditLogs(100),
    ])
    status.value = sData
    auditLogs.value = aData || []
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
