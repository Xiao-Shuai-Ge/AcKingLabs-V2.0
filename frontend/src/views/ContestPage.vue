<script setup lang="ts">
// 比赛页：平台筛选 + 状态标记 + 预约（批量查询预约态） + 管理员编辑/精选/删除
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import {
  getContestList,
  getBookingMap,
  toggleBooking,
  adminUpdateContest,
  adminDeleteContest,
  adminSetRecommend,
} from '@/api/contest'
import type { ContestItem } from '@/api/types'
import { formatDateTime, formatDuration, contestStatus } from '@/utils/format'
import { useMessage } from '@/composables/useMessage'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { addMessage, codeHandler } = useMessage()

const platforms = [
  { key: 'all', label: '全部平台' },
  { key: 'Codeforces', label: 'Codeforces' },
  { key: 'AtCoder', label: 'AtCoder' },
  { key: 'Nowcoder', label: '牛客网' },
  { key: 'AcKing', label: '其他' },
  { key: 'recommend', label: '精选比赛' },
]

const platformLogos: Record<string, string> = {
  Codeforces: '/assets/codeforces.png',
  AtCoder: '/assets/atcoder.png',
  Nowcoder: '/assets/nowcoder.png',
  AcKing: '/assets/acking-contest.png',
}

const platform = ref((route.query.platform as string) || 'all')
const page = ref(Number(route.query.page) || 1)
const count = 5

const contests = ref<ContestItem[]>([])
const total = ref(0)
const loading = ref(false)
const bookedMap = ref<Record<string, boolean>>({})

const statusMeta: Record<string, { label: string; cls: string }> = {
  upcoming: { label: '未开始', cls: 'bg-blue-100 text-blue-600' },
  ongoing: { label: '进行中', cls: 'bg-green-100 text-green-600' },
  ended: { label: '已结束', cls: 'bg-gray-200 text-gray-500' },
}

function status(c: ContestItem) {
  return statusMeta[contestStatus(c.start_time, c.end_time)]
}

function canBook(c: ContestItem) {
  return Date.now() < c.start_time - 20 * 60 * 1000
}

const booking = ref<string>('')

async function load() {
  loading.value = true
  try {
    const res = await getContestList({ platform: platform.value, page: page.value, count })
    contests.value = res.contests
    total.value = res.total
    // 批量拉取预约状态（一次请求）
    if (userStore.isLogin && contests.value.length > 0) {
      try {
        const b = await getBookingMap(contests.value.map((c) => c.id))
        bookedMap.value = b.booked
      } catch {
        /* 未登录/失败忽略 */
      }
    }
  } catch {
    addMessage('加载比赛失败', 'error')
  } finally {
    loading.value = false
  }
}

function pickPlatform(p: string) {
  platform.value = p
  page.value = 1
  router.replace({
    query: { ...route.query, platform: p !== 'all' ? p : undefined, page: undefined },
  })
  load()
}

function changePage(p: number) {
  page.value = p
  router.replace({
    query: {
      ...route.query,
      platform: platform.value !== 'all' ? platform.value : undefined,
      page: p > 1 ? String(p) : undefined,
    },
  })
  load()
}

async function book(c: ContestItem) {
  if (!userStore.isLogin) {
    router.push('/login')
    return
  }
  if (booking.value) return
  booking.value = c.id
  try {
    const res = await toggleBooking(c.id)
    bookedMap.value[c.id] = res.booked
    addMessage(res.booked ? '预约成功，开赛前 20 分钟将邮件提醒' : '已取消预约', 'success')
  } catch (err) {
    codeHandler(err)
  } finally {
    booking.value = ''
  }
}

// ---- 管理员 ----
const editVisible = ref(false)
const editForm = reactive({ id: '', title: '', url: '', range: [] as string[] })

function openEdit(c: ContestItem) {
  editForm.id = c.id
  editForm.title = c.title
  editForm.url = c.url
  editForm.range = [String(c.start_time), String(c.end_time)]
  editVisible.value = true
}

async function saveEdit() {
  if (!editForm.title || !editForm.url || editForm.range.length !== 2) {
    addMessage('请填写完整', 'warning')
    return
  }
  try {
    await adminUpdateContest({
      id: editForm.id,
      title: editForm.title,
      url: editForm.url,
      start_time: Number(editForm.range[0]),
      end_time: Number(editForm.range[1]),
    })
    addMessage('修改成功', 'success')
    editVisible.value = false
    load()
  } catch (err) {
    codeHandler(err)
  }
}

async function toggleRecommend(c: ContestItem) {
  try {
    await adminSetRecommend(c.id, !c.is_recommend)
    c.is_recommend = !c.is_recommend
    addMessage(c.is_recommend ? '已设为精选比赛' : '已取消精选', 'success')
  } catch (err) {
    codeHandler(err)
  }
}

async function removeContest(c: ContestItem) {
  try {
    await ElMessageBox.confirm(`确定删除比赛「${c.title}」吗？`, '删除比赛', { type: 'warning' })
  } catch {
    return
  }
  try {
    await adminDeleteContest(c.id)
    addMessage('已删除', 'success')
    load()
  } catch (err) {
    codeHandler(err)
  }
}

onMounted(load)
</script>

<template>
  <div class="pt-[60px]">
    <div class="max-w-4xl mx-auto px-4 py-8">
      <!-- 平台筛选 -->
      <div class="flex items-center flex-wrap gap-2 mb-6">
        <button
          v-for="p in platforms"
          :key="p.key"
          class="px-4 py-1.5 rounded-full text-sm border transition-colors"
          :class="
            platform === p.key
              ? 'bg-black text-white border-black'
              : 'border-gray-300 text-gray-600 hover:border-gray-500'
          "
          @click="pickPlatform(p.key)"
        >
          {{ p.label }}
        </button>
        <router-link
          v-if="userStore.isAdmin"
          to="/contest/create"
          class="ml-auto px-4 py-1.5 rounded-full bg-blue-500 text-white text-sm hover:bg-blue-600"
        >
          + 创建比赛
        </router-link>
      </div>

      <!-- 比赛列表 -->
      <div class="space-y-4" v-loading="loading">
        <div
          v-for="c in contests"
          :key="c.id"
          class="bg-white rounded-lg border border-gray-200 p-4 md:p-5 flex flex-col md:flex-row gap-4 hover:shadow-lg transition-shadow fade-in-up"
        >
          <img
            :src="platformLogos[c.platform] ?? platformLogos.AcKing"
            class="w-12 h-12 rounded-lg object-contain shrink-0"
            :alt="c.platform"
          />
          <div class="flex-1 min-w-0 cursor-pointer" @click="router.push(`/contest/${c.id}`)">
            <div class="flex items-center gap-2 flex-wrap">
              <span
                class="font-bold text-base md:text-lg truncate"
                :class="c.is_recommend ? 'text-yellow-500' : ''"
              >
                {{ c.title }}
              </span>
              <span class="shrink-0 px-2 py-0.5 rounded text-xs" :class="status(c)?.cls">{{
                status(c)?.label
              }}</span>
              <span
                v-if="c.is_recommend"
                class="shrink-0 px-2 py-0.5 rounded text-xs bg-yellow-100 text-yellow-700"
                >精选</span
              >
            </div>
            <div class="text-sm text-gray-500 mt-1.5">
              <i class="fa-regular fa-clock mr-1" />
              {{ formatDateTime(c.start_time) }} 至 {{ formatDateTime(c.end_time) }}
              <span class="text-gray-400">（{{ formatDuration(c.duration) }}）</span>
            </div>
            <div class="text-xs text-gray-400 mt-1">{{ c.platform }}</div>
          </div>

          <div class="flex items-center gap-2 shrink-0">
            <template v-if="canBook(c)">
              <button
                class="px-4 py-1.5 rounded-lg text-sm font-medium border transition-colors"
                :class="
                  bookedMap[c.id]
                    ? 'border-blue-500 text-blue-500 bg-blue-50'
                    : 'bg-blue-500 text-white border-blue-500 hover:bg-blue-600'
                "
                :disabled="booking === c.id"
                @click="book(c)"
              >
                {{ bookedMap[c.id] ? '已预约' : '预约' }}
              </button>
            </template>
            <button
              class="px-4 py-1.5 rounded-lg text-sm border border-gray-300 text-gray-600 hover:bg-gray-50"
              @click="router.push(`/contest/${c.id}`)"
            >
              题解
            </button>

            <template v-if="userStore.isAdmin">
              <el-tooltip content="精选">
                <button
                  class="w-8 h-8 rounded-full border border-gray-300 text-gray-500 hover:text-yellow-500 hover:border-yellow-500 flex items-center justify-center"
                  @click="toggleRecommend(c)"
                >
                  <i class="fa-solid fa-star text-xs" />
                </button>
              </el-tooltip>
              <el-tooltip content="编辑">
                <button
                  class="w-8 h-8 rounded-full border border-gray-300 text-gray-500 hover:text-blue-500 hover:border-blue-500 flex items-center justify-center"
                  @click="openEdit(c)"
                >
                  <i class="fa-solid fa-pen text-xs" />
                </button>
              </el-tooltip>
              <el-tooltip content="删除">
                <button
                  class="w-8 h-8 rounded-full border border-gray-300 text-gray-500 hover:text-red-500 hover:border-red-500 flex items-center justify-center"
                  @click="removeContest(c)"
                >
                  <i class="fa-solid fa-trash text-xs" />
                </button>
              </el-tooltip>
            </template>
          </div>
        </div>

        <div
          v-if="contests.length === 0 && !loading"
          class="py-16 text-center text-gray-400 bg-white rounded-lg border border-gray-200"
        >
          暂无比赛数据
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="total > count" class="mt-6 flex justify-center">
        <el-pagination
          layout="prev, pager, next"
          :total="total"
          :page-size="count"
          :current-page="page"
          @current-change="changePage"
        />
      </div>
    </div>

    <!-- 管理员编辑对话框（datetimerange 直接绑定毫秒值，避免时区换算） -->
    <el-dialog v-model="editVisible" title="比赛设置" width="520px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="比赛标题">
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item label="比赛时间">
          <el-date-picker
            v-model="editForm.range"
            type="datetimerange"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="x"
            class="w-full"
          />
        </el-form-item>
        <el-form-item label="比赛链接">
          <el-input v-model="editForm.url" />
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
