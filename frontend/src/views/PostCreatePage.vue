<script setup lang="ts">
// 帖子发布：打卡模式（标题/周期锁定 + 公开私密）与学习模式（自由标题/类型/来源）
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { createPost } from '@/api/post'
import { uploadImage } from '@/api/file'
import { getContentLimit } from '@/utils/contentLimit'
import { getWeekCode } from '@/utils/week'
import { postMetaMap } from '@/utils/postMeta'
import { useMessage } from '@/composables/useMessage'
import type { UserBrief } from '@/api/user'
import UserSelector from '@/components/UserSelector.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { currentUser } = storeToRefs(userStore)
const { addMessage, codeHandler } = useMessage()

// diary / learn（由路由 meta 决定）
const mode = (route.meta.postType as string) === 'diary' ? 'diary' : 'learn'

const week = getWeekCode(new Date())

const learnTypes = ['tutorial', 'solution', 'contest', 'fun', 'help']
if (userStore.isAdmin) learnTypes.unshift('official')

const form = reactive({
  title: '',
  content: '',
  type: mode === 'diary' ? 'diary' : '',
  source: '',
  isPrivate: false,
})

// 比赛题解预填（从比赛详情页跳转）
const contestPrefill = ref(false)

const limits = computed(() => getContentLimit(currentUser.value?.role ?? 0))
const contentOver = computed(() => form.content.length > limits.value.maxPostLength)

const showUserSelector = ref(false)
const userCandidates = ref<UserBrief[]>([])

// 自定义工具栏：@用户
let currentEditor: any = null
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

const draftKey = mode === 'diary' ? 'draft-diary-content' : 'draft-learn-content'
const draftSaved = ref(false)

// 自动草稿（1s 防抖）
let draftTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => form.content,
  (val) => {
    draftSaved.value = false
    if (draftTimer) clearTimeout(draftTimer)
    draftTimer = setTimeout(() => {
      localStorage.setItem(draftKey, val)
      draftSaved.value = true
    }, 1000)
  },
)

function insertAtCursor(text: string) {
  if (currentEditor?.insert) {
    currentEditor.insert((selected: string) => ({
      text: `${selected}${text}`,
      selected: undefined,
    }))
  } else {
    form.content += text
  }
}

function onAtSelect(user: UserBrief) {
  insertAtCursor(`[@${user.username}](/profile/${user.id})`)
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

async function handleSubmit() {
  if (!form.content.trim()) {
    addMessage('内容不能为空', 'warning')
    return
  }
  if (contentOver.value) {
    addMessage('内容超出长度限制', 'warning')
    return
  }
  if (mode === 'learn' && !form.type) {
    addMessage('请选择帖子类型', 'warning')
    return
  }
  try {
    const res = await createPost({
      title: form.title,
      content: form.content,
      type: mode === 'diary' ? 'diary' : form.type,
      source: mode === 'diary' ? '' : form.source,
      is_private: mode === 'diary' ? form.isPrivate : false,
    })
    localStorage.removeItem(draftKey)
    addMessage('发布成功！', 'success')
    router.push(`/${mode === 'diary' ? 'diary' : 'learn'}/${res.id}`)
  } catch (err) {
    codeHandler(err)
  }
}

onMounted(() => {
  // 恢复草稿
  const draft = localStorage.getItem(draftKey)
  if (draft) {
    form.content = draft
    addMessage('已恢复上次未发布的草稿', 'info')
  }
  // 比赛题解预填
  const contestUrl = route.query.contest_url as string
  const contestTitle = route.query.contest_title as string
  if (mode === 'learn' && contestUrl) {
    contestPrefill.value = true
    form.type = 'solution'
    form.source = contestUrl
    if (contestTitle) form.title = `${contestTitle} 题解`
  }
  if (mode === 'diary') {
    form.title = `${week.name} 学习周记`
  }
})
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-7xl mx-auto px-4 py-8 flex flex-col lg:flex-row gap-6">
      <!-- 编辑区 -->
      <div class="flex-1 min-w-0">
        <h1 class="text-2xl font-bold mb-6">
          {{ mode === 'diary' ? '发布学习周记' : '发布帖子' }}
        </h1>

        <!-- 表单 -->
        <div class="bg-white rounded-lg border border-gray-200 p-5 space-y-4 mb-4">
          <div class="flex flex-col md:flex-row md:items-center gap-3">
            <div class="flex-1">
              <label class="block text-sm font-medium text-gray-700 mb-1">标题</label>
              <el-input
                v-model="form.title"
                :placeholder="
                  mode === 'diary' ? '周记标题（留空自动生成）' : '请输入标题（必填，最多 50 字）'
                "
                maxlength="50"
                :disabled="contestPrefill"
              />
            </div>
            <div v-if="mode === 'diary'" class="md:w-44">
              <label class="block text-sm font-medium text-gray-700 mb-1">时间</label>
              <el-input :model-value="week.name" disabled />
            </div>
            <div v-else class="md:w-44">
              <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
              <el-select
                v-model="form.type"
                placeholder="请选择类型"
                :disabled="contestPrefill"
                class="w-full"
              >
                <el-option
                  v-for="t in learnTypes"
                  :key="t"
                  :value="t"
                  :label="postMetaMap[t]?.name ?? t"
                />
              </el-select>
            </div>
          </div>

          <div v-if="mode === 'learn'" class="flex flex-col md:flex-row gap-3">
            <div class="flex-1">
              <label class="block text-sm font-medium text-gray-700 mb-1"
                >来源 / 网址（选填）</label
              >
              <el-input
                v-model="form.source"
                placeholder="例如比赛链接、参考教程地址"
                maxlength="255"
                :disabled="contestPrefill"
              />
            </div>
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
                公开（双倍经验）
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

        <!-- 编辑器 -->
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

        <div class="flex items-center justify-between mt-3">
          <span class="text-xs" :class="contentOver ? 'text-red-500' : 'text-gray-400'">
            {{ draftSaved ? '草稿已自动保存' : '正在编辑...' }} · {{ form.content.length }} /
            {{ limits.maxPostLength }}
          </span>
        </div>

        <div class="flex gap-3 mt-4">
          <button
            class="px-8 py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800 transition-colors"
            @click="handleSubmit"
          >
            发布帖子
          </button>
          <button
            class="px-8 py-2.5 rounded-lg border-2 border-gray-300 text-gray-600 font-medium hover:bg-gray-100"
            @click="router.back()"
          >
            返回
          </button>
        </div>
      </div>

      <!-- 预览区（桌面端） -->
      <div class="hidden lg:block w-5/12">
        <div class="bg-gray-50 rounded-lg border p-4 sticky top-20">
          <div class="text-xs text-gray-400 mb-2">实时预览</div>
          <div class="text-lg font-bold mb-1">{{ form.title || '帖子标题预览' }}</div>
          <div class="text-xs text-gray-400 mb-3">
            {{ postMetaMap[form.type]?.name ?? (mode === 'diary' ? '周记' : '未选择类型') }} ·
            {{ week.name }}
          </div>
          <div class="h-[70vh] overflow-y-auto bg-white rounded border p-4">
            <v-md-preview :text="form.content" />
          </div>
        </div>
      </div>

      <UserSelector v-model="showUserSelector" :candidates="userCandidates" @select="onAtSelect" />
    </div>
  </div>
</template>
