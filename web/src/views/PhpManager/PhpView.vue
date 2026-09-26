<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Quản lý PHP (Side-by-side)</h1>
        <p class="text-xs text-slate-400 mt-1">Cài đặt song song nhiều phiên bản PHP, mỗi phiên bản có dịch vụ PHP-FPM độc lập</p>
      </div>

      <button
        @click="isAddModalOpen = true"
        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
      >
        + Thêm phiên bản PHP
      </button>
    </div>

    <!-- Table of PHP Versions -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Phiên bản</th>
            <th class="py-3 px-4">Dịch vụ PHP-FPM</th>
            <th class="py-3 px-4">Đường dẫn Binary</th>
            <th class="py-3 px-4">Ngày cấu hình</th>
            <th class="py-3 px-4 text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="loading" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Đang tải danh sách PHP...</td>
          </tr>
          <tr v-else-if="versions.length === 0" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Chưa có phiên bản PHP nào. Bấm "Thêm phiên bản PHP".</td>
          </tr>
          <tr v-for="p in versions" :key="p.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-bold text-blue-400">
              PHP {{ p.version }}
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-300">
              {{ p.fpm_service }}
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-300">
              {{ p.binary_path }}
            </td>
            <td class="py-3 px-4 text-slate-400">
              {{ p.installed_at }}
            </td>
            <td class="py-3 px-4 text-right space-x-2">
              <button
                @click="reloadPhp(p.version)"
                class="px-2.5 py-1 text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 rounded border border-slate-700 transition text-xs"
              >
                Reload FPM
              </button>
              <button
                @click="deletePhp(p.id)"
                class="px-2.5 py-1 text-red-400 hover:text-red-300 hover:bg-red-950/40 rounded border border-red-900/60 transition text-xs"
              >
                Gỡ bỏ
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add PHP Modal -->
    <div v-if="isAddModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-md w-full shadow-2xl overflow-hidden p-6 space-y-4">
        <h3 class="text-base font-semibold text-slate-100">Đăng ký phiên bản PHP mới</h3>

        <form @submit.prevent="submitAddPhp" class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-medium mb-1">Phiên bản (ví dụ: 8.3)</label>
            <input
              type="text"
              v-model="newPhp.version"
              placeholder="8.3"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Tên systemd service (tùy chọn)</label>
            <input
              type="text"
              v-model="newPhp.fpm_service"
              placeholder="Mặc định: php8.3-fpm"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Đường dẫn binary (tùy chọn)</label>
            <input
              type="text"
              v-model="newPhp.binary_path"
              placeholder="Mặc định: /usr/bin/php8.3"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
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
              Lưu
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { phpApi } from '../../api/php'

const versions = ref([])
const loading = ref(false)
const isAddModalOpen = ref(false)

const newPhp = ref({
  version: '',
  fpm_service: '',
  binary_path: '',
})

async function loadData() {
  loading.value = true
  try {
    versions.value = await phpApi.listVersions()
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function submitAddPhp() {
  try {
    await phpApi.createVersion(newPhp.value)
    isAddModalOpen.value = false
    newPhp.value = { version: '', fpm_service: '', binary_path: '' }
    await loadData()
  } catch (err) {
    alert(`Lỗi: ${err.message}`)
  }
}

async function reloadPhp(ver) {
  try {
    await phpApi.reload(ver)
    alert(`Dịch vụ PHP ${ver} đã được reload.`)
  } catch (err) {
    alert(`Lỗi reload: ${err.message}`)
  }
}

async function deletePhp(id) {
  if (!confirm('Bạn có chắc muốn gỡ phiên bản PHP này?')) return
  try {
    await phpApi.deleteVersion(id)
    await loadData()
  } catch (err) {
    alert(`Không thể gỡ bỏ: ${err.message}`)
  }
}

onMounted(() => {
  loadData()
})
</script>
