// 全局 toast（模块级单例，组件用 useMessage() 取同一份状态）
import { ref } from 'vue'
import { ApiError } from '@/api/http'

export interface ToastItem {
  id: number
  content: string
  type: 'info' | 'success' | 'warning' | 'error'
}

const toasts = ref<ToastItem[]>([])
let nextId = 1

export function useMessage() {
  function remove(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function addMessage(content: string, type: ToastItem['type'] = 'info') {
    const id = nextId++
    toasts.value.push({ id, content, type })
    setTimeout(() => remove(id), 3000)
  }

  // 统一的业务码处理：已知码给友好提示，未映射的直接显示后端 message
  function codeHandler(err: unknown, successMsg?: string) {
    if (err instanceof ApiError) {
      if (err.code === 0) {
        if (successMsg) addMessage(successMsg, 'success')
        return
      }
      addMessage(err.message, 'error')
      return
    }
    if (err instanceof Error) {
      addMessage(err.message || '网络错误，请稍后重试', 'error')
      return
    }
    addMessage('未知错误', 'error')
  }

  return { toasts, addMessage, codeHandler }
}
