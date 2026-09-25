<script setup lang="ts">
// @用户选择器（编辑器工具栏用）：搜索并插入 [@用户名](/profile/id)
import { ref } from 'vue'
import type { UserBrief } from '@/api/user'
import { batchUsers } from '@/api/user'
import { GetRoleLabel } from '@/utils/level'
import { avatarUrl } from '@/utils/format'

const emit = defineEmits<{ (e: 'select', user: UserBrief): void }>()
const visible = defineModel<boolean>({ default: false })

const keyword = ref('')
const users = ref<UserBrief[]>([])
const searching = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | null = null

function onInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(doSearch, 300)
}

// 本地简单检索：后端未提供用户搜索公开接口时，这里用批量接口拉取作者头像信息由调用方维护；
// 主要场景是评论区 @ 已在页面出现的用户，直接全量展示传入列表。
async function doSearch() {
  searching.value = true
  try {
    const kw = keyword.value.trim()
    const ids = props.candidates.map((c) => c.id)
    if (ids.length === 0) {
      users.value = []
    } else {
      const res = await batchUsers(ids)
      users.value = res.users.filter((u) => !kw || u.username.includes(kw))
    }
  } finally {
    searching.value = false
  }
}

const props = defineProps<{ candidates: UserBrief[] }>()

function pick(user: UserBrief) {
  emit('select', user)
  visible.value = false
  keyword.value = ''
}
</script>

<template>
  <el-dialog v-model="visible" title="提及用户" width="420px" append-to-body>
    <el-input
      v-model="keyword"
      placeholder="输入用户名筛选"
      clearable
      @input="onInput"
      class="mb-3"
    />
    <div class="max-h-72 overflow-y-auto divide-y">
      <div v-if="users.length === 0 && !searching" class="py-8 text-center text-sm text-gray-400">
        暂无可提及的用户
      </div>
      <button
        v-for="u in users"
        :key="u.id"
        class="w-full flex items-center gap-3 px-2 py-2.5 hover:bg-gray-50 text-left"
        @click="pick(u)"
      >
        <img :src="avatarUrl(u.avatar)" class="w-9 h-9 rounded-full object-cover" alt="" />
        <div class="flex-1 min-w-0">
          <div class="text-sm font-bold truncate">{{ u.username }}</div>
          <div class="text-xs text-gray-400">{{ GetRoleLabel(u.role) }} · XP: {{ u.xp }}</div>
        </div>
      </button>
    </div>
  </el-dialog>
</template>
