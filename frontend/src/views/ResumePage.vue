<script setup lang="ts">
// 投递简历（两步向导）：第一步验证邮箱 → 第二步填写简历。
// 审核通过后自动开通账号，初始密码随机生成并发送到邮箱；未通过可修改后重新投递。
import { computed, onMounted, reactive, ref } from 'vue'
import { sendCode, submitResume, updateResume, getResumeBySelf } from '@/api/resume'
import { uploadImage } from '@/api/file'
import { ApiError } from '@/api/http'
import { useMessage } from '@/composables/useMessage'

const { addMessage, codeHandler } = useMessage()

const step = ref<1 | 2>(1)
const verifying = ref(false)
const codeExpired = ref(false) // 填表时间过长导致验证码过期时，第二步内联重发

const form = reactive({
  avatar: '',
  real_name: '',
  grade: 25,
  student_no: '',
  email: '',
  username: '',
  code: '',
  extra: {
    information: '',
    skills: '',
    reason: '',
    understanding: '',
    future_plan: '',
  },
})

const resumeId = ref('') // 有已投递记录时用于"更新简历 / 重新投递"
const resumeStatus = ref<number | null>(null)
const submitting = ref(false)
const uploading = ref(false)
const countdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const statusMeta: Record<number, { label: string; cls: string; hint: string }> = {
  0: {
    label: '待审核',
    cls: 'bg-blue-100 text-blue-600',
    hint: '您的简历正在审核中，期间可继续修改；审核通过后将自动开通账号。',
  },
  1: {
    label: '已通过',
    cls: 'bg-green-100 text-green-600',
    hint: '恭喜！账号已自动开通，初始密码已发送到您的邮箱，请查收后登录。',
  },
  [-1]: {
    label: '未通过',
    cls: 'bg-red-100 text-red-600',
    hint: '很遗憾，您的简历未通过审核；可修改简历后重新投递。',
  },
}

// 已通过 = 账号已开通，资料以账号为准，表单整体锁定
const locked = computed(() => resumeStatus.value === 1)
const isRejected = computed(() => resumeStatus.value === -1)

const emailValid = computed(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim()))

const valid = computed(
  () =>
    !locked.value &&
    form.real_name.trim().length >= 2 &&
    form.student_no.trim() !== '' &&
    form.grade > 0 &&
    form.username.trim().length >= 2,
)

async function handleSendCode() {
  if (!emailValid.value) {
    addMessage('请先填写正确的邮箱', 'warning')
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

// 第一步"下一步"：验证邮箱并识别投递状态（新投递 / 待审核 / 未通过 / 已通过 / 已注册）
async function nextStep() {
  if (verifying.value) return
  if (!emailValid.value) {
    addMessage('请输入正确的邮箱', 'warning')
    return
  }
  if (form.code.trim().length !== 6) {
    addMessage('请填写 6 位邮箱验证码', 'warning')
    return
  }
  verifying.value = true
  let ok = false
  try {
    try {
      const r = await getResumeBySelf(form.email.trim(), form.code.trim())
      applyResume(r)
    } catch (err) {
      if (err instanceof ApiError && err.code === 40400) {
        startNewResume() // 该邮箱还没有投递过，进入新投递
      } else {
        if (err instanceof ApiError) addMessage(err.message, 'error')
        else addMessage((err as Error).message || '网络错误，请稍后重试', 'error')
        return
      }
    }
    ok = true
  } finally {
    verifying.value = false
    if (ok) {
      codeExpired.value = false
      step.value = 2
    }
  }
}

function startNewResume() {
  resumeId.value = ''
  resumeStatus.value = null
  form.avatar = ''
  form.real_name = ''
  form.grade = 25
  form.student_no = ''
  form.username = ''
  form.extra = {
    information: '',
    skills: '',
    reason: '',
    understanding: '',
    future_plan: '',
  }
}

function applyResume(r: {
  id: string
  avatar: string
  real_name: string
  grade: number
  student_no: string
  email: string
  username: string
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
  form.username = r.username
  form.extra = { ...r.extra }
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

async function handleSubmit() {
  if (!valid.value || submitting.value) return
  submitting.value = true
  try {
    if (resumeId.value) {
      const wasRejected = isRejected.value
      await updateResume({
        ...form,
        id: resumeId.value,
        email: form.email.trim(),
        username: form.username.trim(),
        grade: form.grade,
      })
      // 未通过 -> 重新投递后回到待审核
      resumeStatus.value = 0
      addMessage(wasRejected ? '简历已重新投递，进入新一轮审核！' : '简历更新成功！', 'success')
    } else {
      const res = await submitResume({
        ...form,
        email: form.email.trim(),
        username: form.username.trim(),
      })
      resumeId.value = res.id
      resumeStatus.value = 0
      addMessage('投递成功！审核通过后账号将自动开通，初始密码会发送到您的邮箱。', 'success')
    }
    codeExpired.value = false
  } catch (err) {
    if (err instanceof ApiError && err.code === 41001) {
      codeExpired.value = true
      addMessage('验证码已过期，请重新获取后再提交', 'warning')
    } else {
      codeHandler(err)
    }
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
      <p class="text-sm text-gray-500 mb-6">加入 AcKing 实验室，审核通过后自动为您开通平台账号</p>

      <!-- 步骤指示 -->
      <div class="flex items-center gap-3 mb-6 text-sm">
        <span
          class="flex items-center gap-1.5"
          :class="step === 1 ? 'text-black font-medium' : 'text-gray-400'"
        >
          <span
            class="w-5 h-5 rounded-full flex items-center justify-center text-xs"
            :class="step === 1 ? 'bg-black text-white' : 'bg-gray-200 text-gray-500'"
            >1</span
          >
          验证邮箱
        </span>
        <span class="flex-1 h-px bg-gray-200" />
        <span
          class="flex items-center gap-1.5"
          :class="step === 2 ? 'text-black font-medium' : 'text-gray-400'"
        >
          <span
            class="w-5 h-5 rounded-full flex items-center justify-center text-xs"
            :class="step === 2 ? 'bg-black text-white' : 'bg-gray-200 text-gray-500'"
            >2</span
          >
          填写简历
        </span>
      </div>

      <!-- 第一步：验证邮箱 -->
      <div v-if="step === 1" class="bg-white rounded-lg border border-gray-200 p-6 space-y-5">
        <p class="text-sm text-gray-500 leading-relaxed">
          使用常用邮箱验证身份；审核结果与账号初始密码都将发送到该邮箱。
        </p>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1"
            >邮箱 <span class="text-red-500">*</span></label
          >
          <el-input
            v-model="form.email"
            placeholder="请输入邮箱"
            size="large"
            @blur="persistEmail"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1"
            >邮箱验证码 <span class="text-red-500">*</span></label
          >
          <div class="flex gap-2">
            <el-input
              v-model="form.code"
              placeholder="6 位验证码"
              maxlength="6"
              size="large"
              @keyup.enter="nextStep"
            />
            <button
              type="button"
              class="shrink-0 px-4 rounded-lg border text-sm"
              :class="
                countdown > 0
                  ? 'text-gray-400 border-gray-200'
                  : 'text-blue-500 border-blue-500 hover:bg-blue-50'
              "
              :disabled="countdown > 0"
              @click="handleSendCode"
            >
              {{ countdown > 0 ? `${countdown}s 后重发` : '获取验证码' }}
            </button>
          </div>
        </div>
        <button
          type="button"
          class="w-full py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800 transition-colors disabled:opacity-50"
          :disabled="!emailValid || form.code.trim().length !== 6 || verifying"
          @click="nextStep"
        >
          {{ verifying ? '验证中...' : '下一步，填写简历' }}
        </button>
        <p class="text-xs text-gray-400">
          已经投递过？同样输入邮箱验证即可查看进度、继续修改或重新投递。
        </p>
      </div>

      <!-- 第二步：填写简历 -->
      <template v-else>
        <!-- 已验证邮箱条 -->
        <div class="bg-white rounded-lg border border-gray-200 p-4 mb-4 flex items-center gap-3">
          <i class="fa-solid fa-circle-check text-green-500" />
          <span class="text-sm text-gray-600 flex-1 min-w-0 truncate">
            已验证邮箱：{{ form.email }}
          </span>
          <button
            type="button"
            class="text-sm text-gray-400 hover:text-gray-600 shrink-0"
            @click="step = 1"
          >
            更换邮箱
          </button>
        </div>

        <!-- 状态卡片 -->
        <div
          v-if="resumeStatus !== null"
          class="bg-white rounded-lg border border-gray-200 p-5 mb-4 flex items-center gap-4 flex-wrap"
        >
          <div class="text-sm text-gray-500">当前状态：</div>
          <span
            class="px-3 py-1 rounded-full text-sm font-medium"
            :class="statusMeta[resumeStatus]?.cls"
          >
            {{ statusMeta[resumeStatus]?.label ?? '未知' }}
          </span>
          <span class="text-sm text-gray-500 flex-1">{{ statusMeta[resumeStatus]?.hint }}</span>
          <router-link
            v-if="locked"
            to="/login"
            class="px-4 py-1.5 rounded-lg bg-black text-white text-sm"
          >
            去登录
          </router-link>
        </div>

        <!-- 验证码过期：内联重新获取 -->
        <div
          v-if="codeExpired"
          class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4 flex flex-col md:flex-row md:items-center gap-3"
        >
          <p class="text-sm text-yellow-700 flex-1">验证码已过期，请重新获取后再提交：</p>
          <div class="flex gap-2">
            <el-input v-model="form.code" placeholder="6 位验证码" maxlength="6" class="md:!w-36" />
            <button
              type="button"
              class="shrink-0 px-4 rounded-lg border text-sm"
              :class="
                countdown > 0
                  ? 'text-gray-400 border-gray-200'
                  : 'text-yellow-600 border-yellow-400 hover:bg-yellow-100'
              "
              :disabled="countdown > 0"
              @click="handleSendCode"
            >
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </button>
          </div>
        </div>

        <fieldset
          class="bg-white rounded-lg border border-gray-200 p-6 space-y-5 min-w-0"
          :disabled="locked"
        >
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
              <el-input
                v-model="form.avatar"
                placeholder="照片链接，或点击右侧上传"
                class="flex-1"
              />
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

          <!-- 账号信息 -->
          <div class="border-t border-gray-100 pt-5">
            <div class="text-sm font-bold text-gray-700 mb-3">
              账号信息
              <span class="text-xs text-gray-400 font-normal">
                审核通过后自动开通账号；初始密码由系统随机生成并发送到邮箱
              </span>
            </div>
            <label class="block text-sm font-medium text-gray-700 mb-1"
              >用户名 <span class="text-red-500">*</span></label
            >
            <el-input
              v-model="form.username"
              maxlength="30"
              placeholder="2~30 个字符，登录后展示用"
              class="md:!w-80"
            />
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
            <el-input
              v-model="form.extra.understanding"
              type="textarea"
              :rows="3"
              maxlength="2000"
            />
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
              {{
                submitting
                  ? '提交中...'
                  : locked
                    ? '已开通账号'
                    : resumeId !== ''
                      ? isRejected
                        ? '重新投递'
                        : '更新简历'
                      : '投递简历'
              }}
            </button>
            <button
              type="button"
              class="px-6 py-2.5 rounded-lg border-2 border-gray-300 text-gray-600 font-medium hover:bg-gray-100"
              @click="step = 1"
            >
              上一步
            </button>
          </div>
          <p class="text-xs text-gray-400">
            待审核期间可返回修改；未通过可修改后重新投递；审核通过后账号自动开通。
          </p>
        </fieldset>
      </template>
    </div>
  </div>
</template>
