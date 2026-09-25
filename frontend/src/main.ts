import './assets/tailwind.css'
import '@fortawesome/fontawesome-free/css/all.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

// Markdown 编辑器与预览（GitHub 主题 + 代码高亮 + KaTeX 公式）
import VMdEditor from '@kangc/v-md-editor'
import VMdPreview from '@kangc/v-md-editor/lib/preview'
import '@kangc/v-md-editor/lib/style/base-editor.css'
import githubTheme from '@kangc/v-md-editor/lib/theme/github.js'
import '@kangc/v-md-editor/lib/theme/style/github.css'
import createKatexPlugin from '@kangc/v-md-editor/lib/plugins/katex/cdn'
import katex from 'katex'
import 'katex/dist/katex.css'
import hljs from 'highlight.js'

import App from './App.vue'
import router from './router'

const katexPlugin = createKatexPlugin({ katex })
VMdEditor.use(githubTheme, { Hljs: hljs })
VMdEditor.use(katexPlugin)
VMdPreview.use(githubTheme, { Hljs: hljs })
VMdPreview.use(katexPlugin)

const app = createApp(App)
app.use(VMdEditor).use(VMdPreview).use(router).use(createPinia()).use(ElementPlus)
app.mount('#app')
