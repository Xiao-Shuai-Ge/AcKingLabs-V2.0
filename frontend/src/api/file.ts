// 图片上传（multipart）。后端返回相对路径时按同源处理（开发走 vite 代理）。
import { getAccessToken, BACKEND_URL } from './http'

export async function uploadImage(file: File): Promise<string> {
  const form = new FormData()
  form.append('file', file)
  const resp = await fetch(BACKEND_URL + '/api/file/upload', {
    method: 'POST',
    headers: { Authorization: `Bearer ${getAccessToken()}` },
    body: form,
  })
  const body = await resp.json()
  if (body.code !== 0) throw new Error(body.message || '上传失败')
  return body.data.url as string
}
