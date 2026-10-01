// Package router 路由注册。约定：
//
//	/api/auth    认证（无需登录）
//	/api/user    用户信息
//	/api/post    帖子/评论/点赞/打卡
//	/api/contest 比赛
//	/api/resume  简历投递
//	/api/message 站内消息
//	/api/file    文件上传
//	/api/admin/* 管理后台（管理员权限）
package router

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"acking/internal/api"
	"acking/internal/app"
	"acking/internal/middleware"
	"acking/internal/model"
	"acking/internal/task"
)

// Run 启动 HTTP 服务与定时任务
func Run() {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	_ = r.SetTrustedProxies(app.Cfg.App.TrustedProxies)

	r.Use(middleware.CORS(nil))

	// 静态：上传文件与静态资源（二维码等）
	ensureDir(app.Cfg.App.UploadDir)
	r.Static("/uploads", app.Cfg.App.UploadDir)
	r.Static("/static", app.Cfg.App.StaticDir)

	api := r.Group("/api")
	registerRoutes(api)

	task.StartCron()

	addr := fmt.Sprintf("%s:%d", app.Cfg.App.Host, app.Cfg.App.Port)
	slog.Info("HTTP 服务启动", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("HTTP 服务退出", "err", err)
	}
}

func ensureDir(dir string) {
	if dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
}

func registerRoutes(g *gin.RouterGroup) {
	// ---- 认证 ----
	auth := g.Group("/auth")
	{
		auth.POST("/send-code", middleware.Limiter(1.0/15, 4), api.SendCode)
		auth.POST("/register", middleware.Limiter(1.0/15, 4), api.Register)
		auth.POST("/login", middleware.Limiter(1.0/15, 4), api.Login)
		auth.POST("/refresh-token", middleware.Limiter(0.25, 8), api.RefreshToken)
		auth.POST("/reset-password", middleware.Limiter(1.0/15, 4), api.ResetPassword)
	}

	// ---- 用户 ----
	user := g.Group("/user")
	{
		user.GET("/me", middleware.Auth(model.RoleUser), middleware.Limiter(0.5, 10), api.Me)
		user.GET("/batch", middleware.Limiter(2, 10), api.BatchInfo)
		user.GET("/profile", middleware.Limiter(1, 10), api.GetProfile)
		user.POST("/profile", middleware.Auth(model.RoleUser), middleware.Limiter(0.25, 8), api.UpdateProfile)
		user.GET("/rankings", middleware.Limiter(1, 8), api.GetRankings)
		user.GET("/setting", middleware.Auth(model.RoleUser), middleware.Limiter(1, 8), api.GetUserSetting)
		user.POST("/setting", middleware.Auth(model.RoleUser), middleware.Limiter(0.25, 4), api.UpdateUserSetting)
		user.POST("/password", middleware.Auth(model.RoleUser), middleware.Limiter(0.25, 4), api.ChangePassword)
	}

	// ---- 帖子 ----
	post := g.Group("/post")
	{
		post.POST("/create", middleware.Auth(model.RoleUser), middleware.Limiter(1.0/20, 3), api.CreatePost)
		post.POST("/edit", middleware.Auth(model.RoleUser), middleware.Limiter(1.0/20, 3), api.EditPost)
		post.POST("/delete", middleware.Auth(model.RoleUser), middleware.Limiter(1.0/20, 3), api.DeletePost)
		post.GET("/detail", middleware.OptionalAuth(), middleware.Limiter(2, 10), api.GetPostDetail)
		post.GET("/list", middleware.OptionalAuth(), middleware.Limiter(1, 10), api.GetPostList)
		post.GET("/search", middleware.OptionalAuth(), middleware.Limiter(1.0/10, 6), api.SearchPosts)
		post.GET("/more", middleware.OptionalAuth(), middleware.Limiter(1, 10), api.GetPostMore)
		post.GET("/week-status", middleware.Limiter(1, 10), api.GetWeekStatus)
		post.GET("/diary-weeks", middleware.Auth(model.RoleUser), middleware.Limiter(1, 10), api.GetUserDiaryWeeks)
		post.POST("/like", middleware.Auth(model.RoleUser), middleware.Limiter(4, 20), api.TogglePostLike)

		post.GET("/comments", middleware.OptionalAuth(), middleware.Limiter(2, 10), api.GetComments)
		post.POST("/comment", middleware.Auth(model.RoleUser), middleware.Limiter(1, 5), api.CreateComment)
		post.POST("/comment/delete", middleware.Auth(model.RoleUser), middleware.Limiter(1, 5), api.DeleteComment)
		post.POST("/comment/like", middleware.Auth(model.RoleUser), middleware.Limiter(4, 20), api.ToggleCommentLike)
	}

	// ---- 比赛 ----
	contest := g.Group("/contest")
	{
		contest.GET("/list", middleware.Limiter(1, 8), api.GetContestList)
		contest.GET("/detail", middleware.Limiter(1, 8), api.GetContestDetail)
		contest.POST("/booking", middleware.Auth(model.RoleUser), middleware.Limiter(0.5, 4), api.ToggleBooking)
		contest.GET("/booking", middleware.Auth(model.RoleUser), middleware.Limiter(2, 8), api.GetBookingMap)
	}

	// ---- 简历（公开，凭邮箱验证码） ----
	resume := g.Group("/resume")
	{
		resume.POST("/submit", middleware.Limiter(1.0/30, 2), api.SubmitResume)
		resume.POST("/update", middleware.Limiter(1.0/30, 2), api.UpdateResume)
		resume.GET("/detail", middleware.Limiter(0.2, 4), api.GetResumeBySelf)
	}

	// ---- 消息 ----
	msg := g.Group("/message", middleware.Auth(model.RoleUser))
	{
		msg.GET("/count", middleware.Limiter(1, 8), api.GetMessageCounts)
		msg.GET("/list", middleware.Limiter(1, 8), api.GetMessageList)
		msg.POST("/read", middleware.Limiter(1, 8), api.MarkMessageRead)
	}

	// ---- 文件上传（未登录也可用，投递简历头像；登录/未登录分桶限流） ----
	g.POST("/file/upload", middleware.OptionalAuth(),
		middleware.LimiterDual(1.0/20, 5, 1.0/300, 3), api.UploadImage)

	// ---- 管理后台 ----
	admin := g.Group("/admin", middleware.Auth(model.RoleAdmin))
	{
		user := admin.Group("/user")
		{
			user.GET("/list", middleware.Limiter(1, 6), api.AdminUserList)
			user.POST("/update", middleware.Limiter(0.5, 6), api.AdminUpdateUser)
			user.POST("/role", middleware.Limiter(0.5, 6), api.SetRole)
			user.POST("/delete", middleware.Limiter(0.5, 6), api.AdminDeleteUser)
		}
		post := admin.Group("/post")
		{
			post.GET("/list", middleware.Limiter(1, 6), api.AdminPostList)
			post.POST("/hide", middleware.Limiter(0.5, 6), api.AdminSetPostHidden)
			post.POST("/feature", middleware.Limiter(0.5, 6), api.AdminSetPostFeatured)
			post.POST("/delete", middleware.Limiter(0.5, 6), api.AdminDeletePost)
		}
		contest := admin.Group("/contest")
		{
			contest.POST("/create", middleware.Limiter(0.2, 3), api.AdminCreateContest)
			contest.POST("/update", middleware.Limiter(0.2, 3), api.AdminUpdateContest)
			contest.POST("/delete", middleware.Limiter(0.2, 3), api.AdminDeleteContest)
			contest.POST("/recommend", middleware.Limiter(0.5, 6), api.AdminSetRecommend)
		}
		resume := admin.Group("/resume")
		{
			resume.GET("/list", middleware.Limiter(1, 6), api.AdminResumeList)
			resume.GET("/detail", middleware.Limiter(1, 6), api.AdminGetResume)
			resume.POST("/status", middleware.Limiter(0.5, 6), api.AdminSetResumeStatus)
			resume.POST("/delete", middleware.Limiter(0.5, 6), api.AdminDeleteResume)
		}
	}
}
