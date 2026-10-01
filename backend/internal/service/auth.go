// Package service 业务逻辑层。
package service

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"acking/internal/app"
	"acking/internal/config"
	"acking/internal/model"
	"acking/internal/repo"
	"acking/pkg/emailkit"
	"acking/pkg/jwtkit"
	"acking/pkg/randkit"
	"acking/pkg/response"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ---- 验证码存取（Redis 优先，不可用时退化为进程内存，单机部署可用） ----

const (
	codeTTL      = 5 * time.Minute
	codeMaxTries = 5
	codeSendGap  = 60 * time.Second // 同一邮箱 60s 内只能发一次
)

func codeKey(email string) string { return "code:email:" + email }

type memCode struct {
	value   string
	expires time.Time
}
type memCodeEntry struct {
	code     memCode
	tries    int
	lastSend time.Time
}

var (
	memCodeMu sync.Mutex
	memCodes  = make(map[string]*memCodeEntry)
)

func codeGet(email string) (string, bool) {
	if app.RedisOK() {
		v, err := app.RDB.Get(context.Background(), codeKey(email)).Result()
		return v, err == nil
	}
	memCodeMu.Lock()
	defer memCodeMu.Unlock()
	e, ok := memCodes[email]
	if !ok || time.Now().After(e.code.expires) {
		return "", false
	}
	return e.code.value, true
}

func codeSet(email, code string) {
	if app.RedisOK() {
		app.RDB.Set(context.Background(), codeKey(email), code, codeTTL)
		return
	}
	memCodeMu.Lock()
	defer memCodeMu.Unlock()
	memCodes[email] = &memCodeEntry{
		code:     memCode{value: code, expires: time.Now().Add(codeTTL)},
		lastSend: time.Now(),
	}
}

// codeConsumeTries 验证失败计数 +1，返回是否仍可继续尝试
func codeConsumeTries(email string) bool {
	if app.RedisOK() {
		ctx := context.Background()
		key := codeKey(email) + ":tries"
		n, err := app.RDB.Incr(ctx, key).Result()
		if err == nil {
			app.RDB.Expire(ctx, key, codeTTL)
			return n <= codeMaxTries
		}
		return true
	}
	memCodeMu.Lock()
	defer memCodeMu.Unlock()
	e, ok := memCodes[email]
	if !ok {
		memCodes[email] = &memCodeEntry{}
		e = memCodes[email]
	}
	e.tries++
	return e.tries <= codeMaxTries
}

// codeCanSend 发送频率限制（SETNX EX 60）
func codeCanSend(email string) bool {
	if app.RedisOK() {
		ok, err := app.RDB.SetNX(context.Background(), codeKey(email)+":sent", 1, codeSendGap).Result()
		return err == nil && ok
	}
	memCodeMu.Lock()
	defer memCodeMu.Unlock()
	e, ok := memCodes[email]
	if ok && time.Since(e.lastSend) < codeSendGap {
		return false
	}
	if !ok {
		memCodes[email] = &memCodeEntry{}
		e = memCodes[email]
	}
	e.lastSend = time.Now()
	return true
}

func codeDelete(email string) {
	if app.RedisOK() {
		ctx := context.Background()
		app.RDB.Del(ctx, codeKey(email), codeKey(email)+":tries", codeKey(email)+":sent")
		return
	}
	memCodeMu.Lock()
	defer memCodeMu.Unlock()
	delete(memCodes, email)
}

// VerifyCode 校验验证码（消耗尝试次数；成功后删除）
func VerifyCode(email, code string) error {
	return verifyCodeImpl(email, code, false)
}

// VerifyCodeKeep 校验验证码但不消耗，成功后把有效期延长到 resumeCodeTTL。
// 供投递简历两步向导使用：第一步验证邮箱、第二步提交复用同一个验证码。
func VerifyCodeKeep(email, code string) error {
	return verifyCodeImpl(email, code, true)
}

func verifyCodeImpl(email, code string, keep bool) error {
	want, ok := codeGet(email)
	if !ok {
		return response.NewErr(response.CodeVerifyWrong)
	}
	if code != want {
		if !codeConsumeTries(email) {
			codeDelete(email)
			return response.NewErrMsg(response.CodeVerifyWrong, "验证码错误次数过多，请重新获取")
		}
		return response.NewErr(response.CodeVerifyWrong)
	}
	if keep {
		codeExtendTTL(email, resumeCodeTTL)
		return nil
	}
	codeDelete(email)
	return nil
}

// resumeCodeTTL 向导第一步验证后给填表预留的验证码有效期
const resumeCodeTTL = 30 * time.Minute

// codeExtendTTL 延长验证码及尝试计数的有效期（仅对仍存在的验证码生效）
func codeExtendTTL(email string, ttl time.Duration) {
	if app.RedisOK() {
		ctx := context.Background()
		app.RDB.Expire(ctx, codeKey(email), ttl)
		app.RDB.Expire(ctx, codeKey(email)+":tries", ttl)
		return
	}
	memCodeMu.Lock()
	defer memCodeMu.Unlock()
	if e, ok := memCodes[email]; ok && time.Now().Before(e.code.expires) {
		e.code.expires = time.Now().Add(ttl)
	}
}

// ---- 认证业务 ----

// SendCode 发送邮箱验证码
func SendCode(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailRe.MatchString(email) {
		return response.NewErrMsg(response.CodeBadRequest, "邮箱格式不正确")
	}
	if !codeCanSend(email) {
		return response.NewErrMsg(response.CodeRateLimited, "发送太频繁，请稍后再试")
	}
	code := randkit.DigitCode(6)
	codeSet(email, code)
	if !app.Cfg.EmailConfigured() {
		// 开发环境未配置邮箱：验证码只打日志，方便自测
		slog.Warn("邮箱未配置，验证码未真正发送（仅日志）", "email", email, "code", code)
		return nil
	}
	if err := emailkit.SendCode(email, code); err != nil {
		slog.Error("验证码邮件发送失败", "email", email, "err", err)
		return response.NewErrMsg(response.CodeInternal, "验证码发送失败，请稍后重试")
	}
	return nil
}

type RegisterReq struct {
	Email    string
	Code     string
	Password string
	Username string
	Invite   string
}

// Register 注册（邀请码直接注册通道；无邀请码的用户走投递简历，审核通过自动开通账号）
func Register(req RegisterReq) (access, refresh string, err error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	req.Invite = strings.TrimSpace(req.Invite)
	if !emailRe.MatchString(req.Email) {
		return "", "", response.NewErrMsg(response.CodeBadRequest, "邮箱格式不正确")
	}
	if len(req.Code) != 6 {
		return "", "", response.NewErrMsg(response.CodeBadRequest, "验证码格式不正确")
	}
	if n := len([]rune(req.Username)); n < 2 || n > 30 {
		return "", "", response.NewErrMsg(response.CodeBadRequest, "用户名长度需在 2~30 之间")
	}
	if l := len(req.Password); l < 6 || l > 30 {
		return "", "", response.NewErrMsg(response.CodeBadRequest, "密码长度需在 6~30 之间")
	}
	// 只认全局邀请码；放在验证码校验之前，避免输错邀请码烧掉邮箱验证码
	if req.Invite == "" || req.Invite != app.Cfg.Invitation.Code {
		return "", "", response.NewErrMsg(response.CodeInviteInvalid, "邀请码无效；没有邀请码可投递简历，审核通过后自动开通账号")
	}

	if err := VerifyCode(req.Email, req.Code); err != nil {
		return "", "", err
	}

	// 邮箱唯一
	if u, err := repo.GetUserByEmail(app.DB, req.Email); err != nil {
		return "", "", err
	} else if u != nil {
		return "", "", response.NewErr(response.CodeEmailRegistered)
	}
	if exists, err := repo.UsernameExists(app.DB, req.Username, 0); err != nil {
		return "", "", err
	} else if exists {
		return "", "", response.NewErr(response.CodeUsernameTaken)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hash),
		Role:     model.RoleUser,
	}
	if err := repo.CreateUser(app.DB, &user); err != nil {
		return "", "", err
	}

	return jwtkit.GeneratePair(user.ID, user.Role, false)
}

// Login 邮箱密码登录
func Login(email, password string, remember bool) (access, refresh string, err error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := repo.GetUserByEmail(app.DB, email)
	if err != nil {
		return "", "", err
	}
	// 用户不存在与密码错误统一提示，避免账号枚举
	if u == nil {
		return "", "", response.NewErr(response.CodeLoginFailed)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return "", "", response.NewErr(response.CodeLoginFailed)
	}
	return jwtkit.GeneratePair(u.ID, u.Role, remember)
}

// Refresh 用刷新令牌换新 access；角色从数据库重读，保证提权/降级及时生效
func Refresh(refreshToken string) (string, error) {
	claims, err := jwtkit.Parse(refreshToken, jwtkit.TypeRefresh)
	if err != nil {
		return "", response.NewErr(response.CodeUnauthorized)
	}
	u, err := repo.GetUserByID(app.DB, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", response.NewErr(response.CodeUnauthorized)
		}
		return "", err
	}
	return jwtkit.Generate(u.ID, u.Role, jwtkit.TypeAccess, config.AccessTTL)
}

// ResetPassword 找回密码
func ResetPassword(email, code, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if l := len(password); l < 6 || l > 30 {
		return response.NewErrMsg(response.CodeBadRequest, "密码长度需在 6~30 之间")
	}
	u, err := repo.GetUserByEmail(app.DB, email)
	if err != nil {
		return err
	}
	if u == nil {
		return response.NewErr(response.CodeLoginFailed)
	}
	if err := VerifyCode(email, code); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return repo.UpdateUserColumns(app.DB, u.ID, map[string]interface{}{"password": string(hash)})
}
