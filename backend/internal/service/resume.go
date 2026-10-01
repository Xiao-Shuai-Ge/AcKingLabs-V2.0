package service

import (
	"errors"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"acking/internal/app"
	"acking/internal/model"
	"acking/internal/repo"
	"acking/pkg/emailkit"
	"acking/pkg/randkit"
	"acking/pkg/response"
)

// 简历状态文案（全站唯一一份，前端引用相同语义）
var resumeStatusNames = map[int]string{
	model.ResumePending:   "待审核",
	model.ResumeApproved:  "已通过",
	model.ResumeRejected:  "未通过",
}

func ResumeStatusName(status int) string { return resumeStatusNames[status] }

// ResumeExtraMax 单项补充信息字数上限
const ResumeExtraMax = 2000

// resumeTransitionAllowed 管理员审核的状态流转：仅 待审核 -> 已通过 / 未通过
func resumeTransitionAllowed(from, to int) bool {
	if from != model.ResumePending {
		return false
	}
	return to == model.ResumeApproved || to == model.ResumeRejected
}

// resumeStatusEditable 投递人可修改简历的状态：
// 待审核（继续修改）、未通过（修改后重新投递）；已通过则账号已开通，不可在此改。
func resumeStatusEditable(status int) bool {
	return status == model.ResumePending || status == model.ResumeRejected
}

type ResumeSubmitReq struct {
	Avatar    string
	RealName  string
	Grade     int
	StudentNo string
	Email     string
	Username  string
	Extra     model.ResumeExtra
}

func validateResume(req *ResumeSubmitReq) error {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.RealName = strings.TrimSpace(req.RealName)
	req.StudentNo = strings.TrimSpace(req.StudentNo)
	req.Avatar = strings.TrimSpace(req.Avatar)
	req.Username = strings.TrimSpace(req.Username)
	if !emailRe.MatchString(req.Email) {
		return response.NewErrMsg(response.CodeBadRequest, "邮箱格式不正确")
	}
	if n := len([]rune(req.RealName)); n < 2 || n > 20 {
		return response.NewErrMsg(response.CodeBadRequest, "真实姓名长度需在 2~20 之间")
	}
	if req.Grade < 0 || req.Grade > 99 {
		return response.NewErrMsg(response.CodeBadRequest, "年级取值 0~99")
	}
	if req.StudentNo == "" {
		return response.NewErrMsg(response.CodeBadRequest, "学号不能为空")
	}
	if n := len([]rune(req.Username)); n < 2 || n > 30 {
		return response.NewErrMsg(response.CodeBadRequest, "用户名长度需在 2~30 之间")
	}
	for _, s := range []string{req.Extra.Information, req.Extra.Skills, req.Extra.Reason, req.Extra.Understanding, req.Extra.FuturePlan} {
		if len([]rune(s)) > ResumeExtraMax {
			return response.NewErrMsg(response.CodeBadRequest, "简历单项内容最多 2000 字")
		}
	}
	return nil
}

// SubmitResume 投递简历（邮箱验证码校验）。审核通过时系统生成随机初始密码并邮件发放。
func SubmitResume(req ResumeSubmitReq, code string) (int64, error) {
	if err := validateResume(&req); err != nil {
		return 0, err
	}
	if err := VerifyCode(req.Email, code); err != nil {
		return 0, err
	}
	// 已注册的邮箱无需再投递（账号已存在）
	if u, err := repo.GetUserByEmail(app.DB, req.Email); err != nil {
		return 0, err
	} else if u != nil {
		return 0, response.NewErrMsg(response.CodeEmailRegistered, "该邮箱已注册账号，无需投递简历")
	}
	if exist, err := repo.GetResumeByEmail(app.DB, req.Email); err != nil {
		return 0, err
	} else if exist != nil {
		if exist.Status == model.ResumeRejected {
			return 0, response.NewErrMsg(response.CodeResumeExists, "该邮箱的简历未通过，请验证邮箱后修改并重新投递")
		}
		return 0, response.NewErr(response.CodeResumeExists)
	}
	if exists, err := repo.UsernameExists(app.DB, req.Username, 0); err != nil {
		return 0, err
	} else if exists {
		return 0, response.NewErr(response.CodeUsernameTaken)
	}
	r := model.Resume{
		Avatar: req.Avatar, RealName: req.RealName, Grade: req.Grade,
		StudentNo: req.StudentNo, Email: req.Email,
		Username: req.Username,
		Extra: req.Extra, Status: model.ResumePending,
	}
	if err := repo.CreateResume(app.DB, &r); err != nil {
		if isDuplicateEntry(err) {
			return 0, response.NewErr(response.CodeResumeExists)
		}
		return 0, err
	}
	notifyAdminsNewResume(&r)
	return r.ID, nil
}

// notifyAdminsNewResume 给开启了"新简历邮件提醒"的管理员逐个发提醒（失败仅记日志，不影响投递）
func notifyAdminsNewResume(r *model.Resume) {
	if !app.Cfg.EmailConfigured() {
		return
	}
	admins, err := repo.ListAdmins(app.DB)
	if err != nil {
		slog.Error("新简历提醒：查询管理员失败", "err", err)
		return
	}
	for i := range admins {
		if !admins[i].NotifyOf().NewResumeEmail {
			continue
		}
		if err := emailkit.SendNewResumeNotify(admins[i].Email, r.RealName, r.Email); err != nil {
			slog.Error("新简历提醒邮件发送失败", "admin", admins[i].Email, "err", err)
		}
	}
}

// UpdateResume 修改简历（仅待审核/未通过可改；未通过时保存即重新投递，状态重置为待审核）
func UpdateResume(req ResumeSubmitReq, code string, resumeID int64) error {
	if err := validateResume(&req); err != nil {
		return err
	}
	if err := VerifyCode(req.Email, code); err != nil {
		return err
	}
	r, err := repo.GetResumeByID(app.DB, resumeID)
	if err != nil {
		return response.NewErr(response.CodeNotFound)
	}
	if r.Email != req.Email {
		return response.NewErrMsg(response.CodeNoPermission, "只能修改本人邮箱对应的简历")
	}
	if !resumeStatusEditable(r.Status) {
		return response.NewErrMsg(response.CodeConflict, "简历已通过审核、账号已开通，不能在此修改")
	}
	if r.Username != req.Username {
		if exists, err := repo.UsernameExists(app.DB, req.Username, 0); err != nil {
			return err
		} else if exists {
			return response.NewErr(response.CodeUsernameTaken)
		}
	}
	r.Avatar, r.RealName, r.Grade = req.Avatar, req.RealName, req.Grade
	r.StudentNo, r.Extra, r.Username = req.StudentNo, req.Extra, req.Username
	r.Status = model.ResumePending
	return repo.SaveResume(app.DB, r)
}

// ResumeDTO 简历详情
type ResumeDTO struct {
	ID         string            `json:"id"`
	Avatar     string            `json:"avatar"`
	RealName   string            `json:"real_name"`
	Grade      int               `json:"grade"`
	StudentNo  string            `json:"student_no"`
	Email      string            `json:"email"`
	Username   string            `json:"username"`
	Extra      model.ResumeExtra `json:"extra"`
	Status     int               `json:"status"`
	StatusName string            `json:"status_name"`
	CreatedAt  int64             `json:"created_at"`
	UpdatedAt  int64             `json:"updated_at"`
}

func toResumeDTO(r *model.Resume) ResumeDTO {
	return ResumeDTO{
		ID: formatID(r.ID), Avatar: r.Avatar, RealName: r.RealName, Grade: r.Grade,
		StudentNo: r.StudentNo, Email: r.Email, Username: r.Username, Extra: r.Extra,
		Status: r.Status, StatusName: ResumeStatusName(r.Status),
		CreatedAt: r.CreatedAt.UnixMilli(), UpdatedAt: r.UpdatedAt.UnixMilli(),
	}
}

// GetResumeBySelf 投递向导第一步：验证邮箱（验证码不消耗，有效期延长到 30 分钟），
// 返回当前简历状态；已注册的邮箱直接提示去登录。
func GetResumeBySelf(email, code string) (*ResumeDTO, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := VerifyCodeKeep(email, code); err != nil {
		return nil, err
	}
	if u, err := repo.GetUserByEmail(app.DB, email); err != nil {
		return nil, err
	} else if u != nil {
		return nil, response.NewErrMsg(response.CodeEmailRegistered, "该邮箱已注册账号，请直接登录（忘记密码可在登录页找回）")
	}
	r, err := repo.GetResumeByEmail(app.DB, email)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, response.NewErrMsg(response.CodeNotFound, "该邮箱还没有投递过简历")
	}
	dto := toResumeDTO(r)
	return &dto, nil
}

// AdminGetResume 管理员查看简历详情
func AdminGetResume(resumeID int64) (*ResumeDTO, error) {
	r, err := repo.GetResumeByID(app.DB, resumeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewErr(response.CodeNotFound)
		}
		return nil, err
	}
	dto := toResumeDTO(r)
	return &dto, nil
}

// AdminResumeList 管理后台简历列表
func AdminResumeList(keyword string, status, page, count int) ([]ResumeDTO, int64, int64, error) {
	if status == 0 {
		status = -100 // 0 是合法状态"待审核"，用 -100 表示"全部"
	}
	list, total, err := repo.ResumeListPage(app.DB, keyword, status, page, count)
	if err != nil {
		return nil, 0, 0, err
	}
	out := make([]ResumeDTO, 0, len(list))
	for i := range list {
		out = append(out, toResumeDTO(&list[i]))
	}
	pageTotal := (total + int64(fCountPublic(count)) - 1) / int64(fCountPublic(count))
	return out, total, pageTotal, nil
}

func fCountPublic(c int) int {
	if c <= 0 || c > 100 {
		return 20
	}
	return c
}

func AdminDeleteResume(resumeID int64) error {
	if _, err := repo.GetResumeByID(app.DB, resumeID); err != nil {
		return response.NewErr(response.CodeNotFound)
	}
	return app.DB.Delete(&model.Resume{}, resumeID).Error
}

// AdminSetResumeStatus 管理员审核：
//
//	待审核(0) -> 已通过(1：事务内自动开通账号) / 未通过(-1)
//	已通过 / 未通过为终态；未通过的简历可由本人修改后重新投递回到待审核
func AdminSetResumeStatus(resumeID int64, target int) (*ResumeDTO, error) {
	if target != model.ResumeApproved && target != model.ResumeRejected {
		return nil, response.NewErrMsg(response.CodeBadRequest, "目标状态不合法")
	}
	r, err := repo.GetResumeByID(app.DB, resumeID)
	if err != nil {
		return nil, response.NewErr(response.CodeNotFound)
	}
	if !resumeTransitionAllowed(r.Status, target) {
		return nil, response.NewErrMsg(response.CodeConflict, "该简历已处理完成，不能再次变更")
	}

	if target == model.ResumeApproved {
		// 开通账号（随机初始密码随邮件发放）+ 置为已通过，必须在一个事务里
		initialPwd := randkit.AlnumCode(10)
		err = app.DB.Transaction(func(tx *gorm.DB) error {
			fresh, err := repo.GetResumeByID(tx, resumeID)
			if err != nil {
				return err
			}
			if !resumeTransitionAllowed(fresh.Status, target) {
				return response.NewErrMsg(response.CodeConflict, "该简历已处理完成，不能再次变更")
			}
			// 邮箱已有账号（如已用邀请码直接注册）：无需也无法再开通
			if u, err := repo.GetUserByEmail(tx, fresh.Email); err != nil {
				return err
			} else if u != nil {
				return response.NewErrMsg(response.CodeConflict, "该邮箱已有账号，无需开通，请直接删除该简历")
			}
			if exists, err := repo.UsernameExists(tx, fresh.Username, 0); err != nil {
				return err
			} else if exists {
				return response.NewErrMsg(response.CodeUsernameTaken, "该用户名已被注册，请通知申请者修改简历后重新审核")
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(initialPwd), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			u := model.User{
				Username: fresh.Username, Password: string(hash), Email: fresh.Email,
				RealName: fresh.RealName, Grade: fresh.Grade, StudentNo: fresh.StudentNo,
				Role: model.RoleUser,
			}
			if err := repo.CreateUser(tx, &u); err != nil {
				return err
			}
			fresh.Status = model.ResumeApproved
			return repo.SaveResume(tx, fresh)
		})
		if err != nil {
			if isDuplicateEntry(err) {
				return nil, response.NewErrMsg(response.CodeConflict, "该邮箱已有账号，无需开通，请直接删除该简历")
			}
			return nil, err
		}
		r.Status = model.ResumeApproved
		// 通过邮件携带初始密码；发送失败仅记日志（账号已开通，可走登录页找回密码）
		if app.Cfg.EmailConfigured() {
			if err := emailkit.SendResumeApproved(r.Email, r.Username, initialPwd); err != nil {
				slog.Error("简历通过邮件发送失败", "email", r.Email, "err", err)
			}
		}
	} else {
		r.Status = model.ResumeRejected
		if err := repo.SaveResume(app.DB, r); err != nil {
			return nil, err
		}
		// 拒绝邮件（失败不影响状态变更，只记日志）
		if app.Cfg.EmailConfigured() {
			if err := emailkit.SendResumeReject(r.Email); err != nil {
				slog.Error("拒绝邮件发送失败", "email", r.Email, "err", err)
			}
		}
	}

	dto := toResumeDTO(r)
	return &dto, nil
}
