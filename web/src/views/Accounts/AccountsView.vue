<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Quản lý Tài khoản Hosting</h1>
        <p class="text-xs text-slate-400 mt-1">Mỗi tài khoản tương ứng với 1 user Linux thực thể (/home/&lt;username&gt;), bảo đảm cô lập quyền</p>
      </div>

      <button
        @click="openAddModal"
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
            <td class="py-3 px-4 text-right space-x-2">
              <button
                @click="openResetPasswordModal(acc)"
                class="px-2.5 py-1 text-amber-400 hover:text-amber-300 bg-amber-950/40 rounded border border-amber-900/60 text-xs transition"
                title="Đổi mật khẩu Linux, FTP, Panel"
              >
                🔑 Đổi mật khẩu
              </button>
              <button
                @click="deleteAccount(acc.id)"
                class="px-2.5 py-1 text-red-400 hover:text-red-300 bg-red-950/40 rounded border border-red-900/60 text-xs transition"
              >
                Xóa
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal Tạo tài khoản mới -->
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
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none font-mono"
            />
          </div>

          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="text-slate-300 font-medium">Mật khẩu đăng nhập</label>
              <button
                type="button"
                @click="generateAddPassword"
                class="text-blue-400 hover:text-blue-300 text-[11px]"
              >
                🎲 Sinh ngẫu nhiên
              </button>
            </div>
            <div class="relative">
              <input
                :type="showPassword ? 'text' : 'password'"
                v-model="newAcc.password"
                placeholder="Để trống để tự động sinh mật khẩu ngẫu nhiên"
                class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none font-mono pr-14"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-2 top-2 text-[11px] text-slate-400 hover:text-slate-200"
              >
                {{ showPassword ? 'Ẩn' : 'Hiện' }}
              </button>
            </div>
            <p class="text-[10px] text-slate-400 mt-1">
              Mật khẩu này áp dụng cho Đăng nhập Panel, SSH/SFTP và tài khoản FTP chính của user.
            </p>
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

    <!-- Modal Thông tin Tài khoản vừa tạo thành công -->
    <div v-if="isCredentialsModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-md w-full shadow-2xl p-6 space-y-4">
        <div class="flex items-center gap-2 text-emerald-400">
          <span class="text-xl">✅</span>
          <h3 class="text-base font-semibold text-slate-100">Tài khoản đã tạo thành công!</h3>
        </div>

        <p class="text-xs text-slate-300">
          Vui lòng sao lưu lại thông tin đăng nhập dưới đây và gửi cho người dùng:
        </p>

        <div class="bg-[#0f172a] border border-slate-800 rounded-lg p-3 space-y-2 text-xs font-mono">
          <div class="flex justify-between items-center py-1 border-b border-slate-800">
            <span class="text-slate-400">Tên tài khoản:</span>
            <span class="text-slate-100 font-bold">{{ createdInfo.username }}</span>
          </div>

          <div class="flex justify-between items-center py-1 border-b border-slate-800">
            <span class="text-slate-400">Mật khẩu:</span>
            <div class="flex items-center gap-2">
              <span class="text-amber-400 font-bold">{{ createdInfo.password }}</span>
              <button
                @click="copyText(createdInfo.password)"
                class="px-2 py-0.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded text-[10px]"
              >
                Copy
              </button>
            </div>
          </div>

          <div class="flex justify-between items-center py-1">
            <span class="text-slate-400">Thư mục gốc:</span>
            <span class="text-slate-300 text-[11px]">{{ createdInfo.home_dir }}</span>
          </div>
        </div>

        <div class="bg-blue-950/30 border border-blue-900/50 rounded p-3 text-[11px] text-blue-300 space-y-1">
          <div class="font-semibold text-blue-200">🚀 Thông tin kết nối:</div>
          <div>• <strong>Panel:</strong> Đăng nhập tại màn hình đăng nhập WrenPanel</div>
          <div>• <strong>SFTP / SSH:</strong> Cổng 22, User: <code class="bg-slate-900 px-1 py-0.5 rounded">{{ createdInfo.username }}</code></div>
          <div>• <strong>FTP:</strong> Host máy chủ, Cổng 21, User: <code class="bg-slate-900 px-1 py-0.5 rounded">{{ createdInfo.username }}</code></div>
        </div>

        <div class="flex justify-end pt-2">
          <button
            @click="isCredentialsModalOpen = false"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium text-xs transition"
          >
            Đã lưu, Đóng lại
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Đổi / Reset mật khẩu -->
    <div v-if="isResetModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-md w-full shadow-2xl p-6 space-y-4">
        <h3 class="text-base font-semibold text-slate-100">
          Đổi mật khẩu cho: <span class="text-blue-400 font-mono">{{ resetTarget.username }}</span>
        </h3>

        <form @submit.prevent="submitResetPassword" class="space-y-3 text-xs">
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="text-slate-300 font-medium">Mật khẩu mới</label>
              <button
                type="button"
                @click="generateResetPassword"
                class="text-blue-400 hover:text-blue-300 text-[11px]"
              >
                🎲 Sinh ngẫu nhiên
              </button>
            </div>
            <input
              type="text"
              v-model="resetPasswordValue"
              placeholder="Nhập mật khẩu mới hoặc click Sinh ngẫu nhiên"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none font-mono"
            />
            <p class="text-[10px] text-slate-400 mt-1">
              Cập nhật đồng bộ mật khẩu đăng nhập Panel, Linux chpasswd (SSH/SFTP) và FTP.
            </p>
          </div>

          <div class="pt-4 border-t border-slate-700 flex justify-end gap-2">
            <button
              type="button"
              @click="isResetModalOpen = false"
              class="px-4 py-2 rounded border border-slate-700 text-slate-300 hover:bg-slate-800 transition"
            >
              Hủy
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white rounded font-medium transition"
            >
              Cập nhật mật khẩu
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
const isCredentialsModalOpen = ref(false)
const isResetModalOpen = ref(false)
const showPassword = ref(false)

const newAcc = ref({
  username: '',
  password: '',
  email: '',
  disk_quota_mb: 0,
})

const createdInfo = ref({
  username: '',
  password: '',
  home_dir: '',
})

const resetTarget = ref({
  id: null,
  username: '',
})
const resetPasswordValue = ref('')

function generateRandomString(length = 16) {
  const chars = 'abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%^&*'
  let result = ''
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

function openAddModal() {
  newAcc.value = {
    username: '',
    password: '',
    email: '',
    disk_quota_mb: 0,
  }
  showPassword.value = false
  isAddModalOpen.value = true
}

function generateAddPassword() {
  newAcc.value.password = generateRandomString(16)
  showPassword.value = true
}

function openResetPasswordModal(acc) {
  resetTarget.value = { id: acc.id, username: acc.username }
  resetPasswordValue.value = generateRandomString(16)
  isResetModalOpen.value = true
}

function generateResetPassword() {
  resetPasswordValue.value = generateRandomString(16)
}

function copyText(text) {
  navigator.clipboard.writeText(text)
  alert('Đã sao chép vào bộ nhớ tạm!')
}

async function submitAdd() {
  try {
    const res = await accountsApi.create({
      username: newAcc.value.username,
      password: newAcc.value.password ? newAcc.value.password : '',
      email: newAcc.value.email ? newAcc.value.email : null,
      disk_quota_mb: newAcc.value.disk_quota_mb,
    })

    isAddModalOpen.value = false
    createdInfo.value = {
      username: res.username,
      password: res.initial_password,
      home_dir: res.home_dir,
    }
    isCredentialsModalOpen.value = true
    await accountStore.fetchAccounts()
  } catch (err) {
    alert(`Lỗi: ${err.message}`)
  }
}

async function submitResetPassword() {
  try {
    const res = await accountsApi.resetPassword(resetTarget.value.id, resetPasswordValue.value)
    isResetModalOpen.value = false
    alert(`Mật khẩu tài khoản "${res.username}" đã đổi thành công:\n\nMật khẩu mới: ${res.new_password}`)
  } catch (err) {
    alert(`Đổi mật khẩu thất bại: ${err.message}`)
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
