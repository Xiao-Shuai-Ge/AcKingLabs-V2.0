# AGENTS.md — AcKing 学习分享平台 开发规范

> 给人和 AI 都能看懂的开发约定。写代码前先读一遍，改代码时遵守里面的规矩。

## 项目是什么

AcKing 算法竞赛实验室的校内学习分享平台（周记打卡、学习帖子、比赛聚合、招新简历、管理后台）。
目录结构：

```
AcKingLabs/
├── AGENTS.md          # 本文件
├── backend/           # Go 后端
│   ├── cmd/server/    # 入口 main.go
│   ├── internal/      # config / model / repo / service / api / router / middleware / task / database
│   ├── pkg/           # response(统一响应) jwtkit emailkit weekcode(周记周期算法) randkit
│   └── config.yaml    # 本地配置（gitignore，模板是 config.example.yaml）
└── frontend/          # Vue3 前端
    └── src/
        ├── api/       # 所有后端调用，一个模块一个文件，禁止在页面里裸调 axios
        ├── components/  # 跨页面复用组件
        ├── views/     # 页面（views/admin/ 管理后台）
        ├── stores/    # Pinia
        ├── utils/     # 纯函数工具（周记周期、等级颜色、格式化）
        └── router/    # 路由 + 守卫
```

## 技术栈

- 后端：Go 1.24、Gin、GORM + MySQL 8、go-redis v9、JWT(access 2h / refresh 14d)、cron 定时任务、slog 日志
- 前端：Vue 3.5 `<script setup lang="ts">`、TypeScript、Vite、Tailwind CSS、Element Plus、Pinia、v-md-editor(markdown)

## 红线（禁止事项）

1. **不得提交**真实配置与密钥（config.yaml、JWT secret、邮箱密码、OSS 密钥）。入库的只有 config.example.yaml。
2. 不引入重型中间件（消息队列、搜索引擎等）——当前规模用 MySQL + Redis 足够，确有需要先讨论架构。
3. 不得绕过 `src/api/` 直接在组件里发请求；不得绕过 repo 层在 service 里拼 SQL 字符串（参数化查询除外）。
4. 新依赖要克制：先想能不能用标准库/现有依赖解决；引入新包需要说明理由。

## 后端规矩

1. **分层**：api(参数绑定/取操作人) → service(业务校验) → repo(GORM 查询)。上层不写 SQL，下层不做业务判断。
2. **响应**：永远 HTTP 200 + `{code, message, data}`，`code=0` 成功；业务错误码集中在 `pkg/response/codes.go`，不许在 handler 里随手造数字。错误信息不得透出 SQL/内部错误细节。
3. **鉴权**：路由上挂 `middleware.Auth(role)` 或 `OptionalAuth()`；handler 从 context 取 `user_id/role`，不信任请求体里的身份字段。
4. **写操作一致性**：涉及多表写入（点赞计数、删帖级联、删用户级联）必须放在 GORM 事务里。
5. **唯一性靠数据库**：防重复打卡、防重复点赞、防重复预约都用唯一索引 + 冲突处理，不搞"先查后插"。
6. **时间**：库里存 DATETIME 或毫秒时间戳（比赛起止用毫秒），对外 JSON 统一毫秒。前端不要自己再算时区偏移，用工具函数。
7. **随机数**：验证码/邀请码用 `crypto/rand`。
8. **配置**：所有可变参数进 config.yaml；启动时 AutoMigrate（config.database.migrate 控制）；users 表为空时按 config.admin 播种超管。
9. **限流**：公开接口挂 Limiter；以 UID（已登录）或 IP 计。
10. **日志**：slog，info 记业务事件，error 带上下文（哪个帖子/哪个用户），不打敏感信息（密码、验证码明文）。
11. **测试**：纯逻辑（算法、解析、工具函数）必须带表驱动单元测试；改动 `pkg/weekcode` 等既有算法时同步维护两侧（Go 与 `frontend/src/utils/week.ts`）用例。

## 前端规矩

1. **页面 = views，复用 = components**。帖子卡片这类多处出现的 UI 必须是组件（PostCard），不许复制粘贴。
2. **API 层**：`src/api/http.ts` 统一拦截：自动带 token、401 静默刷新后重放、业务码 toast（`CodeHandler`）。页面只调 `src/api/*.ts` 导出的函数。
3. **路由守卫**：`meta: { requiresAuth }` / `meta: { admin: true }` 在 router 里统一校验，页面内不重复判断。
4. **移动端**：用 Tailwind 响应式断点（`md:`），禁止 UA 嗅探；可点击元素用 click，不许只有 hover。
5. **风格**：浅灰背景 `bg-gray-50`、白卡片、黑色为主按钮色、导航四项（打卡/学习/比赛/更多）各自品牌色、等级颜色体系见 `utils/level.ts`。圆角统一 `rounded-lg`，不要发明不存在的 class。
6. **列表页**：作者信息、点赞态由列表接口一次带回（内嵌），禁止循环里逐条再发请求。
7. **草稿**：编辑器内容自动存 localStorage（按帖子类型分 key），发布成功后清除。
8. **字数限制**：正文/评论上限按角色（见 `utils/contentLimit.ts`），前后端保持一致。
9. **响应式状态**：不要在 getter/computed 里混合 localStorage 等非响应式条件（`&&` 短路会丢依赖，缓存过期值）。

## 数据库规矩

- 表名复数蛇形（users/posts/...），字段蛇形；计数列命名 `xxx_count`。
- 统一 `created_at/updated_at`；用户内容表（users/posts/comments/resumes）带 `deleted_at` 软删除。
- 1:1 的小块配置用 JSON 列挂在主表上（如 users.settings 的分组结构），不单独立表。
- 唯一约束优先于应用层判重；高频排序列建复合索引（如 posts(type, hot_score)）。
- 改表 = 改 model + AutoMigrate；需要手工数据修正时写一次性 SQL 放 `backend/migrations/` 并记录。

## 常用命令

```bash
# 后端
cd backend && go run ./cmd/server          # 启动（读 config.yaml）
go test ./...                               # 单元测试（提交前必过）
golangci-lint run ./...                     # 静态检查（提交前必过）

# 前端
cd frontend && npm install
npm run dev                                 # 开发（先启动后端）
npm run test                                # Vitest 单元测试（提交前必过）
npm run lint                                # ESLint（提交前必过）
npm run format                              # Prettier 一键格式化
npm run build                               # 构建（vue-tsc 类型检查 + vite build）
```

## 提交约定

- 提交消息用 conventional commits 格式：`feat: 新功能`、`fix: 修 bug`、`docs: 文档`、`refactor: 重构`、`test: 测试`、`chore: 杂项`；一个提交做一件事。
- 动数据库结构、改错误码、改 API 字段的提交，必须在 AGENTS.md 里补一句说明。
- 提交前本地过一遍：后端 `go test` + `golangci-lint run`，前端 `npm run test` + `npm run lint`（CI 会跑同样的检查）。

## 变更记录

- 2026-10 账号/简历流程重构：注册仅认全局邀请码（简历专属邀请码废除，`resumes.invite_code` 删除）；简历新增 `username`/`password`（bcrypt），状态枚举重排为 `0待审核 / 1已通过 / -1未通过`（待考核态删除），管理员审核通过时在事务内自动开通账号；未通过的简历可由本人修改后重新投递（update 接口覆盖原记录并重置为待审核）；`users.settings.notify` 新增 `new_resume_email`（管理员专属的新简历邮件提醒，个人设置页开关）。
- 2026-10 二轮调整：投递页改两步向导（先验证邮箱再填简历；`/api/resume/detail` 改为验证码不消耗模式、成功后有效期延长 30 分钟，并拒绝已注册邮箱）；`resumes.password` 列删除，初始密码改为审核通过时随机生成（`randkit.AlnumCode`）随通过邮件发放；新增登录态修改密码接口 `POST /api/user/password`（新错误码 `41012` 原密码错误）；前端注册页主入口改为投递简历，邀请码注册降为特殊通道；更多页移除投递简历入口。
