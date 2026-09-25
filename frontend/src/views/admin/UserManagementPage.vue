<script setup lang="ts">
// 用户管理：服务端关键词/角色筛选、编辑、改角色、删除
import { onMounted, reactive, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import { adminUserList, adminUpdateUser, adminSetRole, adminDeleteUser } from '@/api/user'
import type { AdminUser } from '@/api/user'
import AdminSidebar from '@/components/AdminSidebar.vue'
import { avatarUrl, formatDateTime } from '@/utils/format'
import { GetRoleLabel } from '@/utils/level'
import { useMessage } from '@/composables/useMessage'

const { addMessage, codeHandler } = useMessage()

const collapsed = ref(false)
const keyword = ref('')
const roleFilter = ref<number | ''>('')
const page = ref(1)
const count = 20
const total = ref(0)
const users = ref<AdminUser[]>([])
const loading = ref(false)

const roleNames = [0, 1, 2, 3, 4].map((r) => ({ value: r, label: GetRoleLabel(r) }))

async function load() {
  loading.value = true
  try {
    const res = await adminUserList({
      keyword: keyword.value.trim() || undefined,
      role: roleFilter.value === '' ? undefined : roleFilter.value,
      page: page.value,
      count,
    })
    users.value = res.users
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

// ---- 编辑 ----
const editVisible = ref(false)
const editForm = reactive({
  id: '',
  username: '',
  email: '',
  real_name: '',
  student_no: '',
  grade: 0,
  xp: 0,
  avatar: '',
  role: 1,
})

function openEdit(u: AdminUser) {
  editForm.id = u.id
  editForm.username = u.username
  editForm.email = u.email
  editForm.real_name = u.real_name
  editForm.student_no = u.student_no
  editForm.grade = u.grade
  editForm.xp = u.xp
  editForm.avatar = u.avatar
  editForm.role = u.role
  editVisible.value = true
}

async function saveEdit() {
  try {
    await adminUpdateUser({
      id: editForm.id,
      username: editForm.username,
      avatar: editForm.avatar,
      grade: editForm.grade,
      xp: editForm.xp,
      real_name: editForm.real_name,
      student_no: editForm.student_no,
    })
    if (editForm.role !== users.value.find((u) => u.id === editForm.id)?.role) {
      await adminSetRole({ id: editForm.id, role: editForm.role })
    }
    addMessage('保存成功', 'success')
    editVisible.value = false
    load()
  } catch (err) {
    codeHandler(err)
  }
}

async function removeUser(u: AdminUser) {
  try {
    await ElMessageBox.confirm(
      `确定删除用户「${u.username}」吗？其发布的帖子与评论将被一并删除。`,
      '删除用户',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await adminDeleteUser({ id: u.id })
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
          <h1 class="text-2xl font-bold">用户管理</h1>
          <p class="text-sm text-gray-500 mt-1">
            管理系统中的所有用户，包括查看、编辑和删除用户信息
          </p>
        </div>

        <!-- 工具栏 -->
        <div
          class="bg-white rounded-lg border border-gray-200 p-4 mb-4 flex flex-col md:flex-row gap-3"
        >
          <el-input
            v-model="keyword"
            placeholder="搜索用户名/邮箱/真实姓名/学号"
            clearable
            class="md:!w-80"
            @keyup.enter="search"
            @clear="search"
          />
          <el-select
            v-model="roleFilter"
            placeholder="全部角色"
            clearable
            class="md:!w-40"
            @change="search"
          >
            <el-option v-for="r in roleNames" :key="r.value" :value="r.value" :label="r.label" />
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
          <el-table :data="users" style="width: 100%">
            <el-table-column prop="id" label="ID" width="90" />
            <el-table-column label="头像" width="70">
              <template #default="{ row }">
                <img
                  :src="avatarUrl(row.avatar)"
                  class="w-9 h-9 rounded-full object-cover"
                  alt=""
                />
              </template>
            </el-table-column>
            <el-table-column prop="username" label="用户名" min-width="110" show-overflow-tooltip />
            <el-table-column prop="email" label="邮箱" min-width="170" show-overflow-tooltip />
            <el-table-column prop="real_name" label="真实姓名" width="90">
              <template #default="{ row }">{{ row.real_name || '-' }}</template>
            </el-table-column>
            <el-table-column label="角色" width="100">
              <template #default="{ row }">
                <el-tag :type="row.role >= 3 ? 'danger' : row.role === 2 ? 'warning' : 'info'">
                  {{ GetRoleLabel(row.role) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="年级" width="70">
              <template #default="{ row }">{{ row.grade || '未设置' }}</template>
            </el-table-column>
            <el-table-column prop="xp" label="经验值" width="80" />
            <el-table-column label="注册时间" width="160">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button
                  link
                  type="primary"
                  size="small"
                  @click="$router.push(`/profile/${row.id}`)"
                  >主页</el-button
                >
                <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
                <el-button
                  link
                  type="danger"
                  size="small"
                  :disabled="row.role >= 3"
                  @click="removeUser(row)"
                  >删除</el-button
                >
              </template>
            </el-table-column>
          </el-table>

          <div class="flex justify-center py-4">
            <el-pagination
              layout="total, prev, pager, next, sizes"
              :total="total"
              :page-size="count"
              :current-page="page"
              :page-sizes="[10, 20, 50, 100]"
              @current-change="
                (p: number) => {
                  page = p
                  load()
                }
              "
              @size-change="
                (s: number) => {
                  page = 1
                  load()
                }
              "
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editVisible" title="编辑用户信息" width="480px" append-to-body>
      <el-form label-width="90px">
        <el-form-item label="头像链接">
          <el-input v-model="editForm.avatar" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" disabled />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" />
        </el-form-item>
        <el-form-item label="真实姓名">
          <el-input v-model="editForm.real_name" />
        </el-form-item>
        <el-form-item label="学号">
          <el-input v-model="editForm.student_no" />
        </el-form-item>
        <el-form-item label="年级">
          <el-input-number v-model="editForm.grade" :min="0" :max="99" />
        </el-form-item>
        <el-form-item label="经验值">
          <el-input-number v-model="editForm.xp" :min="0" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" class="!w-full">
            <el-option v-for="r in roleNames" :key="r.value" :value="r.value" :label="r.label" />
          </el-select>
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
