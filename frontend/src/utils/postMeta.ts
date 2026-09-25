// 帖子类型元信息（名称与描边式角标配色）

export interface PostMeta {
  key: string
  name: string
  // 描边式角标：文字色 + 边框色
  cls: string
}

export const postMetaMap: Record<string, PostMeta> = {
  diary: { key: 'diary', name: '周记', cls: 'text-gray-600 border-gray-400' },
  official: { key: 'official', name: '官方', cls: 'text-red-500 border-red-500' },
  tutorial: { key: 'tutorial', name: '教程', cls: 'text-blue-500 border-blue-500' },
  solution: { key: 'solution', name: '题解', cls: 'text-green-600 border-green-600' },
  contest: { key: 'contest', name: '比赛', cls: 'text-purple-500 border-purple-500' },
  help: { key: 'help', name: '求助', cls: 'text-yellow-600 border-yellow-600' },
  fun: { key: 'fun', name: '闲聊', cls: 'text-pink-500 border-pink-500' },
}

export function postTypeName(type: string): string {
  return postMetaMap[type]?.name ?? type
}

// 描边式角标 class（border-2 rounded-md px-1）
export function postTypeClass(type: string): string {
  const meta = postMetaMap[type]
  return meta ? meta.cls : 'text-gray-600 border-gray-400'
}
