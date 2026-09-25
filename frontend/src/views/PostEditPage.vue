<script setup lang="ts">
// 帖子编辑页
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { getPostDetail, editPost, deletePost } from '@/api/post'
import { uploadImage } from '@/api/file'
import { getContentLimit } from '@/utils/contentLimit'
import { postMetaMap } from '@/utils/postMeta'
import { useMessage } from '@/composables/useMessage'
import type { PostDetail } from '@/api/types'
import type { UserBrief } from '@/api/user'
import UserSelector from '@/components/UserSelector.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { currentUser } = storeToRefs(userStore)
const { addMessage, codeHandler } = useMessage()

const postId = route.params.id as string
const mode = (route.meta.postType as string) === 'diary' ? 'diary' : 'learn'

const post = ref<PostDetail | null>(null)
const loading = ref(true)

const form = reactive({ title: '', content: '', type: '', source: '', isPrivate: false })

const learnTypes = ['tutorial', 'solution', 'contest', 'fun', 'help']
if (userStore.isAdmin) learnTypes.unshift('official')

const limits = computed(() => getContentLimit(currentUser.value?.role ?? 0))
const contentOver = computed(() => form.content.length > limits.value.maxPostLength)

const showUserSelector = ref(false)
const userCandidates = ref<UserBrief[]>([])

const atUserToolbar = {
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

function onAtSelect(user: UserBrief) {
  const text = `[@${user.username}](/profile/${user.id})`
  if (currentEditor?.insert) {
    currentEditor.insert((selected: string) => ({
      text: `${selected}${text}`,
      selected: undefined,
    }))
  } else {
    form.content += text
  }
}

// v-md-editor 的 upload-image 回调签名为 (event, insertImage, files)，文件在第三个参数
async function handleUploadImage(
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

async function handleSave() {
  if (contentOver.value) {
    addMessage('内容超出长度限制', 'warning')
    return
  }
  try {
    await editPost(postId, {
      title: form.title,
      content: form.content,
      type: form.type,
      source: form.source,
      is_private: form.isPrivate,
    })
    addMessage('修改成功！', 'success')
    setTimeout(() => router.replace(`/${mode}/${postId}`), 600)
  } catch (err) {
    codeHandler(err)
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm('确定要删除这篇帖子吗？删除后不可恢复。', '删除帖子', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deletePost(postId)
    addMessage('删除成功', 'success')
    setTimeout(() => router.replace(`/${mode}`), 600)
  } catch (err) {
    codeHandler(err)
  }
}

onMounted(async () => {
  try {
    post.value = await getPostDetail(postId)
    form.title = post.value.title
    form.content = post.value.content
    form.type = post.value.type
    form.source = post.value.source
    form.isPrivate = post.value.is_private
    if (!post.value.can_edit) {
      addMessage('无权编辑这篇帖子', 'warning')
      router.replace(`/${mode}/${postId}`)
    }
  } catch (err) {
    codeHandler(err)
    router.replace(`/${mode}`)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-7xl mx-auto px-4 py-8" v-loading="loading">
      <div class="flex flex-col lg:flex-row gap-6">
        <div class="flex-1 min-w-0">
          <h1 class="text-2xl font-bold mb-6">编辑帖子</h1>

          <div class="bg-white rounded-lg border border-gray-200 p-5 space-y-4 mb-4">
            <div class="flex flex-col md:flex-row md:items-center gap-3">
              <div class="flex-1">
                <label class="block text-sm font-medium text-gray-700 mb-1">标题</label>
                <el-input
                  v-model="form.title"
                  maxlength="50"
                  :placeholder="mode === 'diary' ? '周记标题（留空自动生成）' : '标题（必填）'"
                />
              </div>
              <div v-if="mode === 'diary'" class="md:w-44">
                <label class="block text-sm font-medium text-gray-700 mb-1">时间</label>
                <el-input :model-value="post?.week_code || ''" disabled />
              </div>
              <div v-else class="md:w-44">
                <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
                <el-select v-model="form.type" class="w-full" :disabled="!userStore.isAdmin">
                  <el-option
                    v-for="t in learnTypes"
                    :key="t"
                    :value="t"
                    :label="postMetaMap[t]?.name ?? t"
                  />
                </el-select>
              </div>
            </div>

            <div v-if="mode !== 'diary'" class="flex-1">
              <label class="block text-sm font-medium text-gray-700 mb-1"
                >来源 / 网址（选填）</label
              >
              <el-input v-model="form.source" maxlength="255" />
            </div>

            <div v-if="mode === 'diary'" class="flex items-center gap-4">
              <label class="text-sm font-medium text-gray-700">隐私</label>
              <div class="flex gap-2">
                <button
                  class="px-4 py-1.5 rounded-full border-2 text-sm"
                  :class="
                    !form.isPrivate
                      ? 'bg-black text-white border-black'
                      : 'border-gray-300 text-gray-600'
                  "
                  @click="form.isPrivate = false"
                >
                  公开
                </button>
                <button
                  class="px-4 py-1.5 rounded-full border-2 text-sm"
                  :class="
                    form.isPrivate
                      ? 'bg-black text-white border-black'
                      : 'border-gray-300 text-gray-600'
                  "
                  @click="form.isPrivate = true"
                >
                  私密
                </button>
              </div>
            </div>
          </div>

          <v-md-editor
            v-model="form.content"
            mode="edit"
            height="420px"
            left-toolbar="undo redo clear | h bold italic strikethrough quote | ul ol table hr | link image code | at-user"
            :disabled-menus="[]"
            right-toolbar="toc"
            :toolbar="atUserToolbar"
            @upload-image="handleUploadImage"
          />

          <div class="text-xs mt-2" :class="contentOver ? 'text-red-500' : 'text-gray-400'">
            {{ form.content.length }} / {{ limits.maxPostLength }}
          </div>

          <div class="flex gap-3 mt-4 mb-8">
            <button
              class="px-8 py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800"
              @click="handleSave"
            >
              保存修改
            </button>
            <button
              v-if="post?.can_delete"
              class="px-8 py-2.5 rounded-lg border-2 border-red-300 text-red-500 font-medium hover:bg-red-50"
              @click="handleDelete"
            >
              删除
            </button>
            <button
              class="px-8 py-2.5 rounded-lg border-2 border-gray-300 text-gray-600 font-medium hover:bg-gray-100"
              @click="router.back()"
            >
              返回
            </button>
          </div>
        </div>

        <div class="hidden lg:block w-5/12">
          <div class="bg-gray-50 rounded-lg border p-4 sticky top-20">
            <div class="text-xs text-gray-400 mb-2">实时预览</div>
            <div class="text-lg font-bold mb-3">{{ form.title || '帖子标题预览' }}</div>
            <div class="h-[70vh] overflow-y-auto bg-white rounded border p-4">
              <v-md-preview :text="form.content" />
            </div>
          </div>
        </div>
      </div>

      <UserSelector v-model="showUserSelector" :candidates="userCandidates" @select="onAtSelect" />
    </div>
  </div>
</template>
