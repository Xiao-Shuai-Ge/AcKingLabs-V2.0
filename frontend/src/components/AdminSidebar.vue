<script setup lang="ts">
// 管理后台侧边栏（响应式：桌面固定左侧，移动端抽屉）
import { useRoute } from 'vue-router'

const route = useRoute()
const collapsed = defineModel<boolean>({ default: false })

const items = [
  { path: '/admin/users', label: '用户管理', icon: 'fa-users' },
  { path: '/admin/posts', label: '内容管理', icon: 'fa-file-lines' },
  { path: '/admin/resumes', label: '简历管理', icon: 'fa-file-signature' },
]
</script>

<template>
  <!-- 移动端遮罩 -->
  <div
    v-if="!collapsed"
    class="md:hidden fixed inset-0 bg-black/40 z-30"
    @click="collapsed = true"
  />
  <aside
    class="fixed left-0 top-[60px] bottom-0 z-40 bg-white border-r w-52 transition-transform"
    :class="[collapsed ? '-translate-x-full' : 'translate-x-0', 'md:translate-x-0']"
  >
    <div class="p-4 border-b">
      <div class="font-bold text-lg">管理后台</div>
    </div>
    <nav class="p-2">
      <router-link
        v-for="item in items"
        :key="item.path"
        :to="item.path"
        class="flex items-center gap-3 px-4 py-3 rounded-lg mb-1 text-sm font-medium transition-colors"
        :class="
          route.path === item.path ? 'bg-black text-white' : 'text-gray-700 hover:bg-gray-100'
        "
      >
        <i class="fa-solid w-5" :class="item.icon" />
        {{ item.label }}
      </router-link>
      <router-link
        to="/"
        class="flex items-center gap-3 px-4 py-3 rounded-lg mb-1 text-sm font-medium text-gray-400 hover:bg-gray-100"
      >
        <i class="fa-solid w-5 fa-arrow-left" />
        返回前台
      </router-link>
    </nav>
  </aside>
  <!-- 桌面端内容占位（fixed sidebar 宽度补偿） -->
  <div class="hidden md:block w-52 shrink-0" />
</template>
