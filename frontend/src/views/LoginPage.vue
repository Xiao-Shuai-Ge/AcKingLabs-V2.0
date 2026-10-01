<script setup lang="ts">
// 登录 / 注册（注册主通道为投递简历，邀请码注册是特殊通道）
import { reactive, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { onMounted } from 'vue'
import { login, register, sendCode } from '@/api/auth'
import { setTokens, clearTokens } from '@/api/http'
import { useUserStore } from '@/stores/user'
import { useMessage } from '@/composables/useMessage'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const { addMessage, codeHandler } = useMessage()

const tab = ref<'login' | 'register'>('login')
const loading = ref(false)

// 注册 tab：choose = 选择加入方式（主视图），invite = 邀请码注册表单
const registerMode = ref<'choose' | 'invite'>('choose')

const loginForm = reactive({ email: '', password: '', remember: true })
const registerForm = reactive({
  username: '',
  email: '',
  code: '',
  password: '',
  confirm: '',
  invitation_code: '',
})

const codeCountdown = ref(0)
let codeTimer: ReturnType<typeof setInterval> | null = null

const passwordMismatch = computed(
  () => registerForm.confirm !== '' && registerForm.confirm !== registerForm.password,
)

async function handleSendCode() {
  if (!registerForm.email) {
    addMessage('请先填写邮箱', 'warning')
    return
  }
  if (codeCountdown.value > 0) return
  try {
    await sendCode(registerForm.email)
    addMessage('验证码已发送，请查收邮箱', 'success')
    codeCountdown.value = 60
    codeTimer = setInterval(() => {
      codeCountdown.value--
      if (codeCountdown.value <= 0 && codeTimer) clearInterval(codeTimer)
    }, 1000)
  } catch (err) {
    codeHandler(err)
  }
}

async function handleLogin() {
  if (loading.value) return
  loading.value = true
  try {
    const res = await login({
      email: loginForm.email,
      password: loginForm.password,
      is_remember: loginForm.remember,
    })
    setTokens(res.access_token, res.refresh_token)
    // 拉取用户信息；偶发失败（网络抖动/旧会话清理竞态）时重试一次
    let ok = await userStore.refreshUser()
    if (!ok) {
      await new Promise((r) => setTimeout(r, 600))
      ok = await userStore.refreshUser()
    }
    addMessage('登录成功', 'success')
    // 无论用户信息是否拉取成功都跳转（顶栏会在路由切换时自动补拉）
    const redirect = (route.query.redirect as string) || '/'
    if (redirect !== route.fullPath) {
      await router.push(redirect).catch(() => {})
    } else {
      await router.push('/').catch(() => {})
    }
  } catch (err) {
    codeHandler(err)
  } finally {
    loading.value = false
  }
}

// 带着过期/无效令牌访问登录页时，顶栏挂载发出的旧请求会触发
// "401 -> 刷新失败 -> 清空令牌"的流程，与正在提交的登录互相干扰。
// 进入登录页即视为放弃当前会话：清掉残留令牌，避免竞态。
onMounted(() => {
  if (!userStore.currentUser) {
    clearTokens()
    userStore.setUser(null)
  }
})

async function handleRegister() {
  if (loading.value) return
  if (passwordMismatch.value) {
    addMessage('两次输入的密码不一致', 'warning')
    return
  }
  loading.value = true
  try {
    const res = await register({
      email: registerForm.email,
      code: registerForm.code,
      password: registerForm.password,
      username: registerForm.username,
      invitation_code: registerForm.invitation_code,
    })
    setTokens(res.access_token, res.refresh_token)
    await userStore.refreshUser()
    addMessage('注册成功，欢迎加入！', 'success')
    router.push('/')
  } catch (err) {
    codeHandler(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="pt-[60px] min-h-screen flex items-center justify-center px-4 py-10">
    <div class="w-full max-w-md bg-white rounded-xl border border-gray-200 shadow-sm p-8">
      <div class="flex justify-center mb-6">
        <img src="/assets/AcKing_black.png" class="h-16 w-16 rounded-2xl" alt="AcKing" />
      </div>

      <!-- Tabs -->
      <div class="flex border-b mb-6">
        <button
          class="flex-1 pb-2 text-center font-medium border-b-2 transition-colors"
          :class="tab === 'login' ? 'border-black text-black' : 'border-transparent text-gray-400'"
          @click="tab = 'login'"
        >
          登录
        </button>
        <button
          class="flex-1 pb-2 text-center font-medium border-b-2 transition-colors"
          :class="
            tab === 'register' ? 'border-black text-black' : 'border-transparent text-gray-400'
          "
          @click="tab = 'register'"
        >
          注册
        </button>
      </div>

      <!-- 登录 -->
      <form v-if="tab === 'login'" class="space-y-4" @submit.prevent="handleLogin">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
          <el-input v-model="loginForm.email" placeholder="请输入邮箱" size="large" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">密码</label>
          <el-input
            v-model="loginForm.password"
            type="password"
            show-password
            placeholder="请输入密码"
            size="large"
          />
        </div>
        <div class="flex items-center justify-between text-sm">
          <label class="flex items-center gap-2 text-gray-600 cursor-pointer">
            <input v-model="loginForm.remember" type="checkbox" class="rounded" />
            记住我
          </label>
          <router-link to="/reset-password" class="text-blue-500 hover:underline"
            >忘记密码？</router-link
          >
        </div>
        <button
          type="submit"
          class="w-full py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800 transition-colors disabled:opacity-50"
          :disabled="loading || !loginForm.email || !loginForm.password"
        >
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>

      <!-- 注册：主视图 = 选择加入方式（投递简历为主通道） -->
      <div v-else-if="registerMode === 'choose'" class="space-y-4">
        <div
          class="rounded-xl border-2 border-black p-5 cursor-pointer hover:shadow-md transition-shadow"
          @click="router.push('/more/resume')"
        >
          <div class="flex items-center gap-3 mb-2">
            <i class="fa-solid fa-file-signature text-lg" />
            <span class="font-bold">投递简历申请加入</span>
            <span
              class="ml-auto px-2 py-0.5 rounded-full bg-black text-white text-xs font-medium"
              >推荐</span
            >
          </div>
          <p class="text-sm text-gray-500 leading-relaxed">
            无需邀请码。填写简历并验证邮箱，审核通过后自动开通账号，初始密码将发送到您的邮箱。
          </p>
        </div>

        <div
          class="rounded-xl border border-gray-200 p-5 cursor-pointer hover:bg-gray-50 transition-colors"
          @click="registerMode = 'invite'"
        >
          <div class="flex items-center gap-3 mb-2">
            <i class="fa-solid fa-key text-lg text-gray-500" />
            <span class="font-medium text-gray-700">我有邀请码，直接注册</span>
            <i class="fa-solid fa-chevron-right text-gray-300 ml-auto" />
          </div>
          <p class="text-sm text-gray-400">内部成员专属通道，凭邀请码立即完成注册。</p>
        </div>
      </div>

      <!-- 注册：邀请码注册表单（特殊通道） -->
      <form v-else class="space-y-4" @submit.prevent="handleRegister">
        <button
          type="button"
          class="text-sm text-gray-400 hover:text-gray-600"
          @click="registerMode = 'choose'"
        >
          ‹ 其他加入方式
        </button>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">用户名</label>
          <el-input
            v-model="registerForm.username"
            placeholder="2~30 个字符"
            maxlength="30"
            size="large"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
          <el-input v-model="registerForm.email" placeholder="请输入邮箱" size="large" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">邮箱验证码</label>
          <div class="flex gap-2">
            <el-input
              v-model="registerForm.code"
              placeholder="6 位验证码"
              maxlength="6"
              size="large"
            />
            <button
              type="button"
              class="shrink-0 px-4 rounded-lg border text-sm"
              :class="
                codeCountdown > 0
                  ? 'text-gray-400 border-gray-200'
                  : 'text-blue-500 border-blue-500 hover:bg-blue-50'
              "
              :disabled="codeCountdown > 0"
              @click="handleSendCode"
            >
              {{ codeCountdown > 0 ? `${codeCountdown}s 后重发` : '获取验证码' }}
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">密码</label>
          <el-input
            v-model="registerForm.password"
            type="password"
            show-password
            placeholder="6~30 位"
            size="large"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">确认密码</label>
          <el-input
            v-model="registerForm.confirm"
            type="password"
            show-password
            placeholder="再次输入密码"
            size="large"
            :class="{ 'is-error': passwordMismatch }"
          />
          <p v-if="passwordMismatch" class="text-xs text-red-500 mt-1">两次输入的密码不一致</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            邀请码
            <span class="text-xs text-gray-400">（内部成员专属通道）</span>
          </label>
          <el-input
            v-model="registerForm.invitation_code"
            placeholder="请输入邀请码"
            size="large"
          />
        </div>
        <button
          type="submit"
          class="w-full py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800 transition-colors disabled:opacity-50"
          :disabled="
            loading ||
            !registerForm.email ||
            !registerForm.code ||
            !registerForm.password ||
            !registerForm.username ||
            !registerForm.invitation_code
          "
        >
          {{ loading ? '注册中...' : '注册' }}
        </button>
      </form>
    </div>
  </div>
</template>
