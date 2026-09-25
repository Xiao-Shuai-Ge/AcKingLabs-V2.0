// 路由与守卫（需要登录的页面未登录时跳转登录页，管理页仅管理员可入）
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  { path: '/', name: 'Home', component: () => import('@/views/HomePage.vue') },
  { path: '/login', name: 'Login', component: () => import('@/views/LoginPage.vue') },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('@/views/ResetPasswordPage.vue'),
  },

  // 打卡（周记）
  {
    path: '/diary',
    name: 'Diary',
    component: () => import('@/views/DiaryPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/diary/create',
    name: 'DiaryCreate',
    component: () => import('@/views/PostCreatePage.vue'),
    meta: { requiresAuth: true, postType: 'diary' },
  },
  {
    path: '/diary/:id',
    name: 'DiaryDetail',
    component: () => import('@/views/PostPage.vue'),
    meta: { postType: 'diary' },
  },
  {
    path: '/diary/:id/edit',
    name: 'DiaryEdit',
    component: () => import('@/views/PostEditPage.vue'),
    meta: { requiresAuth: true, postType: 'diary' },
  },

  // 学习
  { path: '/learn', name: 'Learn', component: () => import('@/views/LearnPage.vue') },
  {
    path: '/learn/create',
    name: 'LearnCreate',
    component: () => import('@/views/PostCreatePage.vue'),
    meta: { requiresAuth: true, postType: 'learn' },
  },
  {
    path: '/learn/:id',
    name: 'LearnDetail',
    component: () => import('@/views/PostPage.vue'),
    meta: { postType: 'learn' },
  },
  {
    path: '/learn/:id/edit',
    name: 'LearnEdit',
    component: () => import('@/views/PostEditPage.vue'),
    meta: { requiresAuth: true, postType: 'learn' },
  },

  // 比赛
  { path: '/contest', name: 'Contest', component: () => import('@/views/ContestPage.vue') },
  {
    path: '/contest/create',
    name: 'ContestCreate',
    component: () => import('@/views/ContestCreatePage.vue'),
    meta: { admin: true },
  },
  {
    path: '/contest/:id',
    name: 'ContestDetail',
    component: () => import('@/views/ContestDetailPage.vue'),
  },

  // 更多
  { path: '/more', name: 'More', component: () => import('@/views/MorePage.vue') },
  { path: '/more/about', name: 'About', component: () => import('@/views/AboutPage.vue') },
  { path: '/more/resume', name: 'Resume', component: () => import('@/views/ResumePage.vue') },
  { path: '/more/rankings', name: 'Rankings', component: () => import('@/views/RankingsPage.vue') },

  // 个人
  { path: '/profile/:id', name: 'Profile', component: () => import('@/views/ProfilePage.vue') },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/message',
    name: 'Message',
    component: () => import('@/views/MessagePage.vue'),
    meta: { requiresAuth: true },
  },

  // 管理后台
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: () => import('@/views/admin/UserManagementPage.vue'),
    meta: { requiresAuth: true, admin: true },
  },
  {
    path: '/admin/posts',
    name: 'AdminPosts',
    component: () => import('@/views/admin/PostManagementPage.vue'),
    meta: { requiresAuth: true, admin: true },
  },
  {
    path: '/admin/resumes',
    name: 'AdminResumes',
    component: () => import('@/views/admin/ResumeManagementPage.vue'),
    meta: { requiresAuth: true, admin: true },
  },

  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach((to) => {
  const user = useUserStore()
  if (to.meta.requiresAuth && !user.isLogin) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.admin && !user.isAdmin) {
    return { path: '/' }
  }
})

export default router
