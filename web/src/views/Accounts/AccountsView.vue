<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Quản lý Tài khoản Hosting</h1>
        <p class="text-xs text-slate-400 mt-1">Mỗi tài khoản tương ứng với 1 user Linux thực thể (/home/&lt;username&gt;), bảo đảm cô lập quyền</p>
      </div>

      <button
        @click="isAddModalOpen = true"
        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
      >
        + Tạo Tài khoản mới
      </button>
    </div>

    <!-- Table -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Tên tài khoản (Linux User)</th>
            <th class="py-3 px-4">Thư mục Home</th>
            <th class="py-3 px-4">Email</th>
            <th class="py-3 px-4">Dung lượng hạn mức</th>
            <th class="py-3 px-4">Trạng thái</th>
            <th class="py-3 px-4 text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="accountStore.loading" class="text-center text-slate-400">
            <td colspan="6" class="py-8">Đang tải danh sách tài khoản...</td>
          </tr>
          <tr v-else-if="accountStore.accounts.length === 0" class="text-center text-slate-400">
            <td colspan="6" class="py-8">Chưa có tài khoản nào. Bấm "Tạo Tài khoản mới".</td>
          </tr>
          <tr v-for="acc in accountStore.accounts" :key="acc.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-bold text-slate-100 font-mono">
              {{ acc.username }}
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-300">
              {{ acc.home_dir }}
            </td>
            <td class="py-3 px-4 text-slate-300">
              {{ acc.email || 'Không có' }}
            </td>
            <td class="py-3 px-4 text-slate-300">
              {{ acc.disk_quota_mb === 0 ? 'Không giới hạn' : `${acc.disk_quota_mb} MB` }}
            </td>
            <td class="py-3 px-4">
              <span class="text-emerald-400 font-medium" v-if="acc.status === 'active'">Hoạt động</span>
              <span class="text-slate-400" v-else>Khóa</span>
            </td>
            <td class="py-3 px-4 text-right">
              <button
                @click="deleteAccount(acc.id)"
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
        <h3 class="text-base font-semibold text-slate-100">Tạo Tài khoản Hosting mới</h3>

        <form @submit.prevent="submitAdd" class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-medium mb-1">Tên tài khoản (Linux username)</label>
            <input
              type="text"
              v-model="newAcc.username"
              placeholder="ví dụ: webmaster"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Email liên hệ</label>
            <input
              type="email"
              v-model="newAcc.email"
              placeholder="user@example.com"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Hạn mức Disk Quota (MB, 0 = không giới hạn)</label>
            <input
              type="number"
              v-model.number="newAcc.disk_quota_mb"
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
              Khởi tạo
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { accountsApi } from '../../api/accounts'
import { useAccountStore } from '../../stores/account'

const accountStore = useAccountStore()
const isAddModalOpen = ref(false)

const newAcc = ref({
  username: '',
  email: '',
  disk_quota_mb: 0,
})

async function submitAdd() {
  try {
    await accountsApi.create({
      username: newAcc.value.username,
      email: newAcc.value.email ? newAcc.value.email : null,
      disk_quota_mb: newAcc.value.disk_quota_mb,
    })
    isAddModalOpen.value = false
    newAcc.value = { username: '', email: '', disk_quota_mb: 0 }
    await accountStore.fetchAccounts()
  } catch (err) {
    alert(`Lỗi: ${err.message}`)
  }
}

async function deleteAccount(id) {
  if (!confirm('Bạn có chắc muốn xóa tài khoản này? Toàn bộ website, file và database của tài khoản sẽ bị xóa theo!')) return
  try {
    await accountsApi.delete(id)
    await accountStore.fetchAccounts()
  } catch (err) {
    alert(`Xóa thất bại: ${err.message}`)
  }
}

onMounted(() => {
  accountStore.fetchAccounts()
})
</script>
