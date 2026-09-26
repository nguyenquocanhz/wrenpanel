<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Quản lý Tài khoản FTP / SFTP</h1>
        <p class="text-xs text-slate-400 mt-1">Tạo tài khoản FTP ảo cô lập chroot trong thư mục website, cổng kết nối tiêu chuẩn</p>
      </div>

      <button
        @click="isAddModalOpen = true"
        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
      >
        + Tạo tài khoản FTP
      </button>
    </div>

    <!-- Table of FTP accounts -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Tài khoản FTP</th>
            <th class="py-3 px-4">Thư mục gốc (Root Directory)</th>
            <th class="py-3 px-4">Trạng thái</th>
            <th class="py-3 px-4">Ngày tạo</th>
            <th class="py-3 px-4 text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="loading" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Đang tải danh sách FTP...</td>
          </tr>
          <tr v-else-if="ftpList.length === 0" class="text-center text-slate-400">
            <td colspan="5" class="py-8">Chưa có tài khoản FTP nào. Bấm "Tạo tài khoản FTP" để bắt đầu.</td>
          </tr>
          <tr v-for="f in ftpList" :key="f.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-bold text-slate-100 font-mono flex items-center gap-2">
              <svg class="w-4 h-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
              </svg>
              <span>{{ f.username }}</span>
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-300">
              {{ f.root_dir }}
            </td>
            <td class="py-3 px-4">
              <span class="text-emerald-400 font-medium" v-if="f.status === 'active'">Hoạt động</span>
              <span class="text-slate-400" v-else>Khóa</span>
            </td>
            <td class="py-3 px-4 text-slate-400">
              {{ f.created_at }}
            </td>
            <td class="py-3 px-4 text-right">
              <button
                @click="deleteFtp(f.id)"
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
        <h3 class="text-base font-semibold text-slate-100">Tạo Tài khoản FTP Mới</h3>

        <form @submit.prevent="submitAdd" class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-medium mb-1">Tên tài khoản (Hệ thống tự thêm tiền tố username_)</label>
            <input
              type="text"
              v-model="newFtp.username"
              placeholder="ví dụ: dev"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Mật khẩu FTP</label>
            <input
              type="password"
              v-model="newFtp.password"
              placeholder="••••••••"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Thư mục giới hạn truy cập (Chroot Root Directory)</label>
            <input
              type="text"
              v-model="newFtp.root_dir"
              placeholder="Mặc định: /home/<user>/public_html"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none font-mono text-[11px]"
            />
            <p class="text-[11px] text-slate-400 mt-1">User chỉ có quyền xem và sửa file trong thư mục này trở xuống.</p>
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
              Tạo tài khoản FTP
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ftpApi } from '../../api/ftp'
import { useAccountStore } from '../../stores/account'

const accountStore = useAccountStore()
const ftpList = ref([])
const loading = ref(false)
const isAddModalOpen = ref(false)

const newFtp = ref({
  username: '',
  password: '',
  root_dir: '',
})

async function loadData() {
  loading.value = true
  try {
    ftpList.value = await ftpApi.list(accountStore.selectedAccountId)
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
    await ftpApi.create({
      account_id: accountStore.currentAccount.id,
      username: newFtp.value.username,
      password: newFtp.value.password,
      root_dir: newFtp.value.root_dir,
    })
    isAddModalOpen.value = false
    newFtp.value = { username: '', password: '', root_dir: '' }
    await loadData()
  } catch (err) {
    alert(`Lỗi: ${err.message}`)
  }
}

async function deleteFtp(id) {
  if (!confirm('Bạn có chắc muốn xóa tài khoản FTP này?')) return
  try {
    await ftpApi.delete(id)
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
