<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Quản lý Ứng dụng (App Manager)</h1>
        <p class="text-xs text-slate-400 mt-1">Mỗi ứng dụng chạy độc lập qua Systemd Service tự sinh, tự động cấp port nội bộ và tự restart khi crash</p>
      </div>

      <button
        @click="openAddAppModal"
        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
      >
        + Triển khai App mới
      </button>
    </div>

    <!-- Table of Apps -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Tên App</th>
            <th class="py-3 px-4">Loại (Runtime)</th>
            <th class="py-3 px-4">Port Nội bộ</th>
            <th class="py-3 px-4">Systemd Unit</th>
            <th class="py-3 px-4">Trạng thái</th>
            <th class="py-3 px-4 text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="loading" class="text-center text-slate-400">
            <td colspan="6" class="py-8">Đang tải danh sách app...</td>
          </tr>
          <tr v-else-if="apps.length === 0" class="text-center text-slate-400">
            <td colspan="6" class="py-8">Chưa có ứng dụng nào. Bấm "Triển khai App mới" để bắt đầu.</td>
          </tr>
          <tr v-for="app in apps" :key="app.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-bold text-slate-100">
              {{ app.name }}
            </td>
            <td class="py-3 px-4">
              <span
                class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase"
                :class="{
                  'bg-emerald-900/60 text-emerald-300 border border-emerald-700': app.kind === 'node',
                  'bg-yellow-900/60 text-yellow-300 border border-yellow-700': app.kind === 'python',
                  'bg-blue-900/60 text-blue-300 border border-blue-700': app.kind === 'docker',
                }"
              >
                {{ app.kind }} {{ app.runtime_version ? `(${app.runtime_version})` : '' }}
              </span>
            </td>
            <td class="py-3 px-4 font-mono font-semibold text-amber-300">
              :{{ app.internal_port || 'Chưa cấp' }}
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-400">
              {{ app.systemd_unit || 'N/A' }}
            </td>
            <td class="py-3 px-4">
              <span v-if="app.status === 'running'" class="text-emerald-400 font-medium flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                Đang chạy
              </span>
              <span v-else class="text-slate-400 flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full bg-slate-500"></span>
                Dừng
              </span>
            </td>
            <td class="py-3 px-4 text-right space-x-1.5">
              <button
                v-if="app.status === 'running'"
                @click="actionApp(app.id, 'restart')"
                class="px-2 py-1 text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 rounded border border-slate-700 text-xs"
              >
                Restart
              </button>
              <button
                v-if="app.status === 'running'"
                @click="actionApp(app.id, 'stop')"
                class="px-2 py-1 text-amber-400 hover:text-amber-300 bg-amber-950/40 rounded border border-amber-900/60 text-xs"
              >
                Stop
              </button>
              <button
                v-else
                @click="actionApp(app.id, 'start')"
                class="px-2 py-1 text-emerald-400 hover:text-emerald-300 bg-emerald-950/40 rounded border border-emerald-900/60 text-xs"
              >
                Start
              </button>
              <button
                @click="deleteApp(app.id)"
                class="px-2 py-1 text-red-400 hover:text-red-300 bg-red-950/40 rounded border border-red-900/60 text-xs"
              >
                Xóa
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create App Modal -->
    <div v-if="isAddModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-lg w-full shadow-2xl overflow-hidden p-6 space-y-4">
        <h3 class="text-base font-semibold text-slate-100">Triển khai Ứng dụng mới</h3>

        <form @submit.prevent="submitCreateApp" class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-medium mb-1">Loại Runtime</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                @click="form.kind = 'node'"
                :class="form.kind === 'node' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'"
                class="py-2 rounded border border-slate-700 transition"
              >
                Node.js
              </button>
              <button
                type="button"
                @click="form.kind = 'python'"
                :class="form.kind === 'python' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'"
                class="py-2 rounded border border-slate-700 transition"
              >
                Python
              </button>
              <button
                type="button"
                @click="form.kind = 'docker'"
                :class="form.kind === 'docker' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'"
                class="py-2 rounded border border-slate-700 transition"
              >
                Docker Compose
              </button>
            </div>
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Tên ứng dụng</label>
            <input
              type="text"
              v-model="form.name"
              placeholder="ví dụ: my-api"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">File chạy chính (Entrypoint)</label>
            <input
              type="text"
              v-model="form.entrypoint"
              :placeholder="form.kind === 'node' ? 'server.js' : 'app.py'"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none font-mono text-[11px]"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Thư mục mã nguồn</label>
            <input
              type="text"
              v-model="form.working_directory"
              placeholder="Mặc định: /home/<user>/apps/<name>"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none font-mono text-[11px]"
            />
          </div>

          <div class="pt-4 border-t border-slate-700 flex justify-end gap-2">
            <button
              type="button"
              @click="isAddModalOpen = false"
              class="px-4 py-2 rounded border border-slate-700 text-slate-300 hover:bg-slate-800 transition"
            >
              Hủy
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium transition"
            >
              Tạo & Khởi động
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { appsApi } from '../../api/apps'
import { useAccountStore } from '../../stores/account'

const accountStore = useAccountStore()
const apps = ref([])
const loading = ref(false)
const isAddModalOpen = ref(false)

const form = ref({
  kind: 'node',
  name: '',
  entrypoint: 'server.js',
  working_directory: '',
})

async function loadData() {
  loading.value = true
  try {
    apps.value = await appsApi.list(accountStore.selectedAccountId)
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

function openAddAppModal() {
  form.value = {
    kind: 'node',
    name: '',
    entrypoint: 'server.js',
    working_directory: '',
  }
  isAddModalOpen.value = true
}

async function submitCreateApp() {
  if (!accountStore.currentAccount) {
    alert('Vui lòng chọn tài khoản hosting.')
    return
  }
  try {
    await appsApi.create({
      account_id: accountStore.currentAccount.id,
      kind: form.value.kind,
      name: form.value.name,
      entrypoint: form.value.entrypoint,
      working_directory: form.value.working_directory,
    })
    isAddModalOpen.value = false
    await loadData()
  } catch (err) {
    alert(`Lỗi tạo app: ${err.message}`)
  }
}

async function actionApp(id, act) {
  try {
    await appsApi.action(id, act)
    await loadData()
  } catch (err) {
    alert(`Thao tác thất bại: ${err.message}`)
  }
}

async function deleteApp(id) {
  if (!confirm('Bạn có chắc muốn xóa ứng dụng này?')) return
  try {
    await appsApi.delete(id)
    await loadData()
  } catch (err) {
    alert(`Xóa thất bại: ${err.message}`)
  }
}

watch(() => accountStore.selectedAccountId, () => {
  loadData()
})

onMounted(() => {
  loadData()
})
</script>
