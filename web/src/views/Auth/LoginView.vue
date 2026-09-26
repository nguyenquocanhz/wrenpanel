<template>
  <div class="min-h-screen flex items-center justify-center bg-[#0b1120] text-slate-100 p-4">
    <div class="max-w-md w-full bg-[#1e293b] border border-slate-800 rounded-xl shadow-2xl p-8 space-y-6">
      <!-- Logo / Title -->
      <div class="text-center space-y-2">
        <div class="w-12 h-12 bg-blue-600 rounded-lg mx-auto flex items-center justify-center text-xl font-bold shadow-lg shadow-blue-500/30">
          W
        </div>
        <h1 class="text-2xl font-bold text-slate-100 tracking-tight">WrenPanel</h1>
        <p class="text-xs text-slate-400">Đăng nhập Quản trị Server & Hosting Accounts</p>
      </div>

      <!-- Alert Error -->
      <div v-if="errorMessage" class="bg-red-950/40 border border-red-800/60 text-red-300 text-xs p-3 rounded-lg flex items-center gap-2">
        <span>⚠️</span>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-4 text-xs">
        <div>
          <label class="block text-slate-300 font-medium mb-1.5">Tên đăng nhập</label>
          <input
            type="text"
            v-model="username"
            required
            placeholder="admin hoặc linux username (ví dụ: webmaster)"
            class="w-full bg-[#0f172a] border border-slate-700 rounded-lg p-2.5 text-slate-200 focus:border-blue-500 focus:outline-none font-mono transition"
          />
        </div>

        <div>
          <div class="flex items-center justify-between mb-1.5">
            <label class="text-slate-300 font-medium">Mật khẩu</label>
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="text-[11px] text-blue-400 hover:text-blue-300"
            >
              {{ showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu' }}
            </button>
          </div>
          <input
            :type="showPassword ? 'text' : 'password'"
            v-model="password"
            required
            placeholder="••••••••••••"
            class="w-full bg-[#0f172a] border border-slate-700 rounded-lg p-2.5 text-slate-200 focus:border-blue-500 focus:outline-none font-mono transition"
          />
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold rounded-lg shadow-lg shadow-blue-600/30 transition text-xs mt-2"
        >
          <span v-if="loading">Đang xác thực...</span>
          <span v-else>Đăng nhập vào Hệ thống</span>
        </button>
      </form>

      <!-- Explanatory note -->
      <div class="border-t border-slate-800/80 pt-4 space-y-2 text-[11px] text-slate-400">
        <div class="font-medium text-slate-300">💡 Hướng dẫn tài khoản:</div>
        <div>• <strong>Admin (Quản trị viên):</strong> Mặc định <code class="text-amber-300">admin</code> / <code class="text-amber-300">admin123456</code> (toàn quyền server).</div>
        <div>• <strong>Hosting Account:</strong> Sử dụng username và mật khẩu được tạo trong mục <em>Quản lý Tài khoản Hosting</em> (hoặc mật khẩu SSH/SFTP).</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../../api/auth'

const router = useRouter()
const username = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const errorMessage = ref('')

async function handleLogin() {
  loading.value = true
  errorMessage.value = ''
  try {
    const res = await authApi.login(username.value, password.value)
    localStorage.setItem('wrenpanel_token', res.token)
    localStorage.setItem('wrenpanel_user', JSON.stringify(res.user))
    router.push('/vhosts')
  } catch (err) {
    errorMessage.value = err.message || 'Đăng nhập không thành công'
  } finally {
    loading.value = false
  }
}
</script>
