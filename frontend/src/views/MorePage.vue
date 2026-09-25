<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const entries = [
  {
    icon: 'fa-ranking-star',
    color: 'text-yellow-500',
    bg: 'bg-yellow-50',
    title: '排行榜',
    desc: '查看网站所有用户的经验排名',
    path: '/more/rankings',
  },
  {
    icon: 'fa-circle-info',
    color: 'text-blue-500',
    bg: 'bg-blue-50',
    title: '关于我们',
    desc: '关于 AcKing 实验室的介绍与贡献成员',
    path: '/more/about',
  },
  {
    icon: 'fa-file-signature',
    color: 'text-green-500',
    bg: 'bg-green-50',
    title: '投递简历',
    desc: '加入 AcKing 实验室，开启您的竞赛之路',
    path: '/more/resume',
  },
]
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-4xl mx-auto px-4 py-10">
      <h1 class="text-2xl font-bold mb-8">更多</h1>
      <div class="space-y-4">
        <div
          v-for="(e, i) in entries"
          :key="e.path"
          class="bg-white rounded-xl border border-gray-200 p-5 flex items-center gap-4 cursor-pointer hover:shadow-lg hover:scale-[1.01] transition-all fade-in-up"
          :style="{ animationDelay: `${i * 0.08}s` }"
          @click="router.push(e.path)"
        >
          <div class="w-12 h-12 rounded-xl flex items-center justify-center shrink-0" :class="e.bg">
            <i class="fa-solid text-xl" :class="[e.icon, e.color]" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="font-bold">{{ e.title }}</div>
            <div class="text-sm text-gray-500">{{ e.desc }}</div>
          </div>
          <i class="fa-solid fa-chevron-right text-gray-300" />
        </div>
      </div>
    </div>
  </div>
</template>
