<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Quản lý Chứng chỉ SSL / TLS</h1>
        <p class="text-xs text-slate-400 mt-1">Tích hợp thư viện ACME Lego Go, tự động xác thực HTTP-01 và cấp chứng chỉ Let's Encrypt</p>
      </div>

      <button
        @click="isIssueModalOpen = true"
        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
      >
        + Xin cấp chứng chỉ mới
      </button>
    </div>

    <!-- Table of Certs -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Tên miền (FQDN)</th>
            <th class="py-3 px-4">Đơn vị cấp (Issuer)</th>
            <th class="py-3 px-4">File Chứng chỉ</th>
            <th class="py-3 px-4">Ngày cấp</th>
            <th class="py-3 px-4">Ngày hết hạn</th>
            <th class="py-3 px-4">Tự động gia hạn</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="loading" class="text-center text-slate-400">
            <td colspan="6" class="py-8">Đang tải danh sách chứng chỉ...</td>
          </tr>
          <tr v-else-if="certs.length === 0" class="text-center text-slate-400">
            <td colspan="6" class="py-8">Chưa có chứng chỉ SSL nào được cài đặt.</td>
          </tr>
          <tr v-for="c in certs" :key="c.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-bold text-slate-100 flex items-center gap-2">
              <svg class="w-4 h-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
              <span>{{ c.fqdn }}</span>
            </td>
            <td class="py-3 px-4">
              <span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-900/60 text-emerald-300 border border-emerald-700">
                {{ c.issuer }}
              </span>
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-400 max-w-xs truncate" :title="c.cert_path">
              {{ c.cert_path }}
            </td>
            <td class="py-3 px-4 text-slate-300">
              {{ c.issued_at || 'N/A' }}
            </td>
            <td class="py-3 px-4 text-amber-300 font-mono">
              {{ c.expires_at || 'N/A' }}
            </td>
            <td class="py-3 px-4">
              <span class="text-emerald-400 font-medium">Bật (Auto 60 ngày)</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Issue Modal -->
    <div v-if="isIssueModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-md w-full shadow-2xl p-6 space-y-4">
        <h3 class="text-base font-semibold text-slate-100">Xin cấp chứng chỉ SSL Let's Encrypt</h3>
        <p class="text-xs text-slate-400">
          Hệ thống sẽ kiểm tra bản ghi DNS A của tên miền trỏ về IP máy chủ trước khi gửi yêu cầu HTTP-01 đến Let's Encrypt.
        </p>

        <form @submit.prevent="submitIssue" class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-medium mb-1">Chọn Website cần cấp SSL</label>
            <select
              v-model="issueForm.vhost_id"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            >
              <option :value="null">-- Chọn website --</option>
              <option v-for="v in vhosts" :key="v.id" :value="v.id">
                {{ v.fqdn }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-slate-300 font-medium mb-1">Email đăng ký ACME</label>
            <input
              type="email"
              v-model="issueForm.email"
              placeholder="admin@example.com"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div class="pt-4 border-t border-slate-700 flex justify-end gap-2">
            <button
              type="button"
              @click="isIssueModalOpen = false"
              class="px-4 py-2 rounded border border-slate-700 text-slate-300 hover:bg-slate-800 transition"
            >
              Hủy
            </button>
            <button
              type="submit"
              :disabled="issuing"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium transition disabled:opacity-50"
            >
              <span v-if="issuing">Đang xác thực DNS & cấp cert...</span>
              <span v-else>Cấp chứng chỉ</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { sslApi } from '../../api/ssl'
import { vhostsApi } from '../../api/vhosts'

const certs = ref([])
const vhosts = ref([])
const loading = ref(false)
const issuing = ref(false)
const isIssueModalOpen = ref(false)

const issueForm = ref({
  vhost_id: null,
  email: 'admin@wrenpanel.local',
})

async function loadData() {
  loading.value = true
  try {
    const [cList, vList] = await Promise.all([
      sslApi.listCerts(),
      vhostsApi.list(),
    ])
    certs.value = cList || []
    vhosts.value = vList || []
    if (vhosts.value.length > 0 && !issueForm.value.vhost_id) {
      issueForm.value.vhost_id = vhosts.value[0].id
    }
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function submitIssue() {
  if (!issueForm.value.vhost_id) return
  issuing.value = true
  try {
    await sslApi.issueCert(issueForm.value.vhost_id, issueForm.value.email)
    isIssueModalOpen.value = false
    alert('Chứng chỉ SSL Let\'s Encrypt đã được cấp và kích hoạt trên Nginx thành công!')
    await loadData()
  } catch (err) {
    alert(`Lỗi cấp SSL: ${err.message}`)
  } finally {
    issuing.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
