<script setup lang="ts">
// 个人主页：资料展示 + 编辑（本人/管理员）+ 经验/CF/获奖 + 打卡/帖子列表
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { getProfile, setProfile } from '@/api/user'
import { getPostList } from '@/api/post'
import { uploadImage } from '@/api/file'
import { CheckLevel, GetTextColor, GetBgColor, GetRoleName, NextLevelLimit } from '@/utils/level'
import { avatarUrl } from '@/utils/format'
import { useMessage } from '@/composables/useMessage'
import type { UserProfile } from '@/api/user'
import type { PostItem } from '@/api/types'
import PostCard from '@/components/PostCard.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { currentUser } = storeToRefs(userStore)
const { addMessage, codeHandler } = useMessage()

const userId = route.params.id as string
const profile = ref<UserProfile | null>(null)
const loading = ref(true)

const isSelf = computed(() => currentUser.value?.id === userId)
const canEdit = computed(() => isSelf.value || userStore.isAdmin)
const level = computed(() => CheckLevel(profile.value?.xp ?? 0, profile.value?.role ?? 0))
const nextXp = computed(() => NextLevelLimit(profile.value?.xp ?? 0, profile.value?.role ?? 0))
const xpPercent = computed(() => {
  if (!profile.value) return 0
  return Math.min(100, Math.round((profile.value.xp / Math.max(nextXp.value, 1)) * 100))
})

// CF 分数颜色
function cfColor(rating: number): string {
  if (rating <= 0) return 'text-gray-500'
  if (rating < 1200) return 'text-gray-600'
  if (rating < 1400) return 'text-green-600'
  if (rating < 1600) return 'text-cyan-600'
  if (rating < 1900) return 'text-blue-600'
  if (rating < 2100) return 'text-purple-600'
  if (rating < 2400) return 'text-yellow-500'
  return 'text-red-500'
}

const medal = ['', '🥇', '🥈', '🥉']

// ---- 帖子标签页 ----
const tab = ref<'diary' | 'learn'>('diary')
const posts = ref<PostItem[]>([])
const postTotal = ref(0)
const postPage = ref(1)
const count = 3

async function loadPosts() {
  try {
    const res = await getPostList({
      type: tab.value === 'diary' ? 'diary' : 'all',
      sort: 'new',
      user_id: userId,
      page: postPage.value,
      count,
    })
    posts.value = res.posts
    postTotal.value = res.total
  } catch {
    /* 静默 */
  }
}

function switchTab(t: 'diary' | 'learn') {
  if (tab.value === t) return
  tab.value = t
  postPage.value = 1
  loadPosts()
}

// ---- 编辑 ----
const editVisible = ref(false)
const editForm = reactive({
  username: '',
  signature: '',
  codeforces_id: '',
  avatar: '',
  real_name: '',
  grade: 0,
  student_no: 0 as number | string,
  role: 1,
  awards: [] as { name: string; level: number }[],
})

function openEdit() {
  if (!profile.value) return
  editForm.username = profile.value.username
  editForm.signature = profile.value.signature
  editForm.codeforces_id = profile.value.codeforces_id
  editForm.avatar = profile.value.avatar
  editForm.real_name = profile.value.real_name ?? ''
  editForm.grade = profile.value.grade
  editForm.student_no = profile.value.student_no ?? ''
  editForm.role = profile.value.role
  editForm.awards = profile.value.awards?.map((a) => ({ ...a })) ?? []
  editVisible.value = true
}

async function handleUploadAvatar(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    editForm.avatar = await uploadImage(file)
  } catch (err) {
    addMessage((err as Error).message || '上传失败', 'error')
  } finally {
    input.value = ''
  }
}

function addAward() {
  editForm.awards.push({ name: '', level: 1 })
}
function removeAward(i: number) {
  editForm.awards.splice(i, 1)
}

async function saveEdit() {
  try {
    await setProfile({
      id: userId,
      username: editForm.username,
      signature: editForm.signature,
      codeforces_id: editForm.codeforces_id,
      avatar: editForm.avatar,
      awards: editForm.awards.filter((a) => a.name.trim() !== ''),
    })
    addMessage('保存成功', 'success')
    editVisible.value = false
    const res = await getProfile(userId)
    profile.value = res
    if (isSelf.value) userStore.setUser({ ...res })
  } catch (err) {
    codeHandler(err)
  }
}

onMounted(async () => {
  try {
    profile.value = await getProfile(userId)
  } catch (err) {
    codeHandler(err)
    setTimeout(() => router.replace('/'), 1000)
  } finally {
    loading.value = false
  }
  loadPosts()
})
</script>

<template>
  <div class="pt-[60px] min-h-screen bg-gray-50" v-loading="loading">
    <div class="max-w-4xl mx-auto px-4 pb-16" v-if="profile">
      <!-- 居中头像区 -->
      <div class="relative w-full flex flex-col items-center mt-10 mb-8">
        <div
          class="w-48 h-48 rounded-full overflow-hidden border-4 border-gray-800 shadow-lg group"
        >
          <img
            :src="avatarUrl(profile.avatar)"
            class="w-full h-full object-cover object-top transition-transform duration-300 group-hover:scale-105"
            alt="avatar"
          />
        </div>
        <button
          v-if="canEdit"
          class="absolute top-28 right-[calc(50%-6rem)] w-9 h-9 rounded-full bg-white border-2 border-gray-800 flex items-center justify-center hover:bg-gray-100 shadow"
          title="更换头像"
          @click="openEdit"
        >
          <i class="fa-solid fa-camera text-gray-700" />
        </button>

        <div class="flex items-center gap-3 mt-5 flex-wrap justify-center">
          <h1 class="text-2xl font-bold" :class="GetTextColor(level)">{{ profile.username }}</h1>
          <span class="px-2 py-0.5 rounded text-xs text-white" :class="GetBgColor(level)">
            {{ GetRoleName(level) }}
          </span>
        </div>
        <p class="text-gray-500 mt-2 text-center max-w-xl">
          {{ profile.signature || '此人很懒什么也没写' }}
        </p>

        <button
          v-if="canEdit"
          class="mt-4 px-5 py-2 rounded-lg border-2 border-gray-800 text-sm font-medium hover:bg-gray-100"
          @click="openEdit"
        >
          <i class="fa-solid fa-pen mr-1.5" />{{ isSelf ? '编辑资料' : '编辑（管理员）' }}
        </button>
      </div>

      <!-- 资料卡片（黑边框） -->
      <div class="w-full bg-white border-2 border-gray-800 rounded-lg p-6 shadow-lg mb-6">
        <div class="mt-1">
          <div class="flex items-center justify-between text-sm mb-1">
            <span class="text-gray-500">
              经验值：{{ profile.xp }}（目前身份：<span :class="GetTextColor(level)">{{
                GetRoleName(level)
              }}</span
              >）
            </span>
            <span class="text-gray-400 text-xs">下一等级 {{ nextXp }} XP</span>
          </div>
          <div class="h-2.5 bg-gray-200 rounded-full overflow-hidden">
            <div
              class="h-full rounded-full transition-all"
              :class="GetBgColor(level)"
              :style="{ width: `${xpPercent}%` }"
            />
          </div>
        </div>

        <div class="mt-5 grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
          <div v-if="profile.codeforces_id" class="flex items-center gap-2 flex-wrap">
            <span class="text-gray-500">Codeforces：</span>
            <a
              :href="`https://codeforces.com/profile/${profile.codeforces_id}`"
              target="_blank"
              rel="noopener"
              class="text-blue-500 hover:underline"
            >
              {{ profile.codeforces_id }}
            </a>
            <span
              v-if="profile.codeforces_rating > 0"
              class="font-bold"
              :class="cfColor(profile.codeforces_rating)"
            >
              {{ profile.codeforces_rating }}
            </span>
          </div>
          <div v-if="profile.grade" class="text-gray-500">年级：{{ profile.grade }} 级</div>
          <div v-if="profile.real_name" class="text-gray-500">
            真实姓名：{{ profile.real_name }}
          </div>
          <div v-if="profile.student_no" class="text-gray-500">学号：{{ profile.student_no }}</div>
        </div>

        <!-- 获奖经历 -->
        <div v-if="profile.awards?.length" class="mt-6 pt-6 border-t">
          <h3 class="font-bold mb-4">🏆 获奖经历</h3>
          <div class="grid grid-cols-1 gap-2">
            <div
              v-for="(a, i) in profile.awards"
              :key="i"
              class="flex items-center gap-2 text-sm bg-gray-50 rounded-lg border-2 border-gray-200 p-3"
            >
              <span class="text-lg">{{ medal[a.level] ?? '🏅' }}</span>
              <span>{{ a.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 帖子标签页 -->
      <div class="flex gap-2 mb-6 justify-center">
        <button
          v-for="t in [
            { key: 'diary', label: '打卡周记' },
            { key: 'learn', label: '帖子' },
          ]"
          :key="t.key"
          class="px-5 py-1.5 rounded-full border-2 text-sm font-medium transition-colors"
          :class="
            tab === t.key
              ? 'bg-black text-white border-black'
              : 'border-gray-800 text-gray-700 hover:bg-gray-100'
          "
          @click="switchTab(t.key as 'diary' | 'learn')"
        >
          {{ t.label }}
        </button>
      </div>

      <div class="space-y-6">
        <PostCard
          v-for="(p, i) in posts"
          :key="p.id"
          :post="p"
          :index="i"
          :show-type="tab === 'learn'"
          dark
        />
        <div
          v-if="posts.length === 0"
          class="py-14 text-center text-gray-400 bg-white rounded-lg border-2 border-gray-800"
        >
          暂无{{ tab === 'diary' ? '周记' : '帖子' }}
        </div>
      </div>

      <div v-if="postTotal > count" class="mt-8 flex justify-center">
        <el-pagination
          layout="prev, pager, next"
          :total="postTotal"
          :page-size="count"
          :current-page="postPage"
          @current-change="
            (p: number) => {
              postPage = p
              loadPosts()
            }
          "
        />
      </div>
    </div>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editVisible" title="编辑资料" width="520px" append-to-body>
      <el-form label-width="100px">
        <el-form-item label="头像">
          <div class="flex items-center gap-3 w-full">
            <img
              :src="avatarUrl(editForm.avatar)"
              class="w-12 h-12 rounded-full object-cover border-2 border-gray-800"
              alt=""
            />
            <el-input v-model="editForm.avatar" placeholder="头像链接" class="flex-1" />
            <label
              class="shrink-0 px-3 py-1.5 rounded-lg border border-blue-500 text-blue-500 text-xs cursor-pointer hover:bg-blue-50"
            >
              上传
              <input
                type="file"
                accept="image/png,image/jpeg,image/webp"
                class="hidden"
                @change="handleUploadAvatar"
              />
            </label>
          </div>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" maxlength="30" />
        </el-form-item>
        <el-form-item label="个性签名">
          <el-input v-model="editForm.signature" type="textarea" :rows="2" maxlength="255" />
        </el-form-item>
        <el-form-item label="Codeforces ID">
          <el-input v-model="editForm.codeforces_id" placeholder="用于自动展示 CF 分数" />
        </el-form-item>

        <el-form-item label="获奖经历">
          <div class="w-full space-y-2">
            <div v-for="(a, i) in editForm.awards" :key="i" class="flex items-center gap-2">
              <el-input v-model="a.name" placeholder="比赛名称" class="flex-1" />
              <el-select v-model="a.level" class="!w-24">
                <el-option :value="1" label="🥇 金" />
                <el-option :value="2" label="🥈 银" />
                <el-option :value="3" label="🥉 铜" />
              </el-select>
              <button class="text-red-400 hover:text-red-600 px-1" @click="removeAward(i)">
                <i class="fa-solid fa-trash" />
              </button>
            </div>
            <button class="text-sm text-blue-500 hover:underline" @click="addAward">
              + 添加获奖经历
            </button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <button class="px-6 py-2 rounded-lg bg-black text-white text-sm" @click="saveEdit">
          保存
        </button>
      </template>
    </el-dialog>
  </div>
</template>
