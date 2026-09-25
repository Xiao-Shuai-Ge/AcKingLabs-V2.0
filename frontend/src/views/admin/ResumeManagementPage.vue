<script setup lang="ts">
// 简历管理：列表筛选、详情、待考核/通过(邀请码)/拒绝/删除；状态语义与投递页一致
import { onMounted, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import {
  adminResumeList,
  adminResumeDetail,
  adminResumeStatus,
  adminDeleteResume,
} from '@/api/resume'
import type { ResumeItem } from '@/api/types'
import AdminSidebar from '@/components/AdminSidebar.vue'
import { avatarUrl, formatDateTime } from '@/utils/format'
import { useMessage } from '@/composables/useMessage'

const { addMessage, codeHandler } = useMessage()

const collapsed = ref(false)
const keyword = ref('')
const statusFilter = ref<number | ''>('')
const page = ref(1)
const count = 20
const total = ref(0)
const resumes = ref<ResumeItem[]>([])
const loading = ref(false)

const statusOptions = [
  { value: 0, label: '待处理' },
  { value: 1, label: '待考核' },
  { value: 2, label: '已通过' },
  { value: -1, label: '未通过' },
]

const statusTag: Record<
  number,
  { label: string; type: 'primary' | 'warning' | 'success' | 'danger' }
> = {
  0: { label: '待处理', type: 'primary' },
  1: { label: '待考核', type: 'warning' },
  2: { label: '已通过', type: 'success' },
  [-1]: { label: '未通过', type: 'danger' },
}

async function load() {
  loading.value = true
  try {
    const res = await adminResumeList({
      keyword: keyword.value.trim() || undefined,
      status: statusFilter.value === '' ? undefined : statusFilter.value,
      page: page.value,
      count,
    })
    resumes.value = res.resumes
    total.value = res.total
  } catch (err) {
    codeHandler(err)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

async function openDetail(r: ResumeItem) {
  try {
    const detail = await adminResumeDetail(r.id)
    current.value = detail
    detailVisible.value = true
  } catch (err) {
    codeHandler(err)
  }
}

const detailVisible = ref(false)
const current = ref<ResumeItem | null>(null)

// 对话框内的操作按钮：先关弹窗再走审核流程
function closeAndSetStatus(r: ResumeItem, status: number) {
  detailVisible.value = false
  setStatus(r, status)
}

async function setStatus(r: ResumeItem, status: number) {
  const action = status === 1 ? '设为待考核' : status === 2 ? '通过（将生成注册邀请码）' : '拒绝'
  try {
    await ElMessageBox.confirm(`确定将该简历${action}吗？将通过邮件通知申请人。`, '简历审核', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    const updated = await adminResumeStatus(r.id, status)
    addMessage('操作成功', 'success')
    if (status === 2 && updated.invite_code) {
      ElMessageBox.alert(
        `邀请码：<b style="font-size:20px;color:#007bff">${updated.invite_code}</b><br/><br/>已发送到申请人邮箱，也可手动转发给申请人。`,
        '注册邀请码',
        { dangerouslyUseHTMLString: true, confirmButtonText: '知道了' },
      ).catch(() => {})
    }
    load()
  } catch (err) {
    codeHandler(err)
  }
}

async function removeResume(r: ResumeItem) {
  try {
    await ElMessageBox.confirm(`确定删除「${r.real_name}」的简历记录吗？`, '删除简历', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await adminDeleteResume(r.id)
    addMessage('已删除', 'success')
    load()
  } catch (err) {
    codeHandler(err)
  }
}

onMounted(load)
</script>

<template>
  <div class="pt-[60px] min-h-screen bg-gray-50">
    <div class="flex">
      <AdminSidebar v-model="collapsed" />
      <div class="flex-1 min-w-0 px-4 md:px-8 py-8">
        <button
          class="md:hidden mb-4 px-3 py-1.5 border rounded-lg bg-white text-sm"
          @click="collapsed = false"
        >
          <i class="fa-solid fa-bars mr-1" />菜单
        </button>

        <div class="mb-6">
          <h1 class="text-2xl font-bold">简历管理</h1>
          <p class="text-sm text-gray-500 mt-1">
            审核投递的简历：待考核 / 通过（发放邀请码）/ 拒绝
          </p>
        </div>

        <!-- 工具栏 -->
        <div
          class="bg-white rounded-lg border border-gray-200 p-4 mb-4 flex flex-col md:flex-row gap-3"
        >
          <el-input
            v-model="keyword"
            placeholder="搜索姓名、学号或邮箱"
            clearable
            class="md:!w-72"
            @keyup.enter="search"
            @clear="search"
          />
          <el-select
            v-model="statusFilter"
            placeholder="全部状态"
            clearable
            class="md:!w-36"
            @change="search"
          >
            <el-option
              v-for="s in statusOptions"
              :key="s.value"
              :value="s.value"
              :label="s.label"
            />
          </el-select>
          <button
            class="px-4 py-2 rounded-lg bg-black text-white text-sm md:ml-auto"
            @click="search"
          >
            <i class="fa-solid fa-magnifying-glass mr-1" />搜索
          </button>
        </div>

        <!-- 表格 -->
        <div class="bg-white rounded-lg border border-gray-200 overflow-hidden" v-loading="loading">
          <el-table :data="resumes" style="width: 100%">
            <el-table-column prop="id" label="ID" width="90" show-overflow-tooltip />
            <el-table-column label="照片" width="70">
              <template #default="{ row }">
                <img
                  :src="avatarUrl(row.avatar)"
                  class="w-10 h-10 rounded-lg object-cover"
                  alt=""
                />
              </template>
            </el-table-column>
            <el-table-column prop="real_name" label="姓名" width="90" />
            <el-table-column prop="student_no" label="学号" width="110" show-overflow-tooltip />
            <el-table-column prop="email" label="邮箱" min-width="170" show-overflow-tooltip />
            <el-table-column prop="grade" label="年级" width="70" />
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="statusTag[row.status]?.type ?? 'info'" size="small">
                  {{ statusTag[row.status]?.label ?? row.status_name }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="投递时间" width="160">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="230" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openDetail(row)"
                  >查看</el-button
                >
                <el-button
                  v-if="row.status === 0"
                  link
                  type="warning"
                  size="small"
                  @click="setStatus(row, 1)"
                  >待考核</el-button
                >
                <el-button
                  v-if="row.status === 0 || row.status === 1"
                  link
                  type="success"
                  size="small"
                  @click="setStatus(row, 2)"
                  >通过</el-button
                >
                <el-button
                  v-if="row.status === 0 || row.status === 1"
                  link
                  type="danger"
                  size="small"
                  @click="setStatus(row, -1)"
                  >拒绝</el-button
                >
                <el-button link type="danger" size="small" @click="removeResume(row)"
                  >删除</el-button
                >
              </template>
            </el-table-column>
          </el-table>

          <div class="flex justify-center py-4">
            <el-pagination
              layout="total, prev, pager, next"
              :total="total"
              :page-size="count"
              :current-page="page"
              @current-change="
                (p: number) => {
                  page = p
                  load()
                }
              "
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="简历详情" width="680px" top="5vh" append-to-body>
      <div v-if="current" class="max-h-[72vh] overflow-y-auto pr-1">
        <div class="flex items-center gap-4 mb-6">
          <img
            :src="avatarUrl(current.avatar)"
            class="w-20 h-20 rounded-lg object-cover border"
            alt=""
          />
          <div class="space-y-1 text-sm">
            <div class="text-lg font-bold">{{ current.real_name }}</div>
            <div class="text-gray-500">
              学号：{{ current.student_no }} · 年级：{{ current.grade }}
            </div>
            <div class="text-gray-500">邮箱：{{ current.email }}</div>
            <div class="text-gray-500">投递：{{ formatDateTime(current.created_at) }}</div>
          </div>
          <el-tag class="ml-auto" :type="statusTag[current.status]?.type ?? 'info'">
            {{ statusTag[current.status]?.label ?? current.status_name }}
          </el-tag>
        </div>

        <div class="space-y-5 text-sm">
          <div>
            <div class="font-bold text-gray-700 mb-1.5">个人介绍</div>
            <p class="text-gray-600 whitespace-pre-wrap leading-relaxed">
              {{ current.extra.information || '（未填写）' }}
            </p>
          </div>
          <div>
            <div class="font-bold text-gray-700 mb-1.5">专业能力</div>
            <p class="text-gray-600 whitespace-pre-wrap leading-relaxed">
              {{ current.extra.skills || '（未填写）' }}
            </p>
          </div>
          <div>
            <div class="font-bold text-gray-700 mb-1.5">加入理由</div>
            <p class="text-gray-600 whitespace-pre-wrap leading-relaxed">
              {{ current.extra.reason || '（未填写）' }}
            </p>
          </div>
          <div>
            <div class="font-bold text-gray-700 mb-1.5">对竞赛的理解</div>
            <p class="text-gray-600 whitespace-pre-wrap leading-relaxed">
              {{ current.extra.understanding || '（未填写）' }}
            </p>
          </div>
          <div>
            <div class="font-bold text-gray-700 mb-1.5">未来计划</div>
            <p class="text-gray-600 whitespace-pre-wrap leading-relaxed">
              {{ current.extra.future_plan || '（未填写）' }}
            </p>
          </div>
        </div>

        <div class="flex gap-3 mt-8 pt-4 border-t">
          <button
            v-if="current.status === 0"
            class="px-5 py-2 rounded-lg bg-yellow-500 text-white text-sm font-medium hover:bg-yellow-600"
            @click="closeAndSetStatus(current, 1)"
          >
            设为待考核
          </button>
          <button
            v-if="current.status === 0 || current.status === 1"
            class="px-5 py-2 rounded-lg bg-green-600 text-white text-sm font-medium hover:bg-green-700"
            @click="closeAndSetStatus(current, 2)"
          >
            通过简历
          </button>
          <button
            v-if="current.status === 0 || current.status === 1"
            class="px-5 py-2 rounded-lg bg-red-500 text-white text-sm font-medium hover:bg-red-600"
            @click="closeAndSetStatus(current, -1)"
          >
            拒绝简历
          </button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>
