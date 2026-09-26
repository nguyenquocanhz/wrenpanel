<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Sao lưu & Phục hồi (Backups)</h1>
        <p class="text-xs text-slate-400 mt-1">Định dạng nén chuẩn tar.zst kèm manifest.json lưu trữ thông tin phiên bản panel, runtime và CSDL</p>
      </div>

      <button
        @click="isCreateModalOpen = true"
        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
      >
        + Tạo bản Sao lưu mới
      </button>
    </div>

    <!-- Table of Backups -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Tên file sao lưu</th>
            <th class="py-3 px-4">Phạm vi (Scope)</th>
            <th class="py-3 px-4">Nơi lưu trữ</th>
            <th class="py-3 px-4">Thời gian tạo</th>
            <th class="py-3 px-4 text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="loading" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Đang tải danh sách bản sao lưu...</td>
          </tr>
          <tr v-else-if="backups.length === 0" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Chưa có bản sao lưu nào. Bấm "Tạo bản Sao lưu mới".</td>
          </tr>
          <tr v-for="b in backups" :key="b.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-mono font-medium text-slate-100">
              {{ b.file_path }}
            </td>
            <td class="py-3 px-4">
              <span
                class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase"
                :class="{
                  'bg-blue-900/60 text-blue-300 border border-blue-700': b.scope === 'site',
                  'bg-emerald-900/60 text-emerald-300 border border-emerald-700': b.scope === 'database',
                  'bg-amber-900/60 text-amber-300 border border-amber-700': b.scope === 'full',
                }"
              >
                {{ b.scope }}
              </span>
            </td>
            <td class="py-3 px-4 text-slate-300 uppercase font-mono text-[11px]">
              {{ b.destination }}
            </td>
            <td class="py-3 px-4 text-slate-400">
              {{ b.created_at }}
            </td>
            <td class="py-3 px-4 text-right">
              <button
                class="px-2.5 py-1 text-blue-400 hover:text-blue-300 bg-blue-950/40 rounded border border-blue-900/60 text-xs"
              >
                Tải về
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <div v-if="isCreateModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-md w-full shadow-2xl p-6 space-y-4">
        <h3 class="text-base font-semibold text-slate-100">Tạo bản Sao lưu mới</h3>

        <form @submit.prevent="submitCreateBackup" class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-medium mb-1">Cấp độ Sao lưu (Scope)</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                @click="form.scope = 'site'"
                :class="form.scope === 'site' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300'"
                class="py-2 rounded border border-slate-700 transition"
              >
                Từng Site
              </button>
              <button
                type="button"
                @click="form.scope = 'database'"
                :class="form.scope === 'database' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300'"
                class="py-2 rounded border border-slate-700 transition"
              >
                Database
              </button>
              <button
                type="button"
                @click="form.scope = 'full'"
                :class="form.scope === 'full' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300'"
                class="py-2 rounded border border-slate-700 transition"
              >
                Full Server
              </button>
            </div>
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Nơi lưu trữ đích</label>
            <select
              v-model="form.destination"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            >
              <option value="local">Cục bộ (Local Staging /var/lib/wrenpanel/backups/)</option>
              <option value="s3">Amazon S3 / S3-compatible</option>
              <option value="ftp">FTP / SFTP Remote</option>
              <option value="rsync">Rsync Server</option>
            </select>
          </div>

          <div class="pt-4 border-t border-slate-700 flex justify-end gap-2">
            <button
              type="button"
              @click="isCreateModalOpen = false"
              class="px-4 py-2 rounded border border-slate-700 text-slate-300 hover:bg-slate-800 transition"
            >
              Hủy
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium transition"
            >
              Bắt đầu sao lưu
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { backupsApi } from '../../api/backups'
import { useAccountStore } from '../../stores/account'

const accountStore = useAccountStore()
const backups = ref([])
const loading = ref(false)
const isCreateModalOpen = ref(false)

const form = ref({
  scope: 'site',
  destination: 'local',
})

async function loadData() {
  loading.value = true
  try {
    backups.value = await backupsApi.list(accountStore.selectedAccountId)
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function submitCreateBackup() {
  if (!accountStore.currentAccount) {
    alert('Vui lòng chọn tài khoản hosting.')
    return
  }
  try {
    await backupsApi.create({
      account_id: accountStore.currentAccount.id,
      scope: form.value.scope,
      destination: form.value.destination,
    })
    isCreateModalOpen.value = false
    alert('Bản sao lưu đã được lập lịch và khởi tạo thành công!')
    await loadData()
  } catch (err) {
    alert(`Lỗi: ${err.message}`)
  }
}

watch(() => accountStore.selectedAccountId, () => {
  loadData()
})

onMounted(() => {
  loadData()
})
</script>
