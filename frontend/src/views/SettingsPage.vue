<script setup lang="ts">
// 个人设置：通知偏好（存 users.settings JSON 列的 notify 分组，
// 站内消息按偏好过滤生成；"系统消息邮件"开关控制异步邮件同步）
import { computed, onMounted, reactive, ref } from 'vue'
import { getUserSetting, updateUserSetting } from '@/api/user'
import type { NotifySettings } from '@/api/user'
import { useMessage } from '@/composables/useMessage'

const { addMessage, codeHandler } = useMessage()

const defaults: NotifySettings = {
  like: true,
  comment: true,
  mention: true,
  help_post: true,
  system_email: false,
}

const form = reactive<NotifySettings>({ ...defaults })
const saved = ref<NotifySettings>({ ...defaults })
const loading = ref(true)
const saving = ref(false)

const dirty = computed(() =>
  (Object.keys(defaults) as (keyof NotifySettings)[]).some((k) => form[k] !== saved.value[k]),
)

const groups: {
  key: keyof NotifySettings
  icon: string
  color: string
  title: string
  desc: string
}[] = [
  {
    key: 'like',
    icon: 'fa-heart',
    color: 'text-red-500 bg-red-50',
    title: '点赞通知',
    desc: '别人赞了你的帖子或评论时提醒',
  },
  {
    key: 'comment',
    icon: 'fa-comment-dots',
    color: 'text-blue-500 bg-blue-50',
    title: '评论 / 回复通知',
    desc: '别人评论你的帖子、回复你的评论时提醒',
  },
  {
    key: 'mention',
    icon: 'fa-at',
    color: 'text-green-500 bg-green-50',
    title: '提及通知',
    desc: '别人在帖子或评论中 @你 时提醒',
  },
  {
    key: 'help_post',
    icon: 'fa-circle-question',
    color: 'text-yellow-500 bg-yellow-50',
    title: '求助帖提醒',
    desc: '有人发布新的求助帖时提醒，帮小伙伴们解答',
  },
  {
    key: 'system_email',
    icon: 'fa-envelope',
    color: 'text-purple-500 bg-purple-50',
    title: '系统消息邮件同步',
    desc: '精选、下架等系统消息同时发送到你的邮箱',
  },
]

async function load() {
  try {
    const s = await getUserSetting()
    Object.assign(form, defaults, s.notify)
    saved.value = { ...form }
  } catch (err) {
    codeHandler(err)
  } finally {
    loading.value = false
  }
}

function resetDefaults() {
  Object.assign(form, defaults)
}

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    await updateUserSetting({ notify: { ...form } })
    saved.value = { ...form }
    addMessage('设置已保存', 'success')
  } catch (err) {
    codeHandler(err)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="pt-[60px] min-h-screen bg-gray-50">
    <div class="max-w-2xl mx-auto px-4 py-10" v-loading="loading">
      <h1 class="text-2xl font-bold">个人设置</h1>
      <p class="text-sm text-gray-500 mt-1 mb-8">管理您的通知偏好设置</p>

      <div class="bg-white rounded-xl border border-gray-200 divide-y">
        <div v-for="g in groups" :key="g.key" class="flex items-center gap-4 p-5">
          <div
            class="w-11 h-11 rounded-xl flex items-center justify-center shrink-0"
            :class="g.color"
          >
            <i class="fa-solid text-lg" :class="g.icon" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="font-medium">{{ g.title }}</div>
            <div class="text-sm text-gray-400 mt-0.5">{{ g.desc }}</div>
          </div>
          <el-switch v-model="form[g.key]" />
        </div>
      </div>

      <div class="flex items-center gap-3 mt-6">
        <button
          class="px-6 py-2.5 rounded-lg bg-black text-white text-sm font-medium hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed"
          :disabled="!dirty || saving"
          @click="save"
        >
          {{ saving ? '保存中...' : '保存设置' }}
        </button>
        <button
          class="px-5 py-2.5 rounded-lg border-2 border-gray-300 text-gray-600 text-sm font-medium hover:bg-gray-100 disabled:opacity-40"
          :disabled="!dirty"
          @click="Object.assign(form, saved)"
        >
          取消
        </button>
        <button class="ml-auto text-sm text-gray-400 hover:text-gray-600" @click="resetDefaults">
          恢复默认设置
        </button>
      </div>

      <p class="text-xs text-gray-400 mt-6 flex items-center gap-1.5">
        <i class="fa-solid fa-circle-info" />
        关闭通知后，您仍然可以在消息中心查看所有历史通知
      </p>
    </div>
  </div>
</template>
