<template>
  <div class="space-y-6">
    <!-- Top Action Bar -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-100">Danh sách Website</h1>
        <p class="text-xs text-slate-400 mt-1">Quản lý VirtualHost, Web Engine Nginx/Apache và cấu hình SSL</p>
      </div>

      <button
        @click="openAddModal"
        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold px-4 py-2 rounded flex items-center gap-2 shadow transition"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Thêm Website / Subdomain
      </button>
    </div>

    <!-- Table-first view (DirectAdmin / aaPanel style) -->
    <div class="bg-[#1e293b] border border-slate-800 rounded-lg shadow overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
          <tr>
            <th class="py-3 px-4">Tên miền (FQDN)</th>
            <th class="py-3 px-4">Loại</th>
            <th class="py-3 px-4">Thư mục gốc (Docroot)</th>
            <th class="py-3 px-4">Engine</th>
            <th class="py-3 px-4">PHP</th>
            <th class="py-3 px-4">SSL</th>
            <th class="py-3 px-4">Trạng thái</th>
            <th class="py-3 px-4 text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60 text-slate-200">
          <tr v-if="loading" class="text-center text-slate-400">
            <td colspan="8" class="py-8">Đang tải danh sách vhost...</td>
          </tr>
          <tr v-else-if="vhosts.length === 0" class="text-center text-slate-400">
            <td colspan="8" class="py-8">Chưa có website nào. Bấm "Thêm Website" để tạo mới.</td>
          </tr>
          <tr v-for="v in vhosts" :key="v.id" class="hover:bg-slate-800/40 transition">
            <td class="py-3 px-4 font-medium text-slate-100 flex items-center gap-2">
              <span class="w-2 h-2 rounded-full" :class="v.status === 'active' ? 'bg-emerald-400' : 'bg-slate-500'"></span>
              <span>{{ v.fqdn }}</span>
            </td>
            <td class="py-3 px-4">
              <span
                class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase"
                :class="{
                  'bg-blue-900/60 text-blue-300 border border-blue-700': v.type === 'primary',
                  'bg-purple-900/60 text-purple-300 border border-purple-700': v.type === 'subdomain',
                  'bg-amber-900/60 text-amber-300 border border-amber-700': v.type === 'addon',
                }"
              >
                {{ v.type }}
              </span>
            </td>
            <td class="py-3 px-4 font-mono text-[11px] text-slate-300 max-w-xs truncate" :title="v.docroot">
              {{ v.docroot }}
            </td>
            <td class="py-3 px-4">
              <span class="text-slate-300 font-medium">
                {{ v.web_engine === 'nginx+apache' ? 'Nginx + Apache' : 'Nginx thuần' }}
              </span>
            </td>
            <td class="py-3 px-4 text-slate-300">
              {{ getPhpLabel(v.php_version_id) }}
            </td>
            <td class="py-3 px-4">
              <span v-if="v.ssl_cert_id" class="text-emerald-400 flex items-center gap-1 font-medium">
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
                SSL Bật
              </span>
              <span v-else class="text-slate-400">Chưa có</span>
            </td>
            <td class="py-3 px-4">
              <span class="text-emerald-400 font-medium" v-if="v.status === 'active'">Hoạt động</span>
              <span class="text-slate-400" v-else>Tạm dừng</span>
            </td>
            <td class="py-3 px-4 text-right">
              <button
                @click="promptDelete(v)"
                class="px-2.5 py-1 text-red-400 hover:text-red-300 hover:bg-red-950/40 rounded border border-red-900/60 transition text-xs"
              >
                Xóa
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create Vhost Modal -->
    <div v-if="isAddModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-lg w-full shadow-2xl overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-700 flex items-center justify-between">
          <h3 class="text-base font-semibold text-slate-100">Thêm Website / Tên miền mới</h3>
          <button @click="isAddModalOpen = false" class="text-slate-400 hover:text-white">&times;</button>
        </div>

        <form @submit.prevent="submitCreateVhost" class="p-6 space-y-4 text-xs">
          <!-- Type -->
          <div>
            <label class="block text-slate-300 font-medium mb-1">Loại tên miền</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                @click="form.type = 'primary'"
                :class="form.type === 'primary' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'"
                class="py-2 rounded border border-slate-700 transition text-center"
              >
                Domain chính
              </button>
              <button
                type="button"
                @click="form.type = 'subdomain'"
                :class="form.type === 'subdomain' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'"
                class="py-2 rounded border border-slate-700 transition text-center"
              >
                Subdomain
              </button>
              <button
                type="button"
                @click="form.type = 'addon'"
                :class="form.type === 'addon' ? 'bg-blue-600 text-white font-semibold' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'"
                class="py-2 rounded border border-slate-700 transition text-center"
              >
                Addon domain
              </button>
            </div>
          </div>

          <!-- Parent domain (if subdomain) -->
          <div v-if="form.type === 'subdomain'">
            <label class="block text-slate-300 font-medium mb-1">Thuộc domain chính nào</label>
            <select
              v-model="form.parent_vhost_id"
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            >
              <option :value="null">-- Chọn domain chính --</option>
              <option v-for="p in primaryVhosts" :key="p.id" :value="p.id">{{ p.fqdn }}</option>
            </select>
          </div>

          <!-- FQDN -->
          <div>
            <label class="block text-slate-300 font-medium mb-1">Tên miền (FQDN)</label>
            <input
              type="text"
              v-model="form.fqdn"
              placeholder="ví dụ: example.com hoặc blog.example.com"
              required
              class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
            />
            <p class="text-[11px] text-slate-400 mt-1">Các prefix hệ thống bị cấm: mail, webmail, wrenpanel, ftp, ns1, ns2...</p>
          </div>

          <!-- Docroot Selection (Mandated by GEMINI.md) -->
          <div>
            <label class="block text-slate-300 font-medium mb-1">Thư mục mã nguồn (Docroot)</label>
            <div class="space-y-2">
              <label class="flex items-center gap-2 cursor-pointer">
                <input type="radio" value="new" v-model="docrootMode" class="text-blue-600 bg-slate-900 border-slate-700">
                <span class="text-slate-300">Tạo thư mục mới riêng biệt (Khuyên dùng)</span>
              </label>

              <label class="flex items-center gap-2 cursor-pointer">
                <input type="radio" value="existing" v-model="docrootMode" class="text-blue-600 bg-slate-900 border-slate-700">
                <span class="text-slate-300">Dùng chung thư mục của website đã có (Alias / Shared)</span>
              </label>
            </div>

            <!-- Existing docroot selector -->
            <div v-if="docrootMode === 'existing'" class="mt-2">
              <select
                v-model="form.docroot_owner_id"
                @change="onDocrootOwnerChange"
                class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
              >
                <option :value="null">-- Chọn website nguồn --</option>
                <option v-for="item in vhosts" :key="item.id" :value="item.id">
                  {{ item.fqdn }} ({{ item.docroot }})
                </option>
              </select>
            </div>

            <div v-else class="mt-2">
              <input
                type="text"
                v-model="form.docroot"
                placeholder="Mặc định: /home/<user>/public_html/..."
                class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-300 focus:border-blue-500 focus:outline-none font-mono text-[11px]"
              />
            </div>
          </div>

          <!-- Web Engine & PHP -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-300 font-medium mb-1">Web Engine</label>
              <select
                v-model="form.web_engine"
                class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
              >
                <option value="nginx">Nginx thuần (Nhanh nhất)</option>
                <option value="nginx+apache">Nginx + Apache (Hỗ trợ .htaccess)</option>
              </select>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">PHP Version</label>
              <select
                v-model="form.php_version_id"
                class="w-full bg-[#0f172a] border border-slate-700 rounded p-2 text-slate-200 focus:border-blue-500 focus:outline-none"
              >
                <option :value="null">Không dùng PHP</option>
                <option v-for="p in phpVersions" :key="p.id" :value="p.id">PHP {{ p.version }}</option>
              </select>
            </div>
          </div>

          <div v-if="errorMessage" class="p-3 rounded bg-red-950/60 border border-red-800 text-red-300">
            {{ errorMessage }}
          </div>

          <!-- Modal Footer -->
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
              :disabled="submitting"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium transition disabled:opacity-50"
            >
              <span v-if="submitting">Đang cấu hình vhost...</span>
              <span v-else>Tạo Website</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Safe Deletion Impact Confirmation Modal -->
    <DeleteConfirmModal
      :is-open="isDeleteModalOpen"
      :impact="deletionImpact"
      :loading="deleting"
      @close="isDeleteModalOpen = false"
      @confirm="executeDelete"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { vhostsApi } from '../../api/vhosts'
import { phpApi } from '../../api/php'
import { useAccountStore } from '../../stores/account'
import DeleteConfirmModal from '../../components/DeleteConfirmModal.vue'

const accountStore = useAccountStore()

const vhosts = ref([])
const phpVersions = ref([])
const loading = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const errorMessage = ref('')

const isAddModalOpen = ref(false)
const docrootMode = ref('new')

const form = ref({
  type: 'primary',
  parent_vhost_id: null,
  fqdn: '',
  docroot: '',
  docroot_owner_id: null,
  web_engine: 'nginx',
  php_version_id: null,
})

// Deletion modal state
const isDeleteModalOpen = ref(false)
const targetVhostToDelete = ref(null)
const deletionImpact = ref(null)

const primaryVhosts = computed(() => {
  return vhosts.value.filter((v) => v.type === 'primary')
})

function getPhpLabel(phpId) {
  if (!phpId) return 'Tắt'
  const p = phpVersions.value.find((item) => item.id === phpId)
  return p ? `PHP ${p.version}` : `PHP #${phpId}`
}

function onDocrootOwnerChange() {
  if (form.value.docroot_owner_id) {
    const owner = vhosts.value.find((v) => v.id === form.value.docroot_owner_id)
    if (owner) {
      form.value.docroot = owner.docroot
    }
  }
}

async function loadData() {
  loading.value = true
  try {
    const [vList, pList] = await Promise.all([
      vhostsApi.list(accountStore.selectedAccountId),
      phpApi.listVersions(),
    ])
    vhosts.value = vList || []
    phpVersions.value = pList || []
    if (phpVersions.value.length > 0 && !form.value.php_version_id) {
      form.value.php_version_id = phpVersions.value[0].id
    }
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

function openAddModal() {
  form.value.type = 'primary'
  form.value.fqdn = ''
  form.value.docroot = ''
  form.value.docroot_owner_id = null
  docrootMode.value = 'new'
  errorMessage.value = ''
  isAddModalOpen.value = true
}

async function submitCreateVhost() {
  if (!accountStore.currentAccount) {
    errorMessage.value = 'Vui lòng chọn tài khoản hosting hợp lệ.'
    return
  }

  submitting.value = true
  errorMessage.value = ''

  try {
    const payload = {
      account_id: accountStore.currentAccount.id,
      parent_vhost_id: form.value.type === 'subdomain' ? form.value.parent_vhost_id : null,
      type: form.value.type,
      fqdn: form.value.fqdn,
      docroot: form.value.docroot,
      docroot_owner_id: docrootMode.value === 'existing' ? form.value.docroot_owner_id : null,
      web_engine: form.value.web_engine,
      php_version_id: form.value.php_version_id,
    }

    await vhostsApi.create(payload)
    isAddModalOpen.value = false
    await loadData()
  } catch (err) {
    errorMessage.value = err.message
  } finally {
    submitting.value = false
  }
}

async function promptDelete(vhost) {
  targetVhostToDelete.value = vhost
  isDeleteModalOpen.value = true
  deletionImpact.value = null

  try {
    const impact = await vhostsApi.getDeletionImpact(vhost.id)
    deletionImpact.value = impact
  } catch (err) {
    alert(`Lỗi phân tích: ${err.message}`)
    isDeleteModalOpen.value = false
  }
}

async function executeDelete(deleteFiles) {
  if (!targetVhostToDelete.value) return
  deleting.value = true
  try {
    await vhostsApi.delete(targetVhostToDelete.value.id, deleteFiles)
    isDeleteModalOpen.value = false
    await loadData()
  } catch (err) {
    alert(`Xóa thất bại: ${err.message}`)
  } finally {
    deleting.value = false
  }
}

watch(() => accountStore.selectedAccountId, () => {
  loadData()
})

onMounted(() => {
  loadData()
})
</script>
