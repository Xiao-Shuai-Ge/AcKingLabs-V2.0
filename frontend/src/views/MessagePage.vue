<script setup lang="ts">
// 消息中心：分类标签（含未读角标）、点击已读（触屏也可消除未读）
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getMessageCount, getMessageList, markMessageRead } from '@/api/message'
import type { MessageItem } from '@/api/types'
import { avatarUrl, timeAgo } from '@/utils/format'
import { useMessage } from '@/composables/useMessage'

const router = useRouter()
const { addMessage } = useMessage()

const tabs = [
  { key: 'all', label: '全部' },
  { key: 'system', label: '系统' },
  { key: 'like', label: '点赞' },
  { key: 'comment', label: '评论' },
] as const

const tab = ref<(typeof tabs)[number]['key']>('all')
const counts = ref({ count: 0, like_count: 0, comment_count: 0, system_count: 0 })
const messages = ref<MessageItem[]>([])
const loading = ref(false)
const noMore = ref(false)

function unreadOf(key: string): number {
  if (key === 'all') return counts.value.count
  if (key === 'like') return counts.value.like_count
  if (key === 'comment') return counts.value.comment_count
  if (key === 'system') return counts.value.system_count
  return 0
}

async function refreshCount() {
  try {
    counts.value = await getMessageCount()
  } catch {
    /* 静默 */
  }
}

async function load(reset = true) {
  if (loading.value) return
  loading.value = true
  try {
    const before =
      reset || messages.value.length === 0 ? '' : messages.value[messages.value.length - 1].id
    const res = await getMessageList({ type: tab.value, before, count: 20 })
    messages.value = reset ? res.messages : [...messages.value, ...res.messages]
    noMore.value = res.messages.length < 20
  } finally {
    loading.value = false
  }
}

function switchTab(key: (typeof tabs)[number]['key']) {
  tab.value = key
  load(true)
}

async function openMessage(m: MessageItem) {
  if (!m.is_read) {
    try {
      await markMessageRead({ id: m.id })
      m.is_read = true
      refreshCount()
    } catch {
      /* 静默 */
    }
  }
  if (m.url) router.push(m.url)
}

async function markAll() {
  try {
    await markMessageRead({ all: true, type: tab.value })
    addMessage('已全部标记为已读', 'success')
    await load(true)
    refreshCount()
  } catch {
    /* 静默 */
  }
}

onMounted(async () => {
  refreshCount()
  load(true)
})
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-4xl mx-auto px-4 py-8">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-bold">消息中心</h1>
        <button
          v-if="unreadOf(tab) > 0"
          class="px-4 py-1.5 rounded-lg border border-gray-300 text-sm text-gray-600 hover:bg-gray-100"
          @click="markAll"
        >
          全部已读
        </button>
      </div>

      <!-- 标签页 -->
      <div class="flex gap-2 mb-5 border-b pb-3 overflow-x-auto hide-scrollbar">
        <button
          v-for="t in tabs"
          :key="t.key"
          class="shrink-0 px-4 py-1.5 rounded-full text-sm relative transition-colors"
          :class="tab === t.key ? 'bg-black text-white' : 'text-gray-600 hover:bg-gray-100'"
          @click="switchTab(t.key)"
        >
          {{ t.label }}
          <span v-if="unreadOf(t.key) > 0" class="ml-1 text-red-500 font-medium">
            {{ unreadOf(t.key) > 99 ? '99+' : unreadOf(t.key) }}
          </span>
        </button>
      </div>

      <!-- 消息列表 -->
      <div class="bg-white rounded-lg border border-gray-200 divide-y" v-loading="loading">
        <div
          v-if="messages.length === 0 && !loading"
          class="py-16 text-center text-sm text-gray-400"
        >
          （暂无消息）
        </div>
        <div
          v-for="m in messages"
          :key="m.id"
          class="flex items-start gap-3 p-4 cursor-pointer hover:bg-gray-50 transition-colors"
          @click="openMessage(m)"
        >
          <div class="relative shrink-0">
            <img
              :src="avatarUrl(m.sender?.avatar)"
              class="w-10 h-10 rounded-full object-cover"
              alt=""
            />
            <span
              class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full bg-red-500"
              v-if="!m.is_read"
            />
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-bold text-sm truncate">{{ m.sender?.username ?? '系统' }}</span>
              <span class="text-xs text-gray-400">{{ timeAgo(m.created_at) }}</span>
            </div>
            <p class="text-sm text-gray-600 truncate mt-0.5">
              {{ m.content }}
              <i v-if="m.url" class="fa-solid fa-arrow-right text-xs text-gray-300 ml-1" />
            </p>
          </div>
        </div>

        <button
          v-if="!noMore && messages.length > 0"
          class="w-full py-3 text-sm text-gray-500 hover:bg-gray-50"
          @click="load(false)"
        >
          加载更多
        </button>
      </div>
    </div>
  </div>
</template>
