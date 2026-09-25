<script setup lang="ts">
// 投递简历：邮箱验证码 + 表单 + 已投递查询（状态文案与后端统一语义）
import { computed, onMounted, reactive, ref } from 'vue'
import { sendCode, submitResume, updateResume, getResumeBySelf } from '@/api/resume'
import { uploadImage } from '@/api/file'
import { useMessage } from '@/composables/useMessage'

const { addMessage, codeHandler } = useMessage()

const form = reactive({
  avatar: '',
  real_name: '',
  grade: 25,
  student_no: '',
  email: '',
  code: '',
  extra: {
    information: '',
    skills: '',
    reason: '',
    understanding: '',
    future_plan: '',
  },
})

const resumeId = ref('') // 有已投递记录时用于"更新简历"
const resumeStatus = ref<number | null>(null)
const submitting = ref(false)
const uploading = ref(false)
const countdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const statusMeta: Record<number, { label: string; cls: string; hint: string }> = {
  0: {
    label: '审核中',
    cls: 'bg-blue-100 text-blue-600',
    hint: '您的简历正在审核中，请耐心等待。',
  },
  1: {
    label: '待考核',
    cls: 'bg-yellow-100 text-yellow-700',
    hint: '您的简历已通过初审，请按邮件指引加入考核群，等待考核通知。',
  },
  2: {
    label: '已通过',
    cls: 'bg-green-100 text-green-600',
    hint: '恭喜！您的简历已通过审核，注册邀请码已发送至您的邮箱。',
  },
  [-1]: {
    label: '未通过',
    cls: 'bg-red-100 text-red-600',
    hint: '很抱歉，您的简历未通过审核。欢迎继续关注我们的其他活动！',
  },
}

const valid = computed(
  () =>
    form.email.trim() !== '' &&
    form.code.trim() !== '' &&
    form.real_name.trim().length >= 2 &&
    form.student_no.trim() !== '' &&
    form.grade > 0,
)

async function handleSendCode() {
  if (!form.email.trim()) {
    addMessage('请先填写邮箱', 'warning')
    return
  }
  if (countdown.value > 0) return
  try {
    await sendCode(form.email.trim())
    addMessage('验证码已发送，请查收邮箱', 'success')
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0 && timer) clearInterval(timer)
    }, 1000)
  } catch (err) {
    codeHandler(err)
  }
}

async function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  if (file.size > 5 * 1024 * 1024) {
    addMessage('图片大小不能超过 5MB', 'warning')
    return
  }
  uploading.value = true
  try {
    form.avatar = await uploadImage(file)
    addMessage('照片上传成功', 'success')
  } catch (err) {
    addMessage((err as Error).message || '上传失败', 'error')
  } finally {
    uploading.value = false
    input.value = ''
  }
}

async function queryResume() {
  if (!form.email.trim() || !form.code.trim()) {
    addMessage('查询需要填写邮箱与验证码', 'warning')
    return
  }
  try {
    const r = await getResumeBySelf(form.email.trim(), form.code.trim())
    applyResume(r)
    addMessage('已加载您的简历信息', 'success')
  } catch (err) {
    codeHandler(err)
  }
}

function applyResume(r: {
  id: string
  avatar: string
  real_name: string
  grade: number
  student_no: string
  email: string
  extra: typeof form.extra
  status: number
}) {
  resumeId.value = r.id
  resumeStatus.value = r.status
  form.avatar = r.avatar
  form.real_name = r.real_name
  form.grade = r.grade
  form.student_no = r.student_no
  form.email = r.email
  form.extra = { ...r.extra }
}

async function handleSubmit() {
  if (!valid.value || submitting.value) return
  submitting.value = true
  try {
    if (resumeId.value) {
      await updateResume({
        ...form,
        id: resumeId.value,
        email: form.email.trim(),
        grade: form.grade,
      })
      addMessage('简历更新成功！', 'success')
    } else {
      await submitResume({ ...form, email: form.email.trim() })
      addMessage('简历投递成功！', 'success')
    }
  } catch (err) {
    codeHandler(err)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  const saved = localStorage.getItem('resume_email')
  if (saved) form.email = saved
})

function persistEmail() {
  localStorage.setItem('resume_email', form.email.trim())
}
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-2xl mx-auto px-4 py-10">
      <h1 class="text-2xl font-bold mb-2">投递简历</h1>
      <p class="text-sm text-gray-500 mb-8">加入 AcKing 实验室，开启您的竞赛之路</p>

      <!-- 状态卡片 -->
      <div
        v-if="resumeStatus !== null"
        class="bg-white rounded-lg border border-gray-200 p-5 mb-6 flex items-center gap-4"
      >
        <div class="text-sm text-gray-500">当前状态：</div>
        <span
          class="px-3 py-1 rounded-full text-sm font-medium"
          :class="statusMeta[resumeStatus]?.cls"
        >
          {{ statusMeta[resumeStatus]?.label ?? '未知' }}
        </span>
        <span class="text-sm text-gray-500 flex-1">{{ statusMeta[resumeStatus]?.hint }}</span>
      </div>

      <div class="bg-white rounded-lg border border-gray-200 p-6 space-y-5">
        <!-- 邮箱 + 验证码 -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1"
              >邮箱 <span class="text-red-500">*</span></label
            >
            <el-input
              v-model="form.email"
              placeholder="用于接收审核结果与邀请码"
              @blur="persistEmail"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1"
              >邮箱验证码 <span class="text-red-500">*</span></label
            >
            <div class="flex gap-2">
              <el-input v-model="form.code" placeholder="6 位验证码" maxlength="6" />
              <button
                class="shrink-0 px-4 rounded-lg border text-sm"
                :class="
                  countdown > 0
                    ? 'text-gray-400 border-gray-200'
                    : 'text-blue-500 border-blue-500 hover:bg-blue-50'
                "
                :disabled="countdown > 0"
                @click="handleSendCode"
              >
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </button>
            </div>
          </div>
        </div>

        <!-- 照片 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">本人照片（选填）</label>
          <div class="flex items-center gap-3">
            <img
              v-if="form.avatar"
              :src="form.avatar"
              class="w-14 h-14 rounded-lg object-cover border"
              alt="照片"
            />
            <el-input v-model="form.avatar" placeholder="照片链接，或点击右侧上传" class="flex-1" />
            <label
              class="shrink-0 px-4 py-2 rounded-lg border border-blue-500 text-blue-500 text-sm cursor-pointer hover:bg-blue-50"
            >
              {{ uploading ? '上传中...' : '上传图片' }}
              <input
                type="file"
                accept="image/png,image/jpeg,image/webp"
                class="hidden"
                @change="handleUpload"
              />
            </label>
          </div>
        </div>

        <!-- 基础信息 -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1"
              >真实姓名 <span class="text-red-500">*</span></label
            >
            <el-input v-model="form.real_name" maxlength="20" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1"
              >年级 <span class="text-red-500">*</span></label
            >
            <el-input-number v-model="form.grade" :min="20" :max="30" class="!w-full" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1"
              >学号 <span class="text-red-500">*</span></label
            >
            <el-input v-model="form.student_no" maxlength="20" />
          </div>
        </div>

        <!-- 补充信息 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">个人介绍</label>
          <el-input
            v-model="form.extra.information"
            type="textarea"
            :rows="3"
            placeholder="介绍一下你自己"
            maxlength="2000"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">专业能力</label>
          <el-input
            v-model="form.extra.skills"
            type="textarea"
            :rows="3"
            placeholder="编程语言、算法基础、获奖情况等"
            maxlength="2000"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">为什么要加入实验室</label>
          <el-input v-model="form.extra.reason" type="textarea" :rows="3" maxlength="2000" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">对竞赛的理解</label>
          <el-input v-model="form.extra.understanding" type="textarea" :rows="3" maxlength="2000" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">未来计划</label>
          <el-input v-model="form.extra.future_plan" type="textarea" :rows="3" maxlength="2000" />
        </div>

        <div class="flex gap-3 pt-2">
          <button
            class="px-8 py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed"
            :disabled="!valid || submitting"
            @click="handleSubmit"
          >
            {{ submitting ? '提交中...' : resumeId ? '更新简历' : '投递简历' }}
          </button>
          <button
            class="px-6 py-2.5 rounded-lg border-2 border-gray-300 text-gray-600 font-medium hover:bg-gray-100"
            @click="queryResume"
          >
            查询已投递简历
          </button>
        </div>
        <p class="text-xs text-gray-400">
          查询或更新均需先获取邮箱验证码；简历进入审核流程后不能再修改。
        </p>
      </div>
    </div>
  </div>
</template>
