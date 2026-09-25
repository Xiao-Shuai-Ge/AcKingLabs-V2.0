package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"acking/internal/app"
	"acking/internal/model"
	"acking/internal/repo"
	"acking/pkg/response"
)

// AuthorBrief 列表内嵌的作者信息（消灭 N+1）
type AuthorBrief struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Xp       int    `json:"xp"`
	Role     int    `json:"role"`
}

func toBrief(u *model.User) *AuthorBrief {
	if u == nil {
		return &AuthorBrief{Username: "已注销用户"}
	}
	return &AuthorBrief{
		ID:       u.ID,
		Username: u.Username,
		Avatar:   u.Avatar,
		Xp:       u.Xp,
		Role:     u.Role,
	}
}

func authorMap(userIDs []int64) map[int64]*model.User {
	m, err := repo.GetUsersByIDs(app.DB, userIDs)
	if err != nil {
		slog.Error("批量获取用户失败", "err", err)
		return map[int64]*model.User{}
	}
	return m
}

// Me 当前登录用户基础信息
func Me(userID int64) (*AuthorBrief, error) {
	u, err := repo.GetUserByID(app.DB, userID)
	if err != nil {
		return nil, err
	}
	return toBrief(u), nil
}

// BatchInfo 批量基础信息（公开字段）
func BatchInfo(ids []int64) ([]*AuthorBrief, error) {
	m := authorMap(ids)
	out := make([]*AuthorBrief, 0, len(ids))
	for _, id := range ids {
		out = append(out, toBrief(m[id]))
	}
	return out, nil
}

// ProfileDTO 个人主页数据（学校内部平台，实名信息按站长决定公开展示）
type ProfileDTO struct {
	AuthorBrief
	Grade            int           `json:"grade"`
	StudentNo        string        `json:"student_no,omitempty"`
	RealName         string        `json:"real_name,omitempty"`
	CodeforcesID     string        `json:"codeforces_id"`
	CodeforcesRating int           `json:"codeforces_rating"`
	Signature        string        `json:"signature"`
	Awards           []model.Award `json:"awards"`
}

func GetProfile(_ int64, _ int, targetID int64) (*ProfileDTO, error) {
	u, err := repo.GetUserByID(app.DB, targetID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NewErr(response.CodeNotFound)
		}
		return nil, err
	}
	// Codeforces 分数懒刷新（2 小时一次）
	refreshCodeforcesRating(u)

	dto := &ProfileDTO{
		AuthorBrief:      *toBrief(u),
		Grade:            u.Grade,
		RealName:         u.RealName,
		StudentNo:        u.StudentNo,
		CodeforcesID:     u.CodeforcesID,
		CodeforcesRating: u.CodeforcesRating,
		Signature:        u.Signature,
		Awards:           u.Awards,
	}
	if u.Awards == nil {
		dto.Awards = []model.Award{}
	}
	return dto, nil
}

// UpdateProfileReq 修改资料；实名三件套（real_name/grade/student_no）仅管理员可改
type UpdateProfileReq struct {
	OperatorID   int64
	OperatorRole int
	TargetID     int64
	Username     *string
	Avatar       *string
	Signature    *string
	CodeforcesID *string
	Awards       *[]model.Award
	Grade        *int
	StudentNo    *string
	RealName     *string
	Xp           *int // 仅管理员可改
}

func UpdateProfile(req UpdateProfileReq) error {
	if req.OperatorID != req.TargetID && req.OperatorRole < model.RoleAdmin {
		return response.NewErr(response.CodeNoPermission)
	}
	u, err := repo.GetUserByID(app.DB, req.TargetID)
	if err != nil {
		return err
	}
	if req.OperatorRole < model.RoleAdmin {
		req.Grade = nil
		req.StudentNo = nil
		req.RealName = nil
	}
	if req.Username != nil {
		name := strings.TrimSpace(*req.Username)
		if n := len([]rune(name)); n < 2 || n > 30 {
			return response.NewErrMsg(response.CodeBadRequest, "用户名长度需在 2~30 之间")
		}
		exists, err := repo.UsernameExists(app.DB, name, req.TargetID)
		if err != nil {
			return err
		}
		if exists {
			return response.NewErr(response.CodeUsernameTaken)
		}
		u.Username = name
	}
	if req.Avatar != nil {
		if strings.HasSuffix(strings.ToLower(*req.Avatar), ".gif") {
			return response.NewErrMsg(response.CodeBadRequest, "头像不支持 GIF")
		}
		u.Avatar = strings.TrimSpace(*req.Avatar)
	}
	if req.Signature != nil {
		if len([]rune(*req.Signature)) > 255 {
			return response.NewErrMsg(response.CodeBadRequest, "签名最多 255 字")
		}
		u.Signature = *req.Signature
	}
	if req.CodeforcesID != nil {
		u.CodeforcesID = strings.TrimSpace(*req.CodeforcesID)
	}
	if req.Awards != nil {
		if len(*req.Awards) > 20 {
			return response.NewErrMsg(response.CodeBadRequest, "获奖经历最多 20 条")
		}
		for _, a := range *req.Awards {
			if a.Level < 1 || a.Level > 3 {
				return response.NewErrMsg(response.CodeBadRequest, "奖项等级只能是 1金/2银/3铜")
			}
		}
		u.Awards = *req.Awards
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 99 {
			return response.NewErrMsg(response.CodeBadRequest, "年级取值 0~99")
		}
		u.Grade = *req.Grade
	}
	if req.StudentNo != nil {
		u.StudentNo = strings.TrimSpace(*req.StudentNo)
	}
	if req.Xp != nil {
		if *req.Xp < 0 || *req.Xp > 1000000 {
			return response.NewErrMsg(response.CodeBadRequest, "经验值取值不合法")
		}
		u.Xp = *req.Xp
	}
	if req.RealName != nil {
		if n := len([]rune(*req.RealName)); n < 2 || n > 20 {
			return response.NewErrMsg(response.CodeBadRequest, "真实姓名长度需在 2~20 之间")
		}
		u.RealName = strings.TrimSpace(*req.RealName)
	}
	return repo.SaveUser(app.DB, u)
}

// GetUserSettings 读取用户设置（未设置的字段返回默认值）
func GetUserSettings(userID int64) (*model.UserSettings, error) {
	u, err := repo.GetUserByID(app.DB, userID)
	if err != nil {
		return nil, err
	}
	s := &model.UserSettings{Notify: model.DefaultNotifySettings()}
	if u.Settings != nil {
		s.Notify = u.Settings.Notify
	}
	return s, nil
}

// UpdateUserSettings 保存用户设置。
// 注意：不能用 UpdateColumns+map（会绕过字段的 JSON 序列化器），走整对象保存。
func UpdateUserSettings(userID int64, s model.UserSettings) error {
	u, err := repo.GetUserByID(app.DB, userID)
	if err != nil {
		return err
	}
	u.Settings = &s
	return repo.SaveUser(app.DB, u)
}

// RankingItem 排行榜行
type RankingItem struct {
	Rank     int    `json:"rank"`
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Xp       int    `json:"xp"`
	Role     int    `json:"role"`
}

// GetRankings 经验值排行榜（rank 为全榜名次，跨页连续）
func GetRankings(page, count int) ([]RankingItem, int64, int64, error) {
	if page < 1 {
		page = 1
	}
	if count <= 0 || count > 50 {
		count = 10
	}
	users, total, err := repo.UserRankingsPage(app.DB, page, count)
	if err != nil {
		return nil, 0, 0, err
	}
	items := make([]RankingItem, 0, len(users))
	for i, u := range users {
		items = append(items, RankingItem{
			Rank: (page-1)*count + i + 1,
			ID:   strconv.FormatInt(u.ID, 10), Username: u.Username,
			Avatar: u.Avatar, Xp: u.Xp, Role: u.Role,
		})
	}
	pageTotal := (total + int64(count) - 1) / int64(count)
	return items, total, pageTotal, nil
}

// SetRole 修改角色（仅超管）
func SetRole(operatorRole int, targetID int64, role int) error {
	if operatorRole < model.RoleSuper {
		return response.NewErr(response.CodeForbidden)
	}
	if role < model.RoleVisitor || role > model.RoleSuper {
		return response.NewErrMsg(response.CodeBadRequest, "角色取值不合法")
	}
	return repo.UpdateUserColumns(app.DB, targetID, map[string]interface{}{"role": role})
}

// AdminUserItem 管理后台用户行
type AdminUserItem struct {
	ID        int64  `json:"id,string"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Xp        int    `json:"xp"`
	Grade     int    `json:"grade"`
	StudentNo string `json:"student_no"`
	RealName  string `json:"real_name"`
	Role      int    `json:"role"`
	CreatedAt int64  `json:"created_at"`
}

// AdminUserList 管理后台用户列表（服务端关键词/角色过滤
func AdminUserList(keyword string, role *int, page, count int) (items []AdminUserItem, total int64, pageTotal int64, err error) {
	if page < 1 {
		page = 1
	}
	if count <= 0 || count > 100 {
		count = 20
	}
	users, total, err := repo.UserListPage(app.DB, repo.UserListFilter{Keyword: keyword, Role: role, Page: page, Count: count})
	if err != nil {
		return
	}
	items = make([]AdminUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, AdminUserItem{
			ID: u.ID, Username: u.Username, Email: u.Email, Avatar: u.Avatar,
			Xp: u.Xp, Grade: u.Grade, StudentNo: u.StudentNo, RealName: u.RealName,
			Role: u.Role, CreatedAt: u.CreatedAt.UnixMilli(),
		})
	}
	pageTotal = (total + int64(count) - 1) / int64(count)
	return
}

// cfRatingThrottle CF 分数刷新节流（内存，2 小时）
var cfLastRefresh = map[int64]time.Time{}

func refreshCodeforcesRating(u *model.User) {
	if u.CodeforcesID == "" {
		return
	}
	if t, ok := cfLastRefresh[u.ID]; ok && time.Since(t) < 2*time.Hour {
		return
	}
	cfLastRefresh[u.ID] = time.Now()
	go func(handle string, userID int64, old int) {
		url := fmt.Sprintf("https://codeforces.com/api/user.info?handles=%s", handle)
		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return
		}
		var res struct {
			Status string `json:"status"`
			Result []struct {
				Rating int `json:"rating"`
			} `json:"result"`
		}
		if json.Unmarshal(body, &res) != nil || res.Status != "OK" || len(res.Result) == 0 {
			return
		}
		rating := res.Result[0].Rating
		if rating != old {
			if err := repo.UpdateUserColumns(app.DB, userID, map[string]interface{}{"codeforces_rating": rating}); err != nil {
				slog.Error("更新 CF 分数失败", "user", userID, "err", err)
			}
		}
	}(u.CodeforcesID, u.ID, u.CodeforcesRating)
}

// DeleteUserAdmin 删除用户（事务级联
func DeleteUserAdmin(operatorID, targetID int64) error {
	if operatorID == targetID {
		return response.NewErrMsg(response.CodeBadRequest, "不能删除自己")
	}
	u, err := repo.GetUserByID(app.DB, targetID)
	if err != nil {
		return err
	}
	if u.Role >= model.RoleSuper {
		return response.NewErrMsg(response.CodeForbidden, "不能删除超级管理员")
	}
	return app.DB.Transaction(func(tx *gorm.DB) error {
		// 用户发的帖子：软删帖 + 清点赞/评论
		var postIDs []int64
		if err := tx.Model(&model.Post{}).Where("user_id = ?", targetID).Pluck("id", &postIDs).Error; err != nil {
			return err
		}
		for _, pid := range postIDs {
			if err := repo.DeletePostCascadeData(tx, pid); err != nil {
				return err
			}
		}
		// 用户发的评论：软删 + 修正计数
		var comments []model.Comment
		if err := tx.Where("user_id = ?", targetID).Find(&comments).Error; err != nil {
			return err
		}
		for _, cm := range comments {
			if err := repo.DeleteCommentCascadeData(tx, &cm); err != nil {
				return err
			}
		}
		// 点赞清理（顺带回补对应帖子/评论的计数，避免计数漂移）
		var postLikes []model.PostLike
		if err := tx.Where("user_id = ?", targetID).Find(&postLikes).Error; err != nil {
			return err
		}
		postLikeDelta := map[int64]int64{}
		for _, l := range postLikes {
			postLikeDelta[l.PostID]++
		}
		if err := tx.Where("user_id = ?", targetID).Delete(&model.PostLike{}).Error; err != nil {
			return err
		}
		for pid, n := range postLikeDelta {
			if err := tx.Model(&model.Post{}).Where("id = ?", pid).
				Update("like_count", gorm.Expr("GREATEST(like_count - ?, 0)", n)).Error; err != nil {
				return err
			}
		}
		var commentLikes []model.CommentLike
		if err := tx.Where("user_id = ?", targetID).Find(&commentLikes).Error; err != nil {
			return err
		}
		commentLikeDelta := map[int64]int64{}
		for _, l := range commentLikes {
			commentLikeDelta[l.CommentID]++
		}
		if err := tx.Where("user_id = ?", targetID).Delete(&model.CommentLike{}).Error; err != nil {
			return err
		}
		for cid, n := range commentLikeDelta {
			if err := tx.Model(&model.Comment{}).Where("id = ?", cid).
				Update("like_count", gorm.Expr("GREATEST(like_count - ?, 0)", n)).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("user_id = ? OR sender_id = ?", targetID, targetID).Delete(&model.Message{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", targetID).Delete(&model.Booking{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, targetID).Error
	})
}
