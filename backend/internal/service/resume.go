package service

import (
	"errors"
	"log/slog"
	"strings"

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
	model.ResumePending:   "待处理",
	model.ResumeAssessing: "待考核",
	model.ResumeAccepted:  "已通过",
	model.ResumeRejected:  "未通过",
}

func ResumeStatusName(status int) string { return resumeStatusNames[status] }

// ResumeExtraMax 单项补充信息字数上限
const ResumeExtraMax = 2000

type ResumeSubmitReq struct {
	Avatar    string
	RealName  string
	Grade     int
	StudentNo string
	Email     string
	Extra     model.ResumeExtra
}

func validateResume(req *ResumeSubmitReq) error {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.RealName = strings.TrimSpace(req.RealName)
	req.StudentNo = strings.TrimSpace(req.StudentNo)
	req.Avatar = strings.TrimSpace(req.Avatar)
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
	for _, s := range []string{req.Extra.Information, req.Extra.Skills, req.Extra.Reason, req.Extra.Understanding, req.Extra.FuturePlan} {
		if len([]rune(s)) > ResumeExtraMax {
			return response.NewErrMsg(response.CodeBadRequest, "简历单项内容最多 2000 字")
		}
	}
	return nil
}

// SubmitResume 投递简历（邮箱验证码校验）
func SubmitResume(req ResumeSubmitReq, code string) (int64, error) {
	if err := validateResume(&req); err != nil {
		return 0, err
	}
	if err := VerifyCode(req.Email, code); err != nil {
		return 0, err
	}
	exist, err := repo.GetResumeByEmail(app.DB, req.Email)
	if err != nil {
		return 0, err
	}
	if exist != nil {
		return 0, response.NewErr(response.CodeResumeExists)
	}
	r := model.Resume{
		Avatar: req.Avatar, RealName: req.RealName, Grade: req.Grade,
		StudentNo: req.StudentNo, Email: req.Email, Extra: req.Extra,
		Status: model.ResumePending,
	}
	if err := repo.CreateResume(app.DB, &r); err != nil {
		if isDuplicateEntry(err) {
			return 0, response.NewErr(response.CodeResumeExists)
		}
		return 0, err
	}
	return r.ID, nil
}

// UpdateResume 修改简历（仅待处理状态可改）
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
	if r.Status != model.ResumePending {
		return response.NewErrMsg(response.CodeConflict, "简历已进入审核流程，不能修改")
	}
	r.Avatar, r.RealName, r.Grade = req.Avatar, req.RealName, req.Grade
	r.StudentNo, r.Extra = req.StudentNo, req.Extra
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
	Extra      model.ResumeExtra `json:"extra"`
	Status     int               `json:"status"`
	StatusName string            `json:"status_name"`
	InviteCode string            `json:"invite_code,omitempty"` // 仅管理员可见
	CreatedAt  int64             `json:"created_at"`
	UpdatedAt  int64             `json:"updated_at"`
}

func toResumeDTO(r *model.Resume, isAdmin bool) ResumeDTO {
	dto := ResumeDTO{
		ID: formatID(r.ID), Avatar: r.Avatar, RealName: r.RealName, Grade: r.Grade,
		StudentNo: r.StudentNo, Email: r.Email, Extra: r.Extra,
		Status: r.Status, StatusName: ResumeStatusName(r.Status),
		CreatedAt: r.CreatedAt.UnixMilli(), UpdatedAt: r.UpdatedAt.UnixMilli(),
	}
	if isAdmin {
		dto.InviteCode = r.InviteCode
	}
	return dto
}

// GetResumeBySelf 投递人凭邮箱 + 验证码查询自己的简历
func GetResumeBySelf(email, code string) (*ResumeDTO, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	r, err := repo.GetResumeByEmail(app.DB, email)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, response.NewErrMsg(response.CodeNotFound, "该邮箱还没有投递过简历")
	}
	if err := VerifyCode(email, code); err != nil {
		return nil, err
	}
	dto := toResumeDTO(r, false)
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
	dto := toResumeDTO(r, true)
	return &dto, nil
}

// AdminResumeList 管理后台简历列表
func AdminResumeList(keyword string, status, page, count int) ([]ResumeDTO, int64, int64, error) {
	if status == 0 {
		status = -100 // 0 是合法状态"待处理"，用 -100 表示"全部"
	}
	list, total, err := repo.ResumeListPage(app.DB, keyword, status, page, count)
	if err != nil {
		return nil, 0, 0, err
	}
	out := make([]ResumeDTO, 0, len(list))
	for i := range list {
		out = append(out, toResumeDTO(&list[i], false))
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

// AdminSetResumeStatus 状态流转：
//
//	待处理(0) -> 待考核(1) / 已通过(2) / 未通过(-1)
//	待考核(1) -> 已通过(2) / 未通过(-1)
//	已通过 / 未通过 为终态
func AdminSetResumeStatus(resumeID int64, target int) (*ResumeDTO, error) {
	r, err := repo.GetResumeByID(app.DB, resumeID)
	if err != nil {
		return nil, response.NewErr(response.CodeNotFound)
	}
	if r.Status == model.ResumeAccepted || r.Status == model.ResumeRejected {
		return nil, response.NewErrMsg(response.CodeConflict, "该简历已处理完成，不能再次变更")
	}
	if r.Status == model.ResumeAssessing && target == model.ResumeAssessing {
		return nil, response.NewErrMsg(response.CodeConflict, "该简历已处于待考核状态")
	}

	switch target {
	case model.ResumeAssessing:
		r.Status = model.ResumeAssessing
	case model.ResumeAccepted:
		r.Status = model.ResumeAccepted
		// 已有邀请码则复用（例如之前通过过又被打回的场景），否则生成
		if r.InviteCode == "" {
			r.InviteCode = randkit.LetterCode(6)
		}
	case model.ResumeRejected:
		r.Status = model.ResumeRejected
	default:
		return nil, response.NewErrMsg(response.CodeBadRequest, "目标状态不合法")
	}

	if err := repo.SaveResume(app.DB, r); err != nil {
		return nil, err
	}

	// 邮件通知（失败不影响状态变更，只记日志）
	if app.Cfg.EmailConfigured() {
		switch target {
		case model.ResumeAssessing:
			if err := emailkit.SendResumePending(r.Email); err != nil {
				slog.Error("待考核邮件发送失败", "email", r.Email, "err", err)
			}
		case model.ResumeAccepted:
			if err := emailkit.SendInviteCode(r.Email, r.InviteCode); err != nil {
				slog.Error("邀请码邮件发送失败", "email", r.Email, "err", err)
			}
		case model.ResumeRejected:
			if err := emailkit.SendResumeReject(r.Email); err != nil {
				slog.Error("拒绝邮件发送失败", "email", r.Email, "err", err)
			}
		}
	}

	dto := toResumeDTO(r, true)
	return &dto, nil
}
