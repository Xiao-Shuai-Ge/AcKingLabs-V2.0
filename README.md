# AcKing 学习分享平台

AcKing 算法竞赛实验室的校内学习分享平台：每周打卡、学习分享、比赛聚合、招新与管理后台。

## 功能

- **认证**：邮箱验证码注册 / 登录，JWT 双令牌（access + refresh 静默续期），找回密码
- **打卡（周记）**：每周一篇学习周记，固定打卡窗口（周日 12:00 ~ 周二 12:00，北京时间），公开双倍经验，支持私密
- **学习**：教程 / 题解 / 比赛 / 求助 / 闲聊帖，Markdown 编辑（代码高亮 + KaTeX 公式）、自动草稿、图片上传、@提及，关键词搜索
- **互动**：两级评论、点赞（帖子与评论）、精选、管理推荐、优质解答、浏览量统计（按访客 30 分钟去重）
- **比赛**：聚合 Codeforces / AtCoder / 牛客近期赛事（30 分钟定时抓取）、手动创建、开赛前 20 分钟邮件预约提醒、比赛关联题解
- **招新**：邮箱验证码投递简历、管理员审核（待考核 / 通过发放一次性邀请码 / 拒绝）、邀请码注册
- **消息**：站内通知（点赞 / 评论 / 提及 / 求助帖广播 / 系统消息），按偏好可关，系统消息可同步邮件
- **排行榜**：经验值排名
- **个人主页**：资料、等级 / 称号、Codeforces 分数、获奖经历、打卡与帖子列表
- **管理后台**：用户管理（角色 / 经验 / 删除级联）、内容管理（隐藏 / 精选 / 删除）、简历审核
- **图片上传**：尺寸优先压缩（长边 1600px + 高质量重采样），阿里云 OSS 存储（未配置时自动落本地磁盘）

## 技术栈

| 端 | 技术 |
| --- | --- |
| 后端 | Go 1.24、Gin、GORM + MySQL 8、go-redis、JWT、robfig/cron、slog |
| 前端 | Vue 3.5 + TypeScript、Vite、Tailwind CSS、Element Plus、Pinia、v-md-editor |
| 存储 | MySQL（业务数据）、Redis（验证码 / 浏览量去重计数）、OSS 或本地磁盘（图片） |

## 目录结构

```
├── README.md
├── AGENTS.md            # 开发规范
├── backend/             # Go 后端
│   ├── cmd/server/      # 入口
│   ├── internal/        # config / model / repo / service / api / router / middleware / task / database
│   ├── pkg/             # response(统一响应) / jwtkit / emailkit / weekcode(周记周期算法) / randkit
│   └── config.example.yaml
└── frontend/            # Vue3 前端
    └── src/
        ├── api/         # 所有后端调用（一个模块一个文件）
        ├── components/  # 复用组件
        ├── views/       # 页面（views/admin/ 为管理后台）
        ├── stores/      # Pinia
        ├── utils/       # 纯函数工具
        └── router/      # 路由与守卫
```

## 快速开始

### 1. 准备依赖

本地启动 MySQL 与 Redis。

### 2. 启动后端

```bash
cd backend
cp config.example.yaml config.yaml   # 按需修改数据库 / Redis / JWT 密钥等
go run ./cmd/server
```

- 首次启动自动建表（`database.migrate: true`）；users 表为空时按 `admin` 配置创建超级管理员。
- 邮箱未配置时验证码只打后端日志，方便本地自测。
- 图片上传：`oss` 留空则存本地 `data/uploads`。

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev          # http://localhost:5173，/api 与 /uploads 已代理到 localhost:8080
```

## 配置说明（config.yaml）

| 段 | 说明 |
| --- | --- |
| `app` | 监听地址、对外 base_url（邮件链接 / 上传 URL 用）、上传与静态目录、信任代理 |
| `database` | MySQL DSN；`migrate` 控制启动时自动建表补列 |
| `redis` | 未配置 / 不可用时自动降级（验证码与浏览量去重走进程内实现） |
| `jwt` | 令牌签名密钥，生产环境务必修改 |
| `email` | SMTP，用于验证码 / 简历审核通知 / 比赛预约提醒 |
| `oss` | 阿里云 OSS，配置完整后图片上传走 OSS，否则本地磁盘 |
| `admin` | 首次启动播种的超管账号 |
| `invitation` | 全局注册邀请码（简历通过的申请人会收到专属一次性邀请码） |

## 部署

```bash
# 后端：编译对应平台二进制，随 config.yaml 一起部署
cd backend && GOOS=linux GOARCH=amd64 go build -o acking ./cmd/server

# 前端：构建产物由 nginx 托管，/api、/uploads 反代到后端
cd frontend && npm run build    # 产物在 dist/
```

## 质量检查

```bash
# 后端（backend/ 下）
go test ./...                 # 单元测试（周期算法 / 令牌 / 响应封装 / 纯函数）
go vet ./...                  # 官方静态检查
golangci-lint run ./...       # 综合静态检查（配置见 .golangci.yml）

# 前端（frontend/ 下）
npm run test                  # Vitest 单元测试（周期算法 / 等级体系 / 工具函数）
npm run lint                  # ESLint（flat config，见 eslint.config.mjs）
npm run format:check          # Prettier 格式检查（npm run format 一键格式化）
npm run build                 # vue-tsc 类型检查 + vite 构建
```

推送 GitHub 后 CI（`.github/workflows/ci.yml`）会自动执行以上全部检查。
