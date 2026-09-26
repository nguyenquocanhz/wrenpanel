<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Quản lý Cơ sở Dữ liệu</h1>
        <p class="text-xs text-slate-400 mt-1">Cấp phát database MySQL, PostgreSQL và MongoDB theo tài khoản/website</p>
      </div>

      <div class="flex items-center gap-2">
        <button
          @click="openPhpMyAdmin"
          class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
          </svg>
          Truy cập phpMyAdmin (SSO)
        </button>

        <button
          @click="isAddModalOpen = true"
          class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
        >
          + Tạo Cơ sở Dữ liệu
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Tên Database</th>
            <th class="py-3 px-4">Engine</th>
            <th class="py-3 px-4">User quản trị</th>
            <th class="py-3 px-4">Ngày tạo</th>
            <th class="py-3 px-4 text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="loading" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Đang tải danh sách cơ sở dữ liệu...</td>
          </tr>
          <tr v-else-if="databases.length === 0" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Chưa có database nào. Bấm "Tạo Cơ sở Dữ liệu".</td>
          </tr>
          <tr v-for="db in databases" :key="db.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-bold text-slate-100 font-mono">
              {{ db.db_name }}
            </td>
            <td class="py-3 px-4">
              <span class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase bg-slate-800 text-slate-300 border border-slate-700">
                {{ db.engine }}
              </span>
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-300">
              {{ db.db_user }}
            </td>
            <td class="py-3 px-4 text-slate-400">
              {{ db.created_at }}
            </td>
            <td class="py-3 px-4 text-right">
              <button
                @click="deleteDb(db.id)"
                class="px-2.5 py-1 text-red-400 hover:text-red-300 bg-red-950/40 rounded border border-red-900/60 text-xs"
              >
                Xóa
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <div v-if="isAddModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-md w-full shadow-2xl p-6 space-y-4">
        <h3 class="text-base font-semibold text-slate-100">Tạo Database Mới</h3>

        <form @submit.prevent="submitAdd" class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-medium mb-1">Hệ quản trị (Engine)</label>
            <select
              v-model="newDb.engine"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            >
              <option value="mysql">MySQL / MariaDB</option>
              <option value="postgresql">PostgreSQL</option>
              <option value="mongodb">MongoDB</option>
            </select>
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Tên Database (Hệ thống tự thêm tiền tố username_)</label>
            <input
              type="text"
              v-model="newDb.db_name"
              placeholder="ví dụ: blog"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Tên User Database</label>
            <input
              type="text"
              v-model="newDb.db_user"
              placeholder="ví dụ: blog_user"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Mật khẩu User</label>
            <input
              type="password"
              v-model="newDb.db_password"
              placeholder="••••••••"
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
              Tạo Database
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { databasesApi } from '../../api/databases'
import { useAccountStore } from '../../stores/account'

const accountStore = useAccountStore()
const databases = ref([])
const loading = ref(false)
const isAddModalOpen = ref(false)

const newDb = ref({
  engine: 'mysql',
  db_name: '',
  db_user: '',
  db_password: '',
})

async function openPhpMyAdmin() {
  try {
    const res = await fetch('/api/db-manager/launch', { method: 'POST' })
    const data = await res.json()
    alert(`Đã khởi tạo phiên phpMyAdmin SSO an toàn:\n${data.launch_url}\n(Chỉ có thể truy cập qua phiên xác thực nội bộ của WrenPanel)`)
  } catch (err) {
    alert(`Lỗi mở phpMyAdmin: ${err.message}`)
  }
}

async function loadData() {
  loading.value = true
  try {
    databases.value = await databasesApi.list(accountStore.selectedAccountId)
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function submitAdd() {
  if (!accountStore.currentAccount) {
    alert('Vui lòng chọn tài khoản hosting.')
    return
  }
  try {
    await databasesApi.create({
      account_id: accountStore.currentAccount.id,
      engine: newDb.value.engine,
      db_name: newDb.value.db_name,
      db_user: newDb.value.db_user,
      db_password: newDb.value.db_password,
    })
    isAddModalOpen.value = false
    newDb.value = { engine: 'mysql', db_name: '', db_user: '', db_password: '' }
    await loadData()
  } catch (err) {
    alert(`Lỗi: ${err.message}`)
  }
}

async function deleteDb(id) {
  if (!confirm('Bạn có chắc muốn xóa database này? Dữ liệu bảng sẽ mất vĩnh viễn!')) return
  try {
    await databasesApi.delete(id)
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
