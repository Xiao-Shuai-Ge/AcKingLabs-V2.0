<script setup lang="ts">
// 个人设置：通知偏好（存 users.settings JSON 列的 notify 分组，
// 站内消息按偏好过滤生成；"系统消息邮件"开关控制异步邮件同步）
import { computed, onMounted, reactive, ref } from 'vue'
import { getUserSetting, updateUserSetting, changePassword } from '@/api/user'
import type { NotifySettings } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { useMessage } from '@/composables/useMessage'

const { addMessage, codeHandler } = useMessage()
const userStore = useUserStore()

const defaults: NotifySettings = {
  like: true,
  comment: true,
  mention: true,
  help_post: true,
  system_email: false,
  new_resume_email: false,
}

const form = reactive<NotifySettings>({ ...defaults })
const saved = ref<NotifySettings>({ ...defaults })
const loading = ref(true)
const saving = ref(false)

const dirty = computed(() =>
  (Object.keys(defaults) as (keyof NotifySettings)[]).some((k) => form[k] !== saved.value[k]),
)

const baseGroups: {
  key: keyof NotifySettings
  icon: string
  color: string
  title: string
  desc: string
}[] = [
  {
    key: 'like',
    icon: 'fa-heart',
    color: 'text-red-500 bg-red-50',
    title: '点赞通知',
    desc: '别人赞了你的帖子或评论时提醒',
  },
  {
    key: 'comment',
    icon: 'fa-comment-dots',
    color: 'text-blue-500 bg-blue-50',
    title: '评论 / 回复通知',
    desc: '别人评论你的帖子、回复你的评论时提醒',
  },
  {
    key: 'mention',
    icon: 'fa-at',
    color: 'text-green-500 bg-green-50',
    title: '提及通知',
    desc: '别人在帖子或评论中 @你 时提醒',
  },
  {
    key: 'help_post',
    icon: 'fa-circle-question',
    color: 'text-yellow-500 bg-yellow-50',
    title: '求助帖提醒',
    desc: '有人发布新的求助帖时提醒，帮小伙伴们解答',
  },
  {
    key: 'system_email',
    icon: 'fa-envelope',
    color: 'text-purple-500 bg-purple-50',
    title: '系统消息邮件同步',
    desc: '精选、下架等系统消息同时发送到你的邮箱',
  },
]

// 新简历邮件提醒：仅管理员可见（后端也会拒绝普通用户开启）
const groups = computed(() =>
  userStore.isAdmin
    ? [
        ...baseGroups,
        {
          key: 'new_resume_email' as keyof NotifySettings,
          icon: 'fa-file-lines',
          color: 'text-orange-500 bg-orange-50',
          title: '新简历邮件提醒',
          desc: '收到新的简历投递时发送邮件提醒你前往审核（管理员专属）',
        },
      ]
    : baseGroups,
)

async function load() {
  try {
    const s = await getUserSetting()
    Object.assign(form, defaults, s.notify)
    saved.value = { ...form }
  } catch (err) {
    codeHandler(err)
  } finally {
    loading.value = false
  }
}

function resetDefaults() {
  Object.assign(form, defaults)
}

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    await updateUserSetting({ notify: { ...form } })
    saved.value = { ...form }
    addMessage('设置已保存', 'success')
  } catch (err) {
    codeHandler(err)
  } finally {
    saving.value = false
  }
}

onMounted(load)

// ---- 修改密码 ----
const pwdForm = reactive({ old_password: '', new_password: '', confirm: '' })
const changingPwd = ref(false)

const pwdMismatch = computed(
  () => pwdForm.confirm !== '' && pwdForm.confirm !== pwdForm.new_password,
)
const pwdValid = computed(
  () =>
    pwdForm.old_password !== '' &&
    pwdForm.new_password.length >= 6 &&
    pwdForm.new_password === pwdForm.confirm,
)

async function savePassword() {
  if (!pwdValid.value || changingPwd.value) return
  changingPwd.value = true
  try {
    await changePassword({
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password,
    })
    addMessage('密码修改成功', 'success')
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm = ''
  } catch (err) {
    codeHandler(err)
  } finally {
    changingPwd.value = false
  }
}
</script>

<template>
  <div class="pt-[60px] min-h-screen bg-gray-50">
    <div class="max-w-2xl mx-auto px-4 py-10" v-loading="loading">
      <h1 class="text-2xl font-bold">个人设置</h1>
      <p class="text-sm text-gray-500 mt-1 mb-8">管理您的通知偏好设置</p>

      <div class="bg-white rounded-xl border border-gray-200 divide-y">
        <div v-for="g in groups" :key="g.key" class="flex items-center gap-4 p-5">
          <div
            class="w-11 h-11 rounded-xl flex items-center justify-center shrink-0"
            :class="g.color"
          >
            <i class="fa-solid text-lg" :class="g.icon" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="font-medium">{{ g.title }}</div>
            <div class="text-sm text-gray-400 mt-0.5">{{ g.desc }}</div>
          </div>
          <el-switch v-model="form[g.key]" />
        </div>
      </div>

      <div class="flex items-center gap-3 mt-6">
        <button
          class="px-6 py-2.5 rounded-lg bg-black text-white text-sm font-medium hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed"
          :disabled="!dirty || saving"
          @click="save"
        >
          {{ saving ? '保存中...' : '保存设置' }}
        </button>
        <button
          class="px-5 py-2.5 rounded-lg border-2 border-gray-300 text-gray-600 text-sm font-medium hover:bg-gray-100 disabled:opacity-40"
          :disabled="!dirty"
          @click="Object.assign(form, saved)"
        >
          取消
        </button>
        <button class="ml-auto text-sm text-gray-400 hover:text-gray-600" @click="resetDefaults">
          恢复默认设置
        </button>
      </div>

      <p class="text-xs text-gray-400 mt-6 flex items-center gap-1.5">
        <i class="fa-solid fa-circle-info" />
        关闭通知后，您仍然可以在消息中心查看所有历史通知
      </p>

      <!-- 修改密码 -->
      <div class="bg-white rounded-xl border border-gray-200 p-6 mt-8">
        <h2 class="font-bold mb-1">修改密码</h2>
        <p class="text-sm text-gray-400 mb-5">
          简历审核通过后发放的是随机初始密码，建议登录后尽快修改为自己的密码。
        </p>
        <div class="space-y-4 md:w-96">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">当前密码</label>
            <el-input
              v-model="pwdForm.old_password"
              type="password"
              show-password
              placeholder="初始密码或当前使用的密码"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">新密码</label>
            <el-input
              v-model="pwdForm.new_password"
              type="password"
              show-password
              placeholder="6~30 位"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">确认新密码</label>
            <el-input
              v-model="pwdForm.confirm"
              type="password"
              show-password
              placeholder="再次输入新密码"
            />
            <p v-if="pwdMismatch" class="text-xs text-red-500 mt-1">两次输入的新密码不一致</p>
          </div>
          <button
            class="px-6 py-2.5 rounded-lg bg-black text-white text-sm font-medium hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed"
            :disabled="!pwdValid || changingPwd"
            @click="savePassword"
          >
            {{ changingPwd ? '提交中...' : '修改密码' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
