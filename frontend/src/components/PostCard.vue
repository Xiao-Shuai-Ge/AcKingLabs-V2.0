<script setup lang="ts">
// 帖子卡片：打卡/学习/个人主页/比赛题解列表共用。
// dark 变体 = 打卡/个人主页的黑边框卡片；默认 = 学习页的灰边框卡片。
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { togglePostLike } from '@/api/post'
import { useMessage } from '@/composables/useMessage'
import { CheckLevel, GetTextColor } from '@/utils/level'
import { postTypeName, postTypeClass } from '@/utils/postMeta'
import { avatarUrl, formatDateTime } from '@/utils/format'
import type { PostItem } from '@/api/types'

const props = defineProps<{ post: PostItem; showType?: boolean; index?: number; dark?: boolean }>()

const router = useRouter()
const userStore = useUserStore()
const { codeHandler } = useMessage()

const liked = ref(props.post.liked)
const likeCount = ref(props.post.like_count)
const liking = ref(false)

const authorLevel = computed(() => CheckLevel(props.post.author.xp, props.post.author.role))

const detailPath = computed(
  () => `/${props.post.type === 'diary' ? 'diary' : 'learn'}/${props.post.id}`,
)

function openDetail() {
  router.push(detailPath.value)
}

async function toggleLike() {
  if (!userStore.isLogin) {
    router.push('/login')
    return
  }
  if (liking.value) return
  liking.value = true
  try {
    const res = await togglePostLike(props.post.id)
    liked.value = res.liked
    likeCount.value = res.like_count
  } catch (err) {
    codeHandler(err)
  } finally {
    setTimeout(() => (liking.value = false), 300)
  }
}
</script>

<template>
  <div
    class="bg-white rounded-lg p-4 md:p-5 cursor-pointer transition-all fade-in-up"
    :class="
      dark
        ? 'border-2 border-gray-800 shadow-md hover:shadow-xl hover:scale-105'
        : 'border border-gray-200 hover:shadow-lg hover:scale-[1.01]'
    "
    :style="{ animationDelay: `${(index ?? 0) * 0.06}s` }"
    @click="openDetail"
  >
    <!-- 作者行（头像/用户名可点击进主页） -->
    <div class="flex items-center gap-2 mb-2">
      <router-link
        :to="`/profile/${post.user_id}`"
        class="flex items-center gap-2 min-w-0 hover:opacity-80"
        @click.stop
      >
        <img
          :src="avatarUrl(post.author.avatar)"
          class="w-7 h-7 rounded-full object-cover object-top"
          :class="dark ? 'border-2 border-gray-800' : ''"
          alt=""
        />
        <span class="text-sm font-bold truncate" :class="GetTextColor(authorLevel)">{{
          post.author.username
        }}</span>
      </router-link>
      <span class="text-xs text-gray-400">{{ formatDateTime(post.created_at) }}</span>
    </div>

    <!-- 标题 + 角标（描边式） -->
    <div class="flex items-center flex-wrap gap-2 mb-1.5">
      <h3 class="text-base md:text-lg font-bold truncate max-w-full">{{ post.title }}</h3>
      <span
        v-if="showType !== false"
        class="shrink-0 border-2 rounded-md px-1 text-sm"
        :class="postTypeClass(post.type)"
      >
        {{ postTypeName(post.type) }}
      </span>
      <span
        v-if="post.is_private"
        class="shrink-0 border-2 rounded-md px-1 text-sm text-blue-500 border-blue-500"
        >私密</span
      >
      <span
        v-if="post.is_featured"
        class="shrink-0 border-2 rounded-md px-1 text-sm text-yellow-500 border-yellow-500"
        >精华</span
      >
      <span
        v-if="post.is_admin_like"
        class="shrink-0 border-2 rounded-md px-1 text-sm text-red-500 border-red-500"
        >管理推荐</span
      >
      <span
        v-if="post.is_hidden"
        class="shrink-0 border-2 rounded-md px-1 text-sm text-gray-500 border-gray-500"
        >已隐藏</span
      >
    </div>

    <!-- 摘要 -->
    <p class="text-sm text-gray-600 leading-relaxed line-clamp-2 mb-3">{{ post.content_short }}</p>

    <!-- 数据行 -->
    <div class="flex items-center gap-5 text-sm text-gray-500">
      <button
        class="flex items-center gap-1 hover:text-red-500 transition-colors"
        :class="{ 'text-red-500': liked }"
        @click.stop="toggleLike"
      >
        <i class="fa-solid" :class="liked ? 'fa-heart' : 'fa-regular fa-heart'" />
        <span v-if="likeCount > 0">{{ likeCount }}</span>
      </button>
      <span class="flex items-center gap-1">
        <i class="fa-regular fa-comment-dots" />
        <span v-if="post.comment_count > 0">{{ post.comment_count }}</span>
      </span>
      <span class="flex items-center gap-1">
        <i class="fa-regular fa-eye" />
        <span>{{ post.view_count }}</span>
      </span>
    </div>
  </div>
</template>
