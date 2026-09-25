<script setup lang="ts">
// 创建自定义比赛（管理员）；datetimerange 用毫秒值绑定，无时区换算
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { adminCreateContest } from '@/api/contest'
import { useMessage } from '@/composables/useMessage'

const router = useRouter()
const { addMessage, codeHandler } = useMessage()

const form = reactive({ title: '', url: '', range: [] as string[] })
const loading = ref(false)

const valid = computed(
  () => form.title.trim() !== '' && form.url.trim() !== '' && form.range.length === 2,
)

async function submit() {
  if (!valid.value || loading.value) return
  loading.value = true
  try {
    await adminCreateContest({
      title: form.title.trim(),
      url: form.url.trim(),
      start_time: Number(form.range[0]),
      end_time: Number(form.range[1]),
    })
    addMessage('创建成功！', 'success')
    router.push('/contest')
  } catch (err) {
    codeHandler(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-2xl mx-auto px-4 py-10">
      <h1 class="text-2xl font-bold mb-8">创建自定义比赛</h1>

      <div class="bg-white rounded-lg border border-gray-200 p-6 space-y-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">标题</label>
          <el-input v-model="form.title" placeholder="请输入比赛标题" maxlength="100" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">时间</label>
          <el-date-picker
            v-model="form.range"
            type="datetimerange"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="x"
            class="!w-full"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">网址（必填）</label>
          <el-input v-model="form.url" placeholder="请输入比赛链接" />
        </div>

        <div class="flex gap-3 pt-2">
          <button
            class="px-8 py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed"
            :disabled="!valid || loading"
            @click="submit"
          >
            {{ loading ? '创建中...' : '创建比赛' }}
          </button>
          <button
            class="px-8 py-2.5 rounded-lg border-2 border-gray-300 text-gray-600 font-medium hover:bg-gray-100"
            @click="router.back()"
          >
            返回
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
