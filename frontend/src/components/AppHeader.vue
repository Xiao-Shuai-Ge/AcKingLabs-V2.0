<script setup lang="ts">
// 全局顶栏：通栏三区布局（logo / 居中导航 / 用户区），移动端可点击展开
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { getMessageCount } from '@/api/message'
import { getAccessToken } from '@/api/http'
import { CheckLevel, GetTextColor } from '@/utils/level'
import { avatarUrl } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { currentUser } = storeToRefs(userStore)

const navItems = [
  { path: '/diary', label: '打卡', icon: 'fa-square-check', color: 'blue' },
  { path: '/learn', label: '学习', icon: 'fa-book-open-reader', color: 'yellow' },
  { path: '/contest', label: '比赛', icon: 'fa-chart-simple', color: 'red' },
  { path: '/more', label: '更多', icon: 'fa-bars', color: 'green' },
]
const colorMap: Record<string, string> = {
  blue: 'border-blue-500',
  yellow: 'border-yellow-500',
  red: 'border-red-500',
  green: 'border-green-500',
}

const unread = ref(0)
const userMenuOpen = ref(false)
const mobileMenuOpen = ref(false)

const level = computed(() => CheckLevel(currentUser.value?.xp ?? 0, currentUser.value?.role ?? -1))

async function refreshUnread() {
  if (!userStore.isLogin) {
    unread.value = 0
    return
  }
  try {
    const res = await getMessageCount()
    unread.value = res.count
  } catch {
    /* 静默 */
  }
}

function logout() {
  userStore.logout()
  router.push('/login')
}

onMounted(() => {
  userStore.refreshUser()
  refreshUnread()
})

// 路由切换兜底：登录后若用户信息尚未就绪（拉取失败过一次），这里自动补上；
// 顺带刷新未读数，避免长时间挂着的页面数字过期。
watch(
  () => route.path,
  () => {
    if (!currentUser.value && getAccessToken()) {
      userStore.refreshUser()
    }
    if (userStore.isLogin) {
      refreshUnread()
    }
  },
)
</script>

<template>
  <header class="fixed top-0 left-0 w-full h-[60px] bg-white shadow-sm z-50">
    <div class="w-full h-full flex items-center justify-between">
      <!-- 左：Logo 区 -->
      <div class="flex items-center ml-4 md:ml-5 w-1/6 min-w-0">
        <router-link to="/" class="flex items-center gap-2 min-w-0">
          <img src="/assets/AcKing_black.png" alt="AcKing" class="h-9 w-9 rounded-full shrink-0" />
          <span class="hidden xl:block text-lg font-bold tracking-wide truncate"
            >AcKing 学习分享平台</span
          >
        </router-link>
      </div>

      <!-- 中：导航（通栏居中，桌面端占满中间区域） -->
      <nav class="hidden md:flex h-full flex-1 justify-center overflow-x-auto hide-scrollbar">
        <div class="flex h-full">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="w-32 flex flex-col items-center justify-center border-b-4 transition-colors"
            :class="
              route.path.startsWith(item.path)
                ? `bg-black text-white ${colorMap[item.color]}`
                : 'text-gray-800 border-transparent hover:bg-gray-100'
            "
          >
            <i class="fa-solid text-xl" :class="item.icon" />
            <span class="text-sm font-medium mt-0.5">{{ item.label }}</span>
          </router-link>
        </div>
      </nav>

      <!-- 右：用户区（占 1/6） -->
      <div class="hidden md:flex items-center justify-end mr-5 w-1/6 relative">
        <template v-if="userStore.isLogin && currentUser">
          <router-link to="/message" class="relative p-2 hover:bg-gray-100 rounded-full mr-1">
            <i class="fa-solid fa-bell text-lg text-gray-600" />
            <span
              v-if="unread > 0"
              class="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[11px] leading-[18px] text-center"
            >
              {{ unread > 99 ? '99+' : unread }}
            </span>
          </router-link>
          <button
            class="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-gray-100"
            @click="userMenuOpen = !userMenuOpen"
          >
            <span
              class="hidden xl:block text-sm font-bold truncate max-w-[90px]"
              :class="GetTextColor(level)"
            >
              {{ currentUser.username }}
            </span>
            <img
              :src="avatarUrl(currentUser.avatar)"
              class="w-9 h-9 rounded-full object-cover"
              alt="avatar"
            />
          </button>
          <transition name="fade">
            <div
              v-if="userMenuOpen"
              class="absolute right-0 top-12 w-40 bg-white border rounded-lg shadow-lg py-1 text-sm z-50"
            >
              <router-link
                :to="`/profile/${currentUser.id}`"
                class="block px-4 py-2 hover:bg-gray-100"
                @click="userMenuOpen = false"
              >
                <i class="fa-solid fa-user mr-2 text-gray-500" />个人主页
              </router-link>
              <router-link
                to="/message"
                class="block px-4 py-2 hover:bg-gray-100"
                @click="userMenuOpen = false"
              >
                <i class="fa-solid fa-bell mr-2 text-gray-500" />消息通知
              </router-link>
              <router-link
                to="/settings"
                class="block px-4 py-2 hover:bg-gray-100"
                @click="userMenuOpen = false"
              >
                <i class="fa-solid fa-gear mr-2 text-gray-500" />个人设置
              </router-link>
              <router-link
                v-if="userStore.isAdmin"
                to="/admin/users"
                class="block px-4 py-2 hover:bg-gray-100"
                @click="userMenuOpen = false"
              >
                <i class="fa-solid fa-gear mr-2 text-gray-500" />管理后台
              </router-link>
              <button
                class="w-full text-left px-4 py-2 hover:bg-gray-100 text-red-600"
                @click="logout"
              >
                <i class="fa-solid fa-arrow-right-from-bracket mr-2" />退出登录
              </button>
            </div>
          </transition>
        </template>
        <template v-else>
          <router-link
            to="/login"
            class="px-4 py-1.5 rounded-full bg-black text-white text-sm hover:bg-gray-800"
          >
            登录 / 注册
          </router-link>
        </template>
      </div>

      <!-- 移动端汉堡 -->
      <button
        class="md:hidden mr-4 p-2 text-gray-700"
        @click="mobileMenuOpen = !mobileMenuOpen"
        aria-label="菜单"
      >
        <i class="fa-solid text-xl" :class="mobileMenuOpen ? 'fa-xmark' : 'fa-bars'" />
      </button>
    </div>

    <!-- 移动端下拉导航 -->
    <div v-if="mobileMenuOpen" class="md:hidden bg-white border-t">
      <nav class="grid grid-cols-4">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="py-3 flex flex-col items-center gap-1 text-sm"
          :class="route.path.startsWith(item.path) ? 'bg-black text-white' : 'text-gray-800'"
          @click="mobileMenuOpen = false"
        >
          <i class="fa-solid text-lg" :class="item.icon" />
          {{ item.label }}
        </router-link>
      </nav>
      <div class="border-t py-2 px-4 flex items-center justify-between">
        <template v-if="userStore.isLogin && currentUser">
          <div class="flex items-center gap-2">
            <img
              :src="avatarUrl(currentUser.avatar)"
              class="w-8 h-8 rounded-full object-cover"
              alt="avatar"
            />
            <span class="text-sm font-bold" :class="GetTextColor(level)">{{
              currentUser.username
            }}</span>
          </div>
          <div class="flex items-center gap-3 text-sm">
            <router-link :to="`/profile/${currentUser.id}`" @click="mobileMenuOpen = false"
              >主页</router-link
            >
            <router-link to="/message" class="relative" @click="mobileMenuOpen = false">
              消息
              <span v-if="unread > 0" class="text-red-500 ml-0.5">{{
                unread > 99 ? '99+' : unread
              }}</span>
            </router-link>
            <router-link to="/settings" @click="mobileMenuOpen = false">设置</router-link>
            <router-link v-if="userStore.isAdmin" to="/admin/users" @click="mobileMenuOpen = false"
              >后台</router-link
            >
            <button class="text-red-600" @click="logout">退出</button>
          </div>
        </template>
        <template v-else>
          <span class="text-sm text-gray-500">未登录</span>
          <router-link
            to="/login"
            class="px-4 py-1.5 rounded-full bg-black text-white text-sm"
            @click="mobileMenuOpen = false"
          >
            登录 / 注册
          </router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
