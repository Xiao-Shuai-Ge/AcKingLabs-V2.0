package response

// 业务错误码。0 表示成功；前端 CodeHandler 依赖这份表给出友好提示。
const (
	CodeOK           = 0
	CodeBadRequest   = 40001 // 参数错误
	CodeUnauthorized = 40100 // 未登录 / 令牌无效（前端触发静默刷新或跳登录）
	CodeForbidden    = 40300 // 权限不足
	CodeNotFound     = 40400 // 资源不存在
	CodeConflict     = 40900 // 状态冲突
	CodeRateLimited  = 42900 // 请求过于频繁
	CodeInternal     = 50000 // 服务器内部错误

	CodeVerifyWrong     = 41001 // 验证码错误或已过期
	CodeEmailRegistered = 41002 // 邮箱已被注册
	CodeLoginFailed     = 41003 // 邮箱或密码错误
	CodeInviteInvalid   = 41004 // 邀请码无效
	CodeDiaryExists     = 41005 // 本周周记已提交
	CodeNotDiaryTime    = 41006 // 当前不在打卡时间窗口内
	CodeContentTooLong  = 41007 // 内容超出长度限制
	CodeNoPermission    = 41008 // 无权操作该资源
	CodeResumeExists    = 41009 // 该邮箱已投递过简历
	CodeContestURLDup   = 41010 // 比赛链接已存在
	CodeUsernameTaken   = 41011 // 用户名已被占用
)

var codeMessages = map[int]string{
	CodeOK:              "成功",
	CodeBadRequest:      "参数错误",
	CodeUnauthorized:    "请先登录",
	CodeForbidden:       "权限不足",
	CodeNotFound:        "资源不存在",
	CodeConflict:        "状态冲突",
	CodeRateLimited:     "请求过于频繁，请稍后再试",
	CodeInternal:        "服务器开小差了，请稍后再试",
	CodeVerifyWrong:     "验证码错误或已过期",
	CodeEmailRegistered: "该邮箱已被注册",
	CodeLoginFailed:     "邮箱或密码错误",
	CodeInviteInvalid:   "邀请码无效",
	CodeDiaryExists:     "本周周记已提交",
	CodeNotDiaryTime:    "当前不在打卡时间窗口内",
	CodeContentTooLong:  "内容超出长度限制",
	CodeNoPermission:    "无权操作该资源",
	CodeResumeExists:    "该邮箱已投递过简历",
	CodeContestURLDup:   "比赛链接已存在",
	CodeUsernameTaken:   "用户名已被占用",
}

// Message 返回错误码默认文案
func Message(code int) string {
	if m, ok := codeMessages[code]; ok {
		return m
	}
	return "未知错误"
}
