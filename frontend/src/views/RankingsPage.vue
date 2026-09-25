<script setup lang="ts">
// 排行榜：经验值排名（渐变头部 + 名次底色），行可点进主页
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getRankings } from '@/api/user'
import type { RankingItem } from '@/api/user'
import { CheckLevel, GetTextColor } from '@/utils/level'
import { avatarUrl } from '@/utils/format'
import { useMessage } from '@/composables/useMessage'

const router = useRouter()
const { addMessage } = useMessage()

const items = ref<RankingItem[]>([])
const total = ref(0)
const page = ref(1)
const count = 10
const loading = ref(true)

// 名次底色
function rankBg(rank: number): string {
  if (rank === 1) return 'bg-red-50'
  if (rank === 2) return 'bg-orange-50'
  if (rank === 3) return 'bg-yellow-50'
  if (rank <= 15) return 'bg-blue-50'
  if (rank <= 30) return 'bg-green-50'
  return ''
}

function rankBadge(rank: number): string {
  if (rank === 1) return 'bg-red-500'
  if (rank === 2) return 'bg-orange-400'
  if (rank === 3) return 'bg-yellow-400'
  return 'bg-gray-300'
}

async function load() {
  loading.value = true
  try {
    const res = await getRankings({ page: page.value, count })
    items.value = res.rankings
    total.value = res.total
  } catch {
    addMessage('加载排行榜失败', 'error')
  } finally {
    loading.value = false
  }
}

function changePage(p: number) {
  page.value = p
  load()
}

onMounted(load)
</script>

<template>
  <div class="pt-[60px] min-h-screen bg-gray-50">
    <!-- 渐变头部 -->
    <div class="bg-gradient-to-r from-yellow-400 to-orange-500 py-12">
      <div class="max-w-4xl mx-auto px-4 text-white">
        <h1 class="text-3xl font-bold flex items-center gap-3">
          <i class="fa-solid fa-trophy" />排行榜
        </h1>
        <p class="text-white/80 mt-2 text-sm">按经验值排名 · 每周公开打卡可获得双倍经验</p>
      </div>
    </div>

    <div class="max-w-4xl mx-auto px-4 py-8" v-loading="loading">
      <div class="space-y-2">
        <div
          v-for="(item, i) in items"
          :key="item.id"
          class="flex items-center gap-4 p-4 rounded-lg border border-gray-200 cursor-pointer transition-all hover:shadow-md fade-in-up"
          :class="rankBg(item.rank)"
          :style="{ animationDelay: `${i * 0.04}s` }"
          @click="router.push(`/profile/${item.id}`)"
        >
          <div
            class="w-9 h-9 rounded-full flex items-center justify-center font-bold text-white shrink-0"
            :class="rankBadge(item.rank)"
          >
            {{ item.rank }}
          </div>
          <img
            :src="avatarUrl(item.avatar)"
            class="w-10 h-10 rounded-full object-cover object-top border-2 border-white shrink-0"
            alt=""
          />
          <div class="flex-1 min-w-0">
            <span class="font-bold truncate" :class="GetTextColor(CheckLevel(item.xp, item.role))">
              {{ item.username }}
            </span>
            <span
              v-if="item.role >= 3"
              class="ml-2 px-1.5 py-0.5 rounded text-xs bg-red-100 text-red-600"
              >管理</span
            >
          </div>
          <div class="text-right shrink-0">
            <div class="font-bold text-lg">{{ item.xp }}</div>
            <div class="text-xs text-gray-400">XP</div>
          </div>
        </div>
        <div
          v-if="items.length === 0 && !loading"
          class="py-16 text-center text-gray-400 bg-white rounded-lg border border-gray-200"
        >
          暂无数据
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
  </div>
</template>
