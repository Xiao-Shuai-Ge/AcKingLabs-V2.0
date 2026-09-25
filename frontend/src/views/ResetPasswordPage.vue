<script setup lang="ts">
// 找回密码
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { sendCode, resetPassword } from '@/api/auth'
import { useMessage } from '@/composables/useMessage'

const router = useRouter()
const { addMessage, codeHandler } = useMessage()

const form = reactive({ email: '', code: '', password: '', confirm: '' })
const loading = ref(false)
const countdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const mismatch = computed(() => form.confirm !== '' && form.confirm !== form.password)

async function handleSendCode() {
  if (!form.email) {
    addMessage('请先填写邮箱', 'warning')
    return
  }
  try {
    await sendCode(form.email)
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

async function handleSubmit() {
  if (loading.value) return
  if (mismatch.value) {
    addMessage('两次输入的密码不一致', 'warning')
    return
  }
  loading.value = true
  try {
    await resetPassword({ email: form.email, code: form.code, password: form.password })
    addMessage('密码重置成功，请使用新密码登录', 'success')
    setTimeout(() => router.push('/login'), 1500)
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
      <h2 class="text-xl font-bold mb-6 text-center">找回密码</h2>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
          <el-input v-model="form.email" placeholder="注册时使用的邮箱" size="large" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">邮箱验证码</label>
          <div class="flex gap-2">
            <el-input v-model="form.code" placeholder="6 位验证码" maxlength="6" size="large" />
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
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">新密码</label>
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="6~30 位"
            size="large"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">确认新密码</label>
          <el-input
            v-model="form.confirm"
            type="password"
            show-password
            placeholder="再次输入新密码"
            size="large"
          />
          <p v-if="mismatch" class="text-xs text-red-500 mt-1">两次输入的密码不一致</p>
        </div>
        <button
          type="submit"
          class="w-full py-2.5 rounded-lg bg-black text-white font-medium hover:bg-gray-800 transition-colors disabled:opacity-50"
          :disabled="loading || !form.email || !form.code || !form.password"
        >
          {{ loading ? '提交中...' : '重置密码' }}
        </button>
        <div class="text-center">
          <router-link to="/login" class="text-sm text-blue-500 hover:underline"
            >返回登录</router-link
          >
        </div>
      </form>
    </div>
  </div>
</template>
