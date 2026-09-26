<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
    <div class="bg-[#1e293b] border border-slate-700 rounded-lg max-w-lg w-full shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-700 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-red-400 flex items-center gap-2">
          <svg class="w-5 h-5 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          Xác nhận xóa: {{ impact?.fqdn }}
        </h3>
        <button @click="$emit('close')" class="text-slate-400 hover:text-white">&times;</button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 space-y-4 text-sm text-slate-300">
        <!-- BLOCKED CASE -->
        <div v-if="impact && !impact.allow_delete" class="p-3 bg-red-950/60 border border-red-800 rounded text-red-200">
          <p class="font-bold mb-1">Thao tác bị chặn:</p>
          <p>{{ impact.reason_if_blocked }}</p>
        </div>

        <!-- ALLOWED CASE WITH IMPACT DETAILS -->
        <div v-else-if="impact" class="space-y-3">
          <p class="text-slate-200">
            Hành động này sẽ thực hiện các thay đổi sau trên hệ thống:
          </p>

          <ul class="space-y-2 border border-slate-700/60 rounded bg-slate-900/60 p-3">
            <li class="flex items-start gap-2">
              <span class="text-red-400 font-bold">✓</span>
              <span>Xóa vhost config: <code class="text-xs bg-slate-800 px-1 py-0.5 rounded text-amber-300">/etc/wrenpanel/nginx/vhosts/{{ impact.fqdn }}.conf</code> và reload Nginx.</span>
            </li>

            <!-- DOCROOT IMPACT -->
            <li class="flex items-start gap-2">
              <span v-if="impact.shared_docroot" class="text-amber-400 font-bold">ℹ</span>
              <span v-else-if="impact.can_delete_files" class="text-red-400 font-bold">!</span>
              <span v-else class="text-slate-400 font-bold">•</span>

              <div class="flex-1">
                <div>Thư mục webroot: <code class="text-xs bg-slate-800 px-1 py-0.5 rounded text-sky-300">{{ impact.docroot }}</code></div>
                
                <p v-if="impact.shared_docroot" class="text-xs text-amber-400 mt-1">
                  Đang có <strong>{{ impact.shared_count }}</strong> vhost khác dùng chung thư mục này. Hệ thống <strong>bắt buộc giữ nguyên toàn bộ mã nguồn</strong> để tránh ảnh hưởng website khác.
                </p>

                <p v-else-if="!impact.is_docroot_owner" class="text-xs text-amber-400 mt-1">
                  Vhost này dùng chung folder từ domain gốc. Hệ thống <strong>giữ nguyên mã nguồn</strong>.
                </p>

                <div v-else-if="impact.can_delete_files" class="mt-2 pt-2 border-t border-slate-700">
                  <label class="flex items-center gap-2 cursor-pointer text-slate-200 hover:text-white">
                    <input type="checkbox" v-model="deleteFilesOption" class="rounded bg-slate-800 border-slate-600 text-red-500 focus:ring-0">
                    <span class="font-medium text-red-300">Đồng thời xóa sạch toàn bộ file mã nguồn trong thư mục này</span>
                  </label>
                  <p class="text-xs text-slate-400 mt-0.5 ml-5">Nếu không chọn, dữ liệu file sẽ được giữ lại an toàn.</p>
                </div>
              </div>
            </li>
          </ul>
        </div>

        <div v-else class="py-4 text-center text-slate-400">
          Đang phân tích tác động hệ thống...
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-4 bg-slate-900/60 border-t border-slate-700 flex justify-end gap-3">
        <button
          @click="$emit('close')"
          class="px-4 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded border border-slate-600 transition"
        >
          Hủy bỏ
        </button>

        <button
          v-if="impact && impact.allow_delete"
          @click="$emit('confirm', deleteFilesOption)"
          :disabled="loading"
          class="px-4 py-2 text-sm font-medium bg-red-600 hover:bg-red-700 text-white rounded transition disabled:opacity-50 flex items-center gap-1.5"
        >
          <span v-if="loading">Đang xóa...</span>
          <span v-else>Xác nhận xóa vhost</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  impact: Object,
  loading: Boolean,
})

const emit = defineEmits(['close', 'confirm'])

const deleteFilesOption = ref(false)

watch(() => props.isOpen, (open) => {
  if (open) {
    deleteFilesOption.value = false
  }
})
</script>
