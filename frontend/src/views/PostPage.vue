<script setup lang="ts">
// 帖子详情：正文渲染 + 点赞 + 两级评论（回复/删除/点赞/表情包/优质解答）
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import {
  getPostDetail,
  togglePostLike,
  getComments,
  createComment,
  deleteComment,
  toggleCommentLike,
} from '@/api/post'
import { uploadImage } from '@/api/file'
import { getContentLimit } from '@/utils/contentLimit'
import { CheckLevel, GetTextColor } from '@/utils/level'
import { postTypeName, postTypeClass } from '@/utils/postMeta'
import { avatarUrl, formatDateTime } from '@/utils/format'
import { useMessage } from '@/composables/useMessage'
import type { PostDetail, CommentItem } from '@/api/types'
import type { UserBrief } from '@/api/user'
import UserSelector from '@/components/UserSelector.vue'
import StickerPicker from '@/components/StickerPicker.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { currentUser } = storeToRefs(userStore)
const { addMessage, codeHandler } = useMessage()

const postId = route.params.id as string
const mode = (route.meta.postType as string) === 'diary' ? 'diary' : 'learn'

const post = ref<PostDetail | null>(null)
const loading = ref(true)
const liking = ref(false)

// ---- 评论状态 ----
interface CommentNode extends CommentItem {
  children: CommentItem[]
  childrenNoMore: boolean
}

const comments = ref<CommentNode[]>([])
const commentLoading = ref(false)
const commentContent = ref('')
const replyTarget = ref<{ comment: CommentItem | null; fatherId: string } | null>(null)
const commentOver = computed(
  () =>
    commentContent.value.length > getContentLimit(currentUser.value?.role ?? 0).maxCommentLength,
)
const submitting = ref(false)

const showUserSelector = ref(false)
const showStickerPicker = ref(false)
const userCandidates = computed<UserBrief[]>(
  () => [post.value?.author, ...comments.value.map((c) => c.author)].filter(Boolean) as UserBrief[],
)

const commentToolbar = {
  stickers: {
    title: '表情包',
    icon: 'fa-solid fa-face-smile',
    action() {
      showStickerPicker.value = true
    },
  },
  'at-user': {
    title: '@用户',
    icon: 'fas fa-at',
    action(editor: any) {
      currentEditor = editor
      showUserSelector.value = true
    },
  },
}
let currentEditor: any = null

function insertText(text: string) {
  if (currentEditor?.insert) {
    currentEditor.insert((selected: string) => ({
      text: `${selected}${text}`,
      selected: undefined,
    }))
  } else {
    commentContent.value += text
  }
}

// v-md-editor 的 upload-image 回调签名为 (event, insertImage, files)，文件在第三个参数
async function handleCommentUploadImage(
  _event: unknown,
  insertImage: (img: { url: string; desc?: string }) => void,
  files: FileList | File[] | undefined,
) {
  const list = Array.from(files ?? [])
  if (list.length === 0) return
  for (const file of list) {
    try {
      const url = await uploadImage(file)
      insertImage({ url, desc: '图片描述' })
    } catch (err) {
      addMessage((err as Error).message || '图片上传失败', 'error')
    }
  }
}

function onStickerSelect(url: string) {
  insertText(`![sticker](${url})`)
}

function onAtSelect(user: UserBrief) {
  insertText(`[@${user.username}](/profile/${user.id})`)
}

// ---- 加载 ----

async function loadPost() {
  try {
    post.value = await getPostDetail(postId)
  } catch (err) {
    codeHandler(err)
    setTimeout(() => router.replace(`/${mode}`), 1200)
  }
}

async function loadComments(reset = true) {
  if (commentLoading.value) return
  commentLoading.value = true
  try {
    const before =
      reset || comments.value.length === 0 ? '' : comments.value[comments.value.length - 1].id
    const res = await getComments({ post_id: postId, before, count: 10 })
    // 子评论由接口内嵌返回，默认展开
    const nodes = res.comments.map((c) => ({
      ...c,
      children: c.children ?? [],
      childrenNoMore: !c.children_more,
    }))
    if (reset) comments.value = nodes
    else comments.value.push(...nodes)
    commentsNoMore.value = res.comments.length < 10
  } catch (err) {
    codeHandler(err)
  } finally {
    commentLoading.value = false
  }
}
const commentsNoMore = ref(false)

async function loadChildren(node: CommentNode) {
  const after = node.children.length === 0 ? '' : node.children[node.children.length - 1].id
  const res = await getComments({ post_id: postId, father_id: node.id, after, count: 20 })
  node.children.push(...res.comments)
  node.childrenNoMore = res.comments.length < 20
}

// ---- 互动 ----

async function likePost() {
  if (!userStore.isLogin) {
    router.push('/login')
    return
  }
  if (liking.value || !post.value) return
  liking.value = true
  try {
    const res = await togglePostLike(postId)
    if (post.value) {
      post.value.liked = res.liked
      post.value.like_count = res.like_count
    }
  } catch (err) {
    codeHandler(err)
  } finally {
    setTimeout(() => (liking.value = false), 300)
  }
}

function startReply(comment: CommentItem) {
  // 顶层评论：father = 该评论；子评论：father = 其顶层，reply_to = 被回复评论
  if (comment.father_id === '0') {
    replyTarget.value = { comment: null, fatherId: comment.id }
  } else {
    replyTarget.value = { comment, fatherId: comment.father_id }
  }
  document.getElementById('comment-box')?.scrollIntoView({ behavior: 'smooth' })
}

function cancelReply() {
  replyTarget.value = null
}

async function submitComment() {
  if (!userStore.isLogin) {
    router.push('/login')
    return
  }
  if (!commentContent.value.trim() || commentOver.value || submitting.value) return
  submitting.value = true
  try {
    const target = replyTarget.value
    await createComment({
      post_id: postId,
      father_id: target?.fatherId ?? '0',
      reply_to_id: target?.comment?.id ?? '0',
      content: commentContent.value,
    })
    addMessage('评论发布成功', 'success')
    commentContent.value = ''
    replyTarget.value = null
    if (post.value) post.value.comment_count++
    await loadComments(true)
  } catch (err) {
    codeHandler(err)
  } finally {
    submitting.value = false
  }
}

async function likeComment(node: CommentNode | CommentItem) {
  if (!userStore.isLogin) {
    router.push('/login')
    return
  }
  try {
    const res = await toggleCommentLike(node.id)
    node.liked = res.liked
    node.like_count = res.like_count
  } catch (err) {
    codeHandler(err)
  }
}

async function removeComment(node: CommentNode | CommentItem) {
  try {
    await deleteComment(node.id)
    addMessage('评论已删除', 'success')
    if (post.value) post.value.comment_count--
    await loadComments(true)
  } catch (err) {
    codeHandler(err)
  }
}

const authorLevel = computed(() =>
  post.value ? CheckLevel(post.value.author.xp, post.value.author.role) : 0,
)

onMounted(async () => {
  await loadPost()
  loading.value = false
  await loadComments(true)
})
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-4xl mx-auto px-4 py-8" v-loading="loading">
      <template v-if="post">
        <!-- 作者信息栏 -->
        <div class="bg-white rounded-lg border border-gray-200 p-4 mb-4 flex items-center gap-3">
          <router-link
            :to="`/profile/${post.user_id}`"
            class="flex items-center gap-3 min-w-0 hover:opacity-80"
          >
            <img
              :src="avatarUrl(post.author.avatar)"
              class="w-11 h-11 rounded-full object-cover"
              alt=""
            />
            <div class="min-w-0">
              <div class="font-bold" :class="GetTextColor(authorLevel)">
                {{ post.author.username }}
              </div>
              <div class="text-xs text-gray-400">{{ formatDateTime(post.created_at) }}</div>
            </div>
          </router-link>
          <div class="ml-auto flex items-center gap-2">
            <button
              v-if="post.can_edit"
              class="px-3 py-1.5 rounded-lg border border-gray-300 text-sm hover:bg-gray-100"
              @click="router.push(`/${mode}/${post.id}/edit`)"
            >
              编辑
            </button>
          </div>
        </div>

        <!-- 正文卡片 -->
        <div class="bg-white rounded-lg border border-gray-200 p-5 md:p-8 mb-6">
          <div class="flex items-center flex-wrap gap-2 mb-3">
            <h1 class="text-xl md:text-2xl font-bold">{{ post.title }}</h1>
            <span class="px-2 py-0.5 rounded text-xs font-medium" :class="postTypeClass(post.type)">
              {{ postTypeName(post.type) }}
            </span>
            <span
              v-if="post.is_private"
              class="px-2 py-0.5 rounded text-xs bg-blue-100 text-blue-600"
              >私密</span
            >
            <span
              v-if="post.is_featured"
              class="px-2 py-0.5 rounded text-xs bg-yellow-100 text-yellow-700"
              >精华</span
            >
            <span
              v-if="post.is_admin_like"
              class="px-2 py-0.5 rounded text-xs bg-red-100 text-red-600"
              >管理推荐</span
            >
          </div>

          <div v-if="post.source" class="mb-4 text-sm text-gray-500 truncate">
            来源：
            <a
              :href="post.source"
              target="_blank"
              rel="noopener"
              class="text-blue-500 hover:underline"
            >
              {{ post.source }}
            </a>
          </div>

          <v-md-preview :text="post.content" class="post-content" />

          <!-- 底部操作栏 -->
          <div class="flex items-center gap-6 mt-8 pt-4 border-t text-sm text-gray-500">
            <button
              class="flex items-center gap-1.5 hover:text-red-500 transition-colors"
              :class="{ 'text-red-500': post.liked }"
              @click="likePost"
            >
              <i
                class="fa-solid text-lg"
                :class="post.liked ? 'fa-heart' : 'fa-regular fa-heart'"
              />
              点赞 {{ post.like_count > 0 ? post.like_count : '' }}
            </button>
            <span class="flex items-center gap-1.5">
              <i class="fa-regular fa-comment-dots text-lg" />
              评论 {{ post.comment_count > 0 ? post.comment_count : '' }}
            </span>
            <span class="flex items-center gap-1.5">
              <i class="fa-regular fa-eye text-lg" />
              浏览 {{ post.view_count }}
            </span>
          </div>
        </div>

        <!-- 评论区 -->
        <div id="comment-box" class="bg-white rounded-lg border border-gray-200 p-5 md:p-6">
          <h2 class="font-bold mb-5">评论 ({{ post.comment_count }})</h2>

          <!-- 未登录引导 -->
          <div
            v-if="!userStore.isLogin"
            class="mb-6 py-8 rounded-lg border-2 border-dashed border-gray-200 flex flex-col items-center gap-3"
          >
            <span class="text-sm text-gray-400">登录后即可参与评论</span>
            <router-link
              to="/login"
              class="px-6 py-2 rounded-lg bg-black text-white text-sm font-medium hover:bg-gray-800"
            >
              去登录
            </router-link>
          </div>

          <!-- 评论输入 -->
          <template v-else>
            <!-- 回复提示 -->
            <div
              v-if="replyTarget"
              class="mb-3 px-3 py-2 rounded-lg bg-blue-50 border border-blue-200 text-sm flex items-center justify-between gap-2"
            >
              <span class="truncate">
                回复 <b>{{ replyTarget.comment?.author.username }}</b> ：{{
                  replyTarget.comment?.content.slice(0, 50)
                }}
              </span>
              <button class="text-gray-400 hover:text-gray-600 shrink-0" @click="cancelReply">
                <i class="fa-solid fa-xmark" />
              </button>
            </div>

            <!-- 编辑器（去掉自带边框，由外层卡片统一提供） -->
            <div
              class="comment-editor-card rounded-lg border border-gray-300 overflow-hidden focus-within:border-gray-500 transition-colors"
            >
              <v-md-editor
                v-model="commentContent"
                mode="edit"
                height="180px"
                left-toolbar="bold italic quote | link image code | stickers at-user"
                :disabled-menus="[]"
                right-toolbar=""
                :toolbar="commentToolbar"
                @upload-image="handleCommentUploadImage"
              />
            </div>
            <div class="flex items-center justify-between mt-3 mb-6">
              <span class="text-xs" :class="commentOver ? 'text-red-500' : 'text-gray-400'">
                {{ commentContent.length }} /
                {{ getContentLimit(currentUser?.role ?? 0).maxCommentLength }}
              </span>
              <button
                class="px-6 py-2 rounded-lg bg-black text-white text-sm font-medium hover:bg-gray-800 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                :disabled="!commentContent.trim() || commentOver || submitting"
                @click="submitComment"
              >
                {{ submitting ? '发布中...' : '发布评论' }}
              </button>
            </div>
          </template>

          <!-- 评论列表 -->
          <div class="mt-6 space-y-4">
            <div
              v-if="comments.length === 0 && !commentLoading"
              class="py-10 text-center text-sm text-gray-400"
            >
              还没有评论，快来抢占一楼吧！
            </div>

            <div v-for="node in comments" :key="node.id" class="fade-in-up">
              <!-- 顶层评论 -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center gap-2 mb-2">
                  <router-link
                    :to="`/profile/${node.user_id}`"
                    class="flex items-center gap-2 min-w-0 hover:opacity-80"
                  >
                    <img
                      :src="avatarUrl(node.author.avatar)"
                      class="w-8 h-8 rounded-full object-cover"
                      alt=""
                    />
                    <span
                      class="text-sm font-bold"
                      :class="GetTextColor(CheckLevel(node.author.xp, node.author.role))"
                    >
                      {{ node.author.username }}
                    </span>
                  </router-link>
                  <span
                    v-if="node.is_admin_like"
                    class="px-1.5 py-0.5 rounded text-[11px] bg-orange-100 text-orange-600"
                  >
                    优质解答
                  </span>
                  <span class="text-xs text-gray-400 ml-auto">{{
                    formatDateTime(node.created_at)
                  }}</span>
                </div>
                <v-md-preview :text="node.content" />
                <div class="flex items-center gap-4 mt-2 text-sm text-gray-500">
                  <button
                    class="hover:text-red-500 flex items-center gap-1"
                    :class="{ 'text-red-500': node.liked }"
                    @click="likeComment(node)"
                  >
                    <i class="fa-solid" :class="node.liked ? 'fa-heart' : 'fa-regular fa-heart'" />
                    <span v-if="node.like_count > 0">{{ node.like_count }}</span>
                  </button>
                  <button class="hover:text-blue-500" @click="startReply(node)">回复</button>
                  <button
                    v-if="node.can_delete"
                    class="hover:text-red-500 ml-auto"
                    @click="removeComment(node)"
                  >
                    删除
                  </button>
                </div>

                <!-- 子评论（默认展开） -->
                <div v-if="node.children.length > 0" class="mt-4 ml-6 md:ml-10 space-y-3">
                  <div
                    v-for="child in node.children"
                    :key="child.id"
                    class="border-l-2 border-gray-200 pl-4 py-1"
                  >
                    <div class="flex items-center gap-2 mb-1">
                      <router-link
                        :to="`/profile/${child.user_id}`"
                        class="flex items-center gap-2 min-w-0 hover:opacity-80"
                      >
                        <img
                          :src="avatarUrl(child.author.avatar)"
                          class="w-6 h-6 rounded-full object-cover"
                          alt=""
                        />
                        <span
                          class="text-sm font-bold"
                          :class="GetTextColor(CheckLevel(child.author.xp, child.author.role))"
                        >
                          {{ child.author.username }}
                        </span>
                      </router-link>
                      <span v-if="child.reply_to" class="text-xs text-gray-400"
                        >回复 @{{ child.reply_to }}</span
                      >
                      <span
                        v-if="child.is_admin_like"
                        class="px-1.5 py-0.5 rounded text-[11px] bg-orange-100 text-orange-600"
                      >
                        优质解答
                      </span>
                      <span class="text-xs text-gray-400 ml-auto">{{
                        formatDateTime(child.created_at)
                      }}</span>
                    </div>
                    <v-md-preview :text="child.content" />
                    <div class="flex items-center gap-4 mt-1 text-sm text-gray-500">
                      <button
                        class="hover:text-red-500 flex items-center gap-1"
                        :class="{ 'text-red-500': child.liked }"
                        @click="likeComment(child)"
                      >
                        <i
                          class="fa-solid"
                          :class="child.liked ? 'fa-heart' : 'fa-regular fa-heart'"
                        />
                        <span v-if="child.like_count > 0">{{ child.like_count }}</span>
                      </button>
                      <button class="hover:text-blue-500" @click="startReply(child)">回复</button>
                      <button
                        v-if="child.can_delete"
                        class="hover:text-red-500 ml-auto"
                        @click="removeComment(child)"
                      >
                        删除
                      </button>
                    </div>
                  </div>
                  <button
                    v-if="!node.childrenNoMore"
                    class="text-sm text-blue-500 hover:underline"
                    @click="loadChildren(node)"
                  >
                    加载更多回复
                  </button>
                </div>
              </div>
            </div>

            <button
              v-if="!commentsNoMore && comments.length > 0"
              class="w-full py-2 rounded-lg border border-gray-300 text-sm text-gray-600 hover:bg-gray-50"
              @click="loadComments(false)"
            >
              加载更多评论
            </button>
          </div>
        </div>
      </template>
    </div>

    <UserSelector v-model="showUserSelector" :candidates="userCandidates" @select="onAtSelect" />
    <StickerPicker v-model="showStickerPicker" @select="onStickerSelect" />
  </div>
</template>

<style scoped>
.comment-editor-card :deep(.v-md-editor) {
  border: none;
  box-shadow: none;
}
</style>
