<script setup lang="ts">
// 学习页：关键词搜索 + 类型/排序筛选 + 分页（列表接口内嵌作者与点赞态，无 N+1）
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getPostList, searchPosts } from '@/api/post'
import type { PostItem } from '@/api/types'
import PostCard from '@/components/PostCard.vue'
import { useMessage } from '@/composables/useMessage'

const route = useRoute()
const router = useRouter()
const { addMessage } = useMessage()

const typeOptions = [
  { key: 'all', label: '全部' },
  { key: 'official', label: '官方' },
  { key: 'tutorial', label: '教程' },
  { key: 'solution', label: '题解' },
  { key: 'contest', label: '比赛' },
  { key: 'help', label: '求助' },
  { key: 'fun', label: '闲聊' },
]
const sortOptions = [
  { key: 'hot', label: '热度' },
  { key: 'new', label: '时间' },
  { key: 'featured', label: '精华' },
]

const keyword = ref('')
const type = ref('all')
const sort = ref('hot')
const page = ref(1)
const count = 5

const posts = ref<PostItem[]>([])
const total = ref(0)
const loading = ref(false)
const searching = ref(false)

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    let res
    if (searching.value && keyword.value.trim()) {
      res = await searchPosts({ keyword: keyword.value.trim(), page: page.value, count })
    } else {
      res = await getPostList({ type: type.value, sort: sort.value, page: page.value, count })
    }
    posts.value = res.posts
    total.value = res.total
  } catch {
    addMessage('加载失败', 'error')
  } finally {
    loading.value = false
  }
}

function syncUrl() {
  router.replace({
    query: {
      ...(keyword.value.trim() ? { search: keyword.value.trim() } : {}),
      ...(type.value !== 'all' ? { type: type.value } : {}),
      ...(sort.value !== 'hot' ? { sort: sort.value } : {}),
      ...(page.value > 1 ? { page: String(page.value) } : {}),
    },
  })
}

function doSearch() {
  searching.value = keyword.value.trim() !== ''
  page.value = 1
  syncUrl()
  load()
}

function pickType(t: string) {
  type.value = t
  page.value = 1
  syncUrl()
  load()
}

function pickSort(s: string) {
  sort.value = s
  page.value = 1
  syncUrl()
  load()
}

function changePage(p: number) {
  page.value = p
  syncUrl()
  load()
}

onMounted(() => {
  // 单一初始化入口（只加载一次）
  if (route.query.search) keyword.value = route.query.search as string
  if (route.query.type) type.value = route.query.type as string
  if (route.query.sort) sort.value = route.query.sort as string
  if (route.query.page) page.value = Number(route.query.page) || 1
  searching.value = !!route.query.search
  load()
})
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-4xl mx-auto px-4 py-8">
      <!-- 筛选卡片 -->
      <div class="bg-white rounded-lg border border-gray-200 p-5 mb-6">
        <h2 class="font-bold mb-4">搜索</h2>
        <div class="flex gap-2 mb-5">
          <el-input
            v-model="keyword"
            placeholder="请输入要搜索的关键词"
            clearable
            @keyup.enter="doSearch"
            @clear="doSearch"
          />
          <button
            class="shrink-0 px-6 rounded-lg bg-black text-white text-sm font-medium hover:bg-gray-800"
            @click="doSearch"
          >
            搜索
          </button>
        </div>

        <div class="space-y-3">
          <div class="flex items-center flex-wrap gap-2">
            <span class="text-sm text-gray-500 w-16 shrink-0">帖子类型</span>
            <button
              v-for="t in typeOptions"
              :key="t.key"
              class="px-3 py-1 rounded-full text-sm border transition-colors"
              :class="
                type === t.key
                  ? 'bg-black text-white border-black'
                  : 'border-gray-300 text-gray-600 hover:border-gray-500'
              "
              @click="pickType(t.key)"
            >
              {{ t.label }}
            </button>
          </div>
          <div v-if="!searching" class="flex items-center flex-wrap gap-2">
            <span class="text-sm text-gray-500 w-16 shrink-0">排序方式</span>
            <button
              v-for="s in sortOptions"
              :key="s.key"
              class="px-3 py-1 rounded-full text-sm border transition-colors"
              :class="
                sort === s.key
                  ? 'bg-black text-white border-black'
                  : 'border-gray-300 text-gray-600 hover:border-gray-500'
              "
              @click="pickSort(s.key)"
            >
              {{ s.label }}
            </button>
          </div>
        </div>
      </div>

      <!-- 帖子列表 -->
      <div class="space-y-4" v-loading="loading">
        <PostCard v-for="(p, i) in posts" :key="p.id" :post="p" :index="i % count" />
        <div
          v-if="posts.length === 0 && !loading"
          class="py-16 text-center text-gray-400 bg-white rounded-lg border border-gray-200"
        >
          {{ searching ? '没有找到相关帖子' : '这里还没有帖子，快来发布第一篇吧！' }}
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="total > count" class="mt-6 flex justify-center">
        <el-pagination
          layout="prev, pager, next"
          :total="total"
          :page-size="count"
          :pager-count="9"
          :current-page="page"
          @current-change="changePage"
        />
      </div>
    </div>

    <!-- 发布按钮 -->
    <router-link
      to="/learn/create"
      class="fixed bottom-8 right-8 z-40 w-14 h-14 rounded-full bg-blue-500 text-white shadow-lg flex items-center justify-center text-2xl hover:bg-blue-600 hover:scale-105 transition-all"
      aria-label="发布帖子"
    >
      <i class="fa-solid fa-pen-nib" />
    </router-link>
  </div>
</template>
