import { get, post } from './http'
import type { PostItem, PostDetail, CommentItem } from './types'

export interface PostListResult {
  posts: PostItem[]
  total: number
  page_total: number
}

export interface PostWriteData {
  title: string
  content: string
  type: string
  source?: string
  is_private?: boolean
}

export const createPost = (data: PostWriteData) => post<{ id: string }>('/api/post/create', data)

export const editPost = (postId: string, data: PostWriteData) =>
  post('/api/post/edit', { post_id: postId, ...data })

export const deletePost = (postId: string) => post('/api/post/delete', { post_id: postId })

export const getPostDetail = (id: string) => get<PostDetail>('/api/post/detail', { id })

export const getPostList = (params: {
  type?: string
  sort?: string
  source?: string
  user_id?: string
  keyword?: string
  page?: number
  count?: number
}) => get<PostListResult>('/api/post/list', params)

export const searchPosts = (params: {
  keyword: string
  page?: number
  count?: number
  type?: string
}) => get<PostListResult>('/api/post/search', params)

export const getPostMore = (params: {
  type?: string
  sort?: string
  source?: string
  user_id?: string
  cursor?: string
  count?: number
}) => get<{ posts: PostItem[]; length: number }>('/api/post/more', params)

export const getWeekStatus = () =>
  get<{
    code: string
    name: string
    valid: boolean
    from: number
    to: number
    study_from: number
    study_to: number
  }>('/api/post/week-status')

export const getDiaryWeeks = (userId: string) =>
  get<{ weeks: { post_id: string; week_code: string; is_private: boolean }[] }>(
    '/api/post/diary-weeks',
    { user_id: userId },
  )

export const togglePostLike = (postId: string) =>
  post<{ liked: boolean; like_count: number }>('/api/post/like', { post_id: postId })

// ---- 评论 ----
export const getComments = (params: {
  post_id: string
  father_id?: string
  before?: string
  after?: string
  count?: number
}) => get<{ comments: CommentItem[]; length: number }>('/api/post/comments', params)

export const createComment = (data: {
  post_id: string
  father_id?: string
  reply_to_id?: string
  content: string
}) => post<CommentItem>('/api/post/comment', data)

export const deleteComment = (commentId: string) =>
  post('/api/post/comment/delete', { comment_id: commentId })

export const toggleCommentLike = (commentId: string) =>
  post<{ liked: boolean; like_count: number }>('/api/post/comment/like', { comment_id: commentId })

// ---- 管理端 ----
export const adminPostList = (params: {
  keyword?: string
  type?: string
  status?: string
  page?: number
  count?: number
}) => get<PostListResult>('/api/admin/post/list', params)

export const adminSetPostHidden = (postId: string, value: boolean) =>
  post('/api/admin/post/hide', { post_id: postId, value })

export const adminSetPostFeatured = (postId: string, value: boolean) =>
  post('/api/admin/post/feature', { post_id: postId, value })

export const adminDeletePost = (postId: string) =>
  post('/api/admin/post/delete', { post_id: postId })
