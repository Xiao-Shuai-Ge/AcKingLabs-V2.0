<script setup lang="ts">
// 比赛详情：信息头部 + 关联题解列表 + 贡献题解入口
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getContestDetail } from '@/api/contest'
import { getPostList } from '@/api/post'
import type { ContestItem, PostItem } from '@/api/types'
import PostCard from '@/components/PostCard.vue'
import { formatDateTime, formatDuration, contestStatus } from '@/utils/format'
import { useMessage } from '@/composables/useMessage'

const route = useRoute()
const router = useRouter()
const { addMessage } = useMessage()

const contestId = route.params.id as string
const contest = ref<ContestItem | null>(null)
const solutions = ref<PostItem[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(true)
const count = 5

const statusLabel: Record<string, string> = {
  upcoming: '未开始',
  ongoing: '进行中',
  ended: '已结束',
}

async function load() {
  loading.value = true
  try {
    const res = await getContestDetail(contestId)
    contest.value = res.contest
    await loadSolutions()
  } catch {
    addMessage('比赛不存在', 'error')
    setTimeout(() => router.replace('/contest'), 1000)
  } finally {
    loading.value = false
  }
}

async function loadSolutions() {
  if (!contest.value) return
  try {
    const res = await getPostList({
      type: 'solution',
      sort: 'new',
      source: contest.value.url,
      page: page.value,
      count,
    })
    solutions.value = res.posts
    total.value = res.total
  } catch {
    /* 静默 */
  }
}

function changePage(p: number) {
  page.value = p
  loadSolutions()
}

function goCreate() {
  if (!contest.value) return
  router.push({
    path: '/learn/create',
    query: { contest_url: contest.value.url, contest_title: contest.value.title },
  })
}

onMounted(load)
</script>

<template>
  <div class="pt-[60px]" v-loading="loading">
    <template v-if="contest">
      <!-- 头部横幅 -->
      <div class="bg-gradient-to-r from-gray-800 to-gray-500 text-white">
        <div class="max-w-4xl mx-auto px-4 py-10">
          <h1 class="text-xl md:text-3xl font-bold">{{ contest.title }}</h1>
          <div class="mt-3 text-sm text-gray-200">
            <i class="fa-regular fa-clock mr-1.5" />
            {{ formatDateTime(contest.start_time) }} 至 {{ formatDateTime(contest.end_time) }}
            <span class="text-gray-300">（{{ formatDuration(contest.duration) }}）</span>
          </div>
          <div class="mt-4 flex items-center gap-3">
            <span class="px-3 py-1 rounded-full text-sm bg-white/20">{{
              statusLabel[contestStatus(contest.start_time, contest.end_time)]
            }}</span>
            <a
              :href="contest.url"
              target="_blank"
              rel="noopener"
              class="px-4 py-1.5 rounded-lg bg-white text-gray-900 text-sm font-medium hover:bg-gray-100"
            >
              前往比赛 <i class="fa-solid fa-arrow-up-right-from-square ml-1 text-xs" />
            </a>
          </div>
        </div>
      </div>

      <!-- 题解列表 -->
      <div class="max-w-4xl mx-auto px-4 py-8">
        <div class="flex items-center justify-between mb-5">
          <h2 class="text-lg font-bold">比赛题解（{{ total }}）</h2>
        </div>
        <div class="space-y-4">
          <PostCard
            v-for="(p, i) in solutions"
            :key="p.id"
            :post="p"
            :index="i"
            :show-type="false"
          />
          <div
            v-if="solutions.length === 0"
            class="py-16 text-center bg-white rounded-lg border border-gray-200 cursor-pointer hover:shadow transition-shadow"
            @click="goCreate"
          >
            <div class="text-gray-400">当前比赛没有人提供题解，点我贡献！</div>
          </div>
        </div>
        <div v-if="total > count" class="mt-6 flex justify-center">
          <el-pagination
            layout="prev, pager, next"
            :total="total"
            :page-size="count"
            :current-page="page"
            @current-change="changePage"
          />
        </div>
      </div>

      <!-- 贡献题解浮动按钮 -->
      <button
        class="fixed bottom-8 right-8 z-40 w-14 h-14 rounded-full bg-blue-500 text-white shadow-lg flex items-center justify-center text-2xl hover:bg-blue-600 hover:scale-105 transition-all"
        aria-label="贡献题解"
        @click="goCreate"
      >
        <i class="fa-solid fa-pen-nib" />
      </button>
    </template>
  </div>
</template>
