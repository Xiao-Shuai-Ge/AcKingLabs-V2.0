<script setup lang="ts">
import { useMessage } from '@/composables/useMessage'

const { toasts } = useMessage()

const colors: Record<string, string> = {
  info: 'bg-blue-500',
  success: 'bg-green-500',
  warning: 'bg-yellow-500',
  error: 'bg-red-500',
}
</script>

<template>
  <div class="fixed bottom-6 right-6 z-[999] flex flex-col gap-2">
    <transition-group name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="px-4 py-2.5 rounded-lg shadow-lg text-white text-sm max-w-xs"
        :class="colors[t.type]"
      >
        {{ t.content }}
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(16px);
}
</style>
