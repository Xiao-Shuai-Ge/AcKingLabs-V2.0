<script setup lang="ts">
// 打卡（周记）页：打卡按钮状态机 + 周期选择 + 推荐/最新排序 + 周帖子流
// 版式：大按钮居中 → 时间说明 → 周期选择 → 排序 → 帖子（黑边框卡片）
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { getWeekStatus, getDiaryWeeks, getPostMore } from '@/api/post'
import type { PostItem } from '@/api/types'
import PostCard from '@/components/PostCard.vue'
import {
  getWeekCode,
  getValidSubmissionTime,
  getStudyTimeString,
  timestampFormat,
} from '@/utils/week'
import { useMessage } from '@/composables/useMessage'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { currentUser } = storeToRefs(userStore)
const { addMessage } = useMessage()

// ---- 打卡周期状态 ----
const weekStatus = ref<Awaited<ReturnType<typeof getWeekStatus>> | null>(null)
const myWeeks = ref<Map<string, { post_id: string; is_private: boolean }>>(new Map())
const selectedTime = ref(Date.now())

const selectedWeek = computed(() => getWeekCode(new Date(selectedTime.value)))
const isCurrentWeek = computed(() => Math.abs(selectedTime.value - Date.now()) < 24 * 3600 * 1000)
const studyWindow = computed(() => getStudyTimeString(new Date(selectedTime.value)))
const validWindow = computed(() => getValidSubmissionTime(new Date(selectedTime.value)))

// 打卡按钮状态：未登录 / 未到时间 / 已提交 / 可发布
const diaryState = computed<'offline' | 'invalid-time' | 'submitted' | 'ready'>(() => {
  if (!userStore.isLogin) return 'offline'
  if (!selectedWeek.value.valid) return 'invalid-time'
  if (myWeeks.value.has(selectedWeek.value.code)) return 'submitted'
  return 'ready'
})

const buttonMeta = computed(() => {
  switch (diaryState.value) {
    case 'ready':
      return { text: '发布周记', icon: 'fa-pen-nib', cls: 'bg-black hover:bg-gray-800' }
    case 'submitted':
      return { text: '已提交', icon: 'fa-circle-check', cls: 'bg-green-500 hover:bg-green-600' }
    case 'invalid-time':
      return { text: '未到时间', icon: 'fa-clock', cls: 'bg-gray-400 cursor-not-allowed' }
    default:
      return { text: '未登录', icon: 'fa-user-lock', cls: 'bg-gray-400 cursor-not-allowed' }
  }
})

const submittedPostId = computed(() => myWeeks.value.get(selectedWeek.value.code)?.post_id)

function prevWeek() {
  selectedTime.value -= 7 * 24 * 3600 * 1000
  syncUrl()
}
function nextWeek() {
  selectedTime.value += 7 * 24 * 3600 * 1000
  syncUrl()
}

function syncUrl() {
  if (!isCurrentWeek.value) {
    router.replace({ query: { ...route.query, time: String(selectedTime.value) } })
  } else {
    const q = { ...route.query }
    delete q.time
    router.replace({ query: q })
  }
  reload()
}

// ---- 帖子列表 ----
const sort = ref<'hot' | 'new'>((route.query.sort as 'hot' | 'new') || 'hot')
const posts = ref<PostItem[]>([])
const loading = ref(false)
const noMore = ref(false)

function cursor(): string {
  if (posts.value.length === 0) return '0'
  const last = posts.value[posts.value.length - 1]
  return sort.value === 'hot' ? String(last.hot_score) : last.id
}

async function loadMore(reset = false) {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getPostMore({
      type: 'diary',
      source: selectedWeek.value.code,
      sort: sort.value,
      cursor: reset ? '0' : cursor(),
      count: 20,
    })
    posts.value = reset ? res.posts : [...posts.value, ...res.posts]
    noMore.value = res.posts.length < 20
  } catch {
    addMessage('加载帖子失败', 'error')
  } finally {
    loading.value = false
  }
}

function switchSort(s: 'hot' | 'new') {
  if (sort.value === s) return
  sort.value = s
  router.replace({ query: { ...route.query, sort: s === 'hot' ? undefined : s } })
  reload()
}

function reload() {
  posts.value = []
  noMore.value = false
  loadMore(true)
}

function goCreateOrEdit() {
  if (diaryState.value === 'submitted' && submittedPostId.value) {
    router.push(`/diary/${submittedPostId.value}/edit`)
  } else {
    router.push('/diary/create')
  }
}

onMounted(async () => {
  // URL 恢复周期与排序（单一初始化入口）
  const t = Number(route.query.time)
  if (t > 0) selectedTime.value = t
  if (route.query.sort === 'new') sort.value = 'new'

  try {
    weekStatus.value = await getWeekStatus()
  } catch {
    /* 周期信息拿不到也可以浏览 */
  }
  if (userStore.isLogin && currentUser.value) {
    try {
      const res = await getDiaryWeeks(currentUser.value.id)
      const map = new Map()
      for (const w of res.weeks)
        map.set(w.week_code, { post_id: w.post_id, is_private: w.is_private })
      myWeeks.value = map
    } catch {
      /* 静默 */
    }
  }
  loadMore(true)
})
</script>

<template>
  <div class="pt-[60px] min-h-screen bg-gray-50 flex flex-col items-center py-10 px-4">
    <!-- 打卡大按钮 -->
    <div class="mb-6 mt-8">
      <button
        class="w-32 h-32 rounded-full text-white flex items-center justify-center shadow-lg transition-all duration-300 whitespace-nowrap hover:scale-105"
        :class="buttonMeta.cls"
        @click="goCreateOrEdit"
      >
        <i class="fa-solid text-2xl" :class="buttonMeta.icon" />
        <span class="ml-2 font-medium">{{ buttonMeta.text }}</span>
      </button>
    </div>

    <!-- 时间说明 -->
    <div class="text-gray-500 text-sm text-center mb-8 leading-relaxed">
      <p>
        记录学习时间: {{ timestampFormat(studyWindow.from) }} ~
        {{ timestampFormat(studyWindow.to) }}
      </p>
      <p>
        有效打卡时间: {{ timestampFormat(validWindow.from) }} ~
        {{ timestampFormat(validWindow.to) }}
      </p>
      <p v-if="diaryState === 'submitted'" class="text-green-600 mt-1">
        本周周记已提交，点击上方按钮可编辑
      </p>
      <p v-else-if="diaryState === 'ready'" class="text-gray-400 mt-1">
        公开打卡可获得双倍经验，私密仅自己与管理员可见
      </p>
    </div>

    <!-- 周期选择 -->
    <div class="w-full max-w-4xl flex items-center justify-between mb-8">
      <button
        class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center hover:bg-gray-300 transition-colors"
        @click="prevWeek"
      >
        <i class="fa-solid fa-chevron-left" />
      </button>
      <div class="text-center">
        <div class="text-lg font-bold">{{ selectedWeek.name }}</div>
        <div v-if="!isCurrentWeek" class="text-xs text-gray-400 mt-0.5">（历史周期）</div>
      </div>
      <button
        class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center hover:bg-gray-300 transition-colors"
        :disabled="isCurrentWeek"
        :class="isCurrentWeek ? 'opacity-30 cursor-not-allowed hover:bg-gray-200' : ''"
        @click="nextWeek"
      >
        <i class="fa-solid fa-chevron-right" />
      </button>
    </div>

    <!-- 排序 + 帖子 -->
    <div class="w-full max-w-4xl">
      <div class="flex items-center gap-2 mb-6">
        <button
          v-for="s in [
            { key: 'hot', label: '推荐' },
            { key: 'new', label: '最新' },
          ]"
          :key="s.key"
          class="px-4 py-1.5 rounded-full border-2 text-sm font-medium transition-colors"
          :class="
            sort === s.key
              ? 'bg-black text-white border-black'
              : 'border-gray-800 text-gray-700 hover:bg-gray-100'
          "
          @click="switchSort(s.key as 'hot' | 'new')"
        >
          {{ s.label }}
        </button>
        <span class="ml-auto text-sm text-gray-400">{{ selectedWeek.name }} 的周记</span>
      </div>

      <div class="space-y-6 mb-64">
        <PostCard
          v-for="(p, i) in posts"
          :key="p.id"
          :post="p"
          :index="i"
          :show-type="false"
          dark
        />
        <div
          v-if="posts.length === 0 && !loading"
          class="py-16 text-center text-gray-400 bg-white rounded-lg border-2 border-gray-800"
        >
          这个星期还没有人发布周记，快来抢占第一篇吧！
        </div>
        <div v-if="loading" class="py-10 text-center text-gray-400">
          <i class="fa-solid fa-spinner fa-spin mr-2" />加载中...
        </div>
        <button
          v-if="!noMore && posts.length > 0 && !loading"
          class="w-full py-2.5 rounded-lg border-2 border-gray-800 text-sm font-medium hover:bg-gray-100"
          @click="loadMore()"
        >
          加载更多
        </button>
        <div v-if="noMore && posts.length > 0" class="py-6 text-center text-sm text-gray-400">
          没有更多帖子了
        </div>
      </div>
    </div>
  </div>
</template>
