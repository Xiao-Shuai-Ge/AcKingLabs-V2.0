// 跨模块共享的类型定义
import type { UserBrief } from './user'

export interface Award {
  name: string
  level: number // 1金 2银 3铜
}

export interface PostItem {
  id: string
  user_id: string
  title: string
  content_short: string
  type: string
  source: string
  like_count: number
  comment_count: number
  view_count: number
  is_admin_like: boolean
  is_featured: boolean
  is_private: boolean
  is_hidden?: boolean
  week_code?: string
  hot_score: number
  created_at: number
  updated_at: number
  author: UserBrief
  liked: boolean
}

export interface PostDetail extends PostItem {
  content: string
  can_edit: boolean
  can_delete: boolean
}

export interface CommentItem {
  id: string
  post_id: string
  father_id: string
  reply_to_id: string
  user_id: string
  content: string
  like_count: number
  is_admin_like: boolean
  created_at: number
  author: UserBrief
  reply_to?: string
  liked: boolean
  can_delete: boolean
  children?: CommentItem[]
  children_more?: boolean
}

export interface ContestItem {
  id: string
  title: string
  start_time: number
  end_time: number
  duration: number
  platform: string
  url: string
  is_recommend: boolean
}

export interface ResumeExtra {
  information: string
  skills: string
  reason: string
  understanding: string
  future_plan: string
}

export interface ResumeItem {
  id: string
  avatar: string
  real_name: string
  grade: number
  student_no: string
  email: string
  username: string
  extra: ResumeExtra
  status: number // 0待审核 1已通过(账号已开通) -1未通过(可修改后重新投递)
  status_name: string
  created_at: number
  updated_at: number
}

export interface MessageItem {
  id: string
  type: 'like' | 'comment' | 'system'
  content: string
  url: string
  is_read: boolean
  created_at: number
  sender: UserBrief
}
