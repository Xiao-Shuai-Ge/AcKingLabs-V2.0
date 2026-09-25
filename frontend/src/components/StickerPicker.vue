<script setup lang="ts">
// 表情包选择器（评论区编辑器工具栏用）
import { ref } from 'vue'
import { stickerCategories } from '@/utils/stickers'

const emit = defineEmits<{ (e: 'select', url: string): void }>()
const visible = defineModel<boolean>({ default: false })
const active = ref(0)

function pick(url: string) {
  emit('select', url)
  visible.value = false
}
</script>

<template>
  <el-dialog v-model="visible" title="选择表情包" width="480px" append-to-body>
    <div class="flex gap-2 mb-3">
      <button
        v-for="(cat, i) in stickerCategories"
        :key="cat.name"
        class="px-3 py-1 rounded-full text-sm"
        :class="active === i ? 'bg-black text-white' : 'bg-gray-100 text-gray-600'"
        @click="active = i"
      >
        {{ cat.name }}
      </button>
    </div>
    <div class="grid grid-cols-4 gap-2 max-h-80 overflow-y-auto">
      <button
        v-for="url in stickerCategories[active]?.list ?? []"
        :key="url"
        class="aspect-square rounded-lg overflow-hidden border hover:border-black"
        @click="pick(url)"
      >
        <img :src="url" class="w-full h-full object-cover" alt="sticker" loading="lazy" />
      </button>
    </div>
  </el-dialog>
</template>
