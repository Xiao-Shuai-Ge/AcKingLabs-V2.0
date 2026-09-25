/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, any>
  export default component
}

// @kangc/v-md-editor 未提供完整类型声明
declare module '@kangc/v-md-editor' {
  const VMdEditor: any
  export default VMdEditor
}
declare module '@kangc/v-md-editor/lib/preview' {
  const VMdPreview: any
  export default VMdPreview
}
declare module '@kangc/v-md-editor/lib/theme/github.js' {
  const theme: any
  export default theme
}
declare module '@kangc/v-md-editor/lib/plugins/katex/cdn' {
  const createKatexPlugin: (opt: any) => any
  export default createKatexPlugin
}
