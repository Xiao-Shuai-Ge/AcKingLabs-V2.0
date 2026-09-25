<script setup lang="ts">
// 帖子管理：关键词/类型/状态筛选，预览、隐藏/恢复、精选/取消、删除（替代 AI 审核队列）
import { onMounted, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import {
  adminPostList,
  adminSetPostHidden,
  adminSetPostFeatured,
  adminDeletePost,
  getPostDetail,
} from '@/api/post'
import type { PostItem, PostDetail } from '@/api/types'
import AdminSidebar from '@/components/AdminSidebar.vue'
import { formatDateTime } from '@/utils/format'
import { postTypeName } from '@/utils/postMeta'
import { useMessage } from '@/composables/useMessage'

const { addMessage, codeHandler } = useMessage()

const collapsed = ref(false)
const keyword = ref('')
const type = ref('')
const status = ref('')
const page = ref(1)
const count = 20
const total = ref(0)
const posts = ref<PostItem[]>([])
const loading = ref(false)

const typeOptions = [
  { value: '', label: '全部类型' },
  { value: 'diary', label: '周记' },
  { value: 'official', label: '官方' },
  { value: 'tutorial', label: '教程' },
  { value: 'solution', label: '题解' },
  { value: 'contest', label: '比赛' },
  { value: 'help', label: '求助' },
  { value: 'fun', label: '闲聊' },
]
const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'hidden', label: '已隐藏' },
  { value: 'featured', label: '精选' },
]

async function load() {
  loading.value = true
  try {
    const res = await adminPostList({
      keyword: keyword.value.trim() || undefined,
      type: type.value || undefined,
      status: status.value || undefined,
      page: page.value,
      count,
    })
    posts.value = res.posts
    total.value = res.total
  } catch (err) {
    codeHandler(err)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

async function toggleHidden(p: PostItem) {
  try {
    await adminSetPostHidden(p.id, !p.is_hidden)
    addMessage(!p.is_hidden ? '已隐藏该帖子' : '已恢复展示', 'success')
    load()
  } catch (err) {
    codeHandler(err)
  }
}

async function toggleFeatured(p: PostItem) {
  try {
    await adminSetPostFeatured(p.id, !p.is_featured)
    addMessage(!p.is_featured ? '已设为精选' : '已取消精选', 'success')
    load()
  } catch (err) {
    codeHandler(err)
  }
}

async function removePost(p: PostItem) {
  try {
    await ElMessageBox.confirm(
      `确定删除帖子《${p.title}》吗？其评论与点赞将一并删除。`,
      '删除帖子',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await adminDeletePost(p.id)
    addMessage('已删除', 'success')
    load()
  } catch (err) {
    codeHandler(err)
  }
}

// 预览
const previewVisible = ref(false)
const previewPost = ref<PostDetail | null>(null)
const previewLoading = ref(false)

async function openPreview(p: PostItem) {
  previewVisible.value = true
  previewLoading.value = true
  try {
    previewPost.value = await getPostDetail(p.id)
  } catch (err) {
    codeHandler(err)
  } finally {
    previewLoading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="pt-[60px] min-h-screen bg-gray-50">
    <div class="flex">
      <AdminSidebar v-model="collapsed" />
      <div class="flex-1 min-w-0 px-4 md:px-8 py-8">
        <button
          class="md:hidden mb-4 px-3 py-1.5 border rounded-lg bg-white text-sm"
          @click="collapsed = false"
        >
          <i class="fa-solid fa-bars mr-1" />菜单
        </button>

        <div class="mb-6">
          <h1 class="text-2xl font-bold">内容管理</h1>
          <p class="text-sm text-gray-500 mt-1">全站帖子管理：隐藏违规内容、设置精选、删除帖子</p>
        </div>

        <!-- 工具栏 -->
        <div
          class="bg-white rounded-lg border border-gray-200 p-4 mb-4 flex flex-col md:flex-row gap-3"
        >
          <el-input
            v-model="keyword"
            placeholder="搜索标题/内容关键词"
            clearable
            class="md:!w-72"
            @keyup.enter="search"
            @clear="search"
          />
          <el-select v-model="type" class="md:!w-36" @change="search">
            <el-option v-for="t in typeOptions" :key="t.value" :value="t.value" :label="t.label" />
          </el-select>
          <el-select v-model="status" class="md:!w-36" @change="search">
            <el-option
              v-for="s in statusOptions"
              :key="s.value"
              :value="s.value"
              :label="s.label"
            />
          </el-select>
          <button
            class="px-4 py-2 rounded-lg bg-black text-white text-sm md:ml-auto"
            @click="search"
          >
            <i class="fa-solid fa-magnifying-glass mr-1" />搜索
          </button>
        </div>

        <!-- 表格 -->
        <div class="bg-white rounded-lg border border-gray-200 overflow-hidden" v-loading="loading">
          <el-table :data="posts" style="width: 100%">
            <el-table-column prop="id" label="ID" width="90" show-overflow-tooltip />
            <el-table-column label="标题" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                <el-link type="primary" @click="openPreview(row)">{{ row.title }}</el-link>
              </template>
            </el-table-column>
            <el-table-column label="作者" width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.author?.username ?? '-' }}</template>
            </el-table-column>
            <el-table-column label="类型" width="80">
              <template #default="{ row }">
                <el-tag size="small">{{ postTypeName(row.type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="数据" width="130">
              <template #default="{ row }">
                <span class="text-xs text-gray-500"
                  >👍{{ row.like_count }} 💬{{ row.comment_count }} 👁{{ row.view_count }}</span
                >
              </template>
            </el-table-column>
            <el-table-column label="状态" width="150">
              <template #default="{ row }">
                <el-tag v-if="row.is_hidden" type="danger" size="small" class="mr-1">已隐藏</el-tag>
                <el-tag v-if="row.is_featured" type="warning" size="small" class="mr-1"
                  >精选</el-tag
                >
                <el-tag v-if="row.is_private" type="info" size="small">私密</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="发布时间" width="160">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="230" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openPreview(row)"
                  >查看</el-button
                >
                <el-button
                  link
                  :type="row.is_hidden ? 'success' : 'warning'"
                  size="small"
                  @click="toggleHidden(row)"
                >
                  {{ row.is_hidden ? '恢复' : '隐藏' }}
                </el-button>
                <el-button
                  link
                  :type="row.is_featured ? 'warning' : 'success'"
                  size="small"
                  @click="toggleFeatured(row)"
                >
                  {{ row.is_featured ? '取消精选' : '精选' }}
                </el-button>
                <el-button link type="danger" size="small" @click="removePost(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="flex justify-center py-4">
            <el-pagination
              layout="total, prev, pager, next"
              :total="total"
              :page-size="count"
              :current-page="page"
              @current-change="
                (p: number) => {
                  page = p
                  load()
                }
              "
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 预览对话框 -->
    <el-dialog
      v-model="previewVisible"
      :title="previewPost?.title ?? '帖子预览'"
      width="720px"
      top="6vh"
      append-to-body
    >
      <div v-loading="previewLoading" class="max-h-[70vh] overflow-y-auto">
        <div v-if="previewPost" class="text-xs text-gray-400 mb-3">
          {{ previewPost.author?.username }} · {{ formatDateTime(previewPost.created_at) }}
        </div>
        <v-md-preview v-if="previewPost" :text="previewPost.content" />
      </div>
    </el-dialog>
  </div>
</template>
