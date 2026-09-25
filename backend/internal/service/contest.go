package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"

	"acking/internal/app"
	"acking/internal/model"
	"acking/internal/repo"
	"acking/pkg/emailkit"
	"acking/pkg/response"
)

// ContestItem 比赛 DTO
type ContestItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Duration    int64  `json:"duration"`
	Platform    string `json:"platform"`
	Url         string `json:"url"`
	IsRecommend bool   `json:"is_recommend"`
}

func toContestItem(c model.Contest) ContestItem {
	return ContestItem{
		ID: formatID(c.ID), Title: c.Title, StartTime: c.StartTime, EndTime: c.EndTime,
		Duration: c.Duration, Platform: c.Platform, Url: c.Url, IsRecommend: c.IsRecommend,
	}
}

// GetContestList 比赛列表
func GetContestList(platform string, page, count int) ([]ContestItem, int64, int64, error) {
	if page < 1 {
		page = 1
	}
	if count <= 0 || count > 50 {
		count = 5
	}
	list, total, err := repo.ContestListPage(app.DB, platform, page, count)
	if err != nil {
		return nil, 0, 0, err
	}
	items := make([]ContestItem, 0, len(list))
	for _, c := range list {
		items = append(items, toContestItem(c))
	}
	pageTotal := (total + int64(count) - 1) / int64(count)
	return items, total, pageTotal, nil
}

// GetContestDetail 比赛详情
func GetContestDetail(id int64) (*ContestItem, error) {
	c, err := repo.GetContestByID(app.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewErr(response.CodeNotFound)
		}
		return nil, err
	}
	item := toContestItem(*c)
	return &item, nil
}

// ToggleBooking 预约/取消预约
func ToggleBooking(userID int64, contestID int64) (booked bool, err error) {
	if _, err := repo.GetContestByID(app.DB, contestID); err != nil {
		return false, response.NewErr(response.CodeNotFound)
	}
	exists, err := repo.IsBooked(app.DB, contestID, userID)
	if err != nil {
		return false, err
	}
	if exists {
		if err := repo.DeleteBooking(app.DB, contestID, userID); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := repo.CreateBooking(app.DB, &model.Booking{ContestID: contestID, UserID: userID}); err != nil {
		if isDuplicateEntry(err) {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

// GetBookedMap 批量查询预约状态
func GetBookedMap(userID int64, contestIDs []int64) (map[string]bool, error) {
	m, err := repo.BookedMap(app.DB, userID, contestIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[formatID(k)] = v
	}
	return out, nil
}

// ---- 管理端 ----

type ContestWriteReq struct {
	Title     string
	StartTime int64 // ms
	EndTime   int64 // ms
	Url       string
}

func AdminCreateContest(req ContestWriteReq) (int64, error) {
	req.Url = strings.TrimSpace(req.Url)
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || req.Url == "" || req.StartTime <= 0 || req.EndTime <= req.StartTime {
		return 0, response.NewErrMsg(response.CodeBadRequest, "请填写完整的比赛信息（标题/起止时间/链接）")
	}
	if exists, err := repo.GetContestURLExists(app.DB, req.Url, 0); err != nil {
		return 0, err
	} else if exists {
		return 0, response.NewErr(response.CodeContestURLDup)
	}
	c := model.Contest{
		Url: req.Url, Platform: "AcKing", Title: req.Title,
		StartTime: req.StartTime, EndTime: req.EndTime,
		Duration: (req.EndTime - req.StartTime) / 1000,
		Source:   "manual",
	}
	if err := repo.UpsertContest(app.DB, &c); err != nil {
		if isDuplicateEntry(err) {
			return 0, response.NewErr(response.CodeContestURLDup)
		}
		return 0, err
	}
	// UpsertContest 未回填自增 ID，按 url 查一次
	exist, err := repo.GetContestURL(app.DB, req.Url)
	if err != nil {
		return 0, err
	}
	return exist.ID, nil
}

func AdminUpdateContest(contestID int64, req ContestWriteReq) error {
	c, err := repo.GetContestByID(app.DB, contestID)
	if err != nil {
		return response.NewErr(response.CodeNotFound)
	}
	req.Url = strings.TrimSpace(req.Url)
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || req.Url == "" || req.StartTime <= 0 || req.EndTime <= req.StartTime {
		return response.NewErrMsg(response.CodeBadRequest, "请填写完整的比赛信息")
	}
	if exists, err := repo.GetContestURLExists(app.DB, req.Url, contestID); err != nil {
		return err
	} else if exists {
		return response.NewErr(response.CodeContestURLDup)
	}
	c.Title, c.Url = req.Title, req.Url
	c.StartTime, c.EndTime = req.StartTime, req.EndTime
	c.Duration = (req.EndTime - req.StartTime) / 1000
	return repo.SaveContest(app.DB, c)
}

func AdminDeleteContest(contestID int64) error {
	return repo.DeleteContestCascade(app.DB, contestID)
}

func AdminSetRecommend(contestID int64, recommend bool) error {
	return repo.UpdateContestRecommend(app.DB, contestID, recommend)
}

// ---- 定时任务 ----

// ScrapeContests 抓取三个平台的近期比赛并入库（30 分钟一次）
func ScrapeContests() {
	var all []model.Contest
	all = append(all, fetchCodeforces()...)
	all = append(all, fetchAtCoder()...)
	all = append(all, fetchNowcoder()...)
	saved := 0
	for i := range all {
		if err := repo.UpsertContest(app.DB, &all[i]); err != nil {
			slog.Error("比赛入库失败", "title", all[i].Title, "err", err)
			continue
		}
		saved++
	}
	slog.Info("比赛抓取完成", "fetched", len(all), "saved", saved)
}

func httpGet(url string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 AcKingBot/2.0")
	return client.Do(req)
}

func fetchCodeforces() []model.Contest {
	resp, err := httpGet("https://codeforces.com/api/contest.list?gym=false", 15*time.Second)
	if err != nil {
		slog.Error("Codeforces API 请求失败", "err", err)
		return nil
	}
	defer resp.Body.Close()
	var res struct {
		Status string `json:"status"`
		Result []struct {
			ID               int    `json:"id"`
			Name             string `json:"name"`
			Phase            string `json:"phase"`
			StartTimeSeconds int64  `json:"startTimeSeconds"`
			DurationSeconds  int64  `json:"durationSeconds"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil || res.Status != "OK" {
		slog.Error("Codeforces API 解析失败", "err", err)
		return nil
	}
	var out []model.Contest
	limit := time.Now().UnixMilli() + 7*24*3600*1000
	for _, c := range res.Result {
		start := c.StartTimeSeconds * 1000
		if c.Phase != "BEFORE" || start > limit {
			continue
		}
		out = append(out, model.Contest{
			Platform: "Codeforces", Title: c.Name,
			StartTime: start, EndTime: start + c.DurationSeconds*1000,
			Duration: c.DurationSeconds, Url: "https://codeforces.com/contests/" + strconv.Itoa(c.ID),
			Source: "scrape",
		})
	}
	return out
}

// fetchAtCoder 解析 AtCoder 近期比赛页。
// 页面时间文本自带 +0900 偏移（如 2026-09-20 21:00:00+0900），
// 直接按偏移解析为准确时刻，不要手工加减小时。
func fetchAtCoder() []model.Contest {
	resp, err := httpGet("https://atcoder.jp/contests/", 15*time.Second)
	if err != nil {
		slog.Error("AtCoder 页面请求失败", "err", err)
		return nil
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		slog.Error("AtCoder HTML 解析失败", "err", err)
		return nil
	}
	var out []model.Contest
	limit := time.Now().UnixMilli() + 7*24*3600*1000
	doc.Find("div.table-responsive table tbody tr").Each(func(i int, row *goquery.Selection) {
		cols := row.Find("td")
		if cols.Length() < 3 {
			return
		}
		title := strings.TrimSpace(cols.Eq(1).Find("a").First().Text())
		if !strings.HasPrefix(title, "AtCoder Beginner Contest") && !strings.HasPrefix(title, "AtCoder Regular Contest") {
			return
		}
		timeText := strings.TrimSpace(cols.Eq(0).Find("a").First().Text())
		start, err := time.Parse("2006-01-02 15:04:05-0700", timeText)
		if err != nil {
			slog.Warn("AtCoder 时间解析失败", "text", timeText)
			return
		}
		startMs := start.UnixMilli()
		if startMs < time.Now().UnixMilli() || startMs > limit {
			return
		}
		// 时长列格式 "01:40"
		var h, m int
		if _, err := fmt.Sscanf(strings.TrimSpace(cols.Eq(2).Text()), "%d:%d", &h, &m); err != nil {
			h, m = 1, 40
		}
		duration := int64(h*3600 + m*60)
		href, _ := cols.Eq(1).Find("a").First().Attr("href")
		out = append(out, model.Contest{
			Platform: "AtCoder", Title: title,
			StartTime: startMs, EndTime: startMs + duration*1000,
			Duration: duration, Url: "https://atcoder.jp" + href,
			Source: "scrape",
		})
	})
	return out
}

func fetchNowcoder() []model.Contest {
	now := time.Now()
	url := fmt.Sprintf("https://ac.nowcoder.com/acm/calendar/contest?token=&month=%d-%d&_=%.3f",
		now.Year(), int(now.Month()), float64(now.UnixNano())/1e9)
	resp, err := httpGet(url, 15*time.Second)
	if err != nil {
		slog.Error("牛客 API 请求失败", "err", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		slog.Error("牛客响应读取失败", "err", err)
		return nil
	}
	var res struct {
		Code int `json:"code"`
		Data []struct {
			ContestName string `json:"contestName"`
			StartTime   int64  `json:"startTime"`
			EndTime     int64  `json:"endTime"`
			Link        string `json:"link"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		slog.Error("牛客 JSON 解析失败", "err", err)
		return nil
	}
	var out []model.Contest
	limit := time.Now().UnixMilli() + 7*24*3600*1000
	for _, c := range res.Data {
		if c.StartTime <= time.Now().UnixMilli() || c.StartTime > limit || c.Link == "" {
			continue
		}
		out = append(out, model.Contest{
			Platform: "Nowcoder", Title: c.ContestName,
			StartTime: c.StartTime, EndTime: c.EndTime,
			Duration: (c.EndTime - c.StartTime) / 1000, Url: c.Link,
			Source: "scrape",
		})
	}
	return out
}

// NotifyBookings 开赛前 20 分钟邮件提醒（1 分钟一次）。
// 发送后只标记 notified，预约状态保留。
func NotifyBookings() {
	if !app.Cfg.EmailConfigured() {
		return
	}
	bookings, err := repo.PendingNotifyBookings(app.DB, time.Now().UnixMilli())
	if err != nil {
		slog.Error("查询待提醒预约失败", "err", err)
		return
	}
	if len(bookings) == 0 {
		return
	}
	var done []int64
	for _, b := range bookings {
		c, err := repo.GetContestByID(app.DB, b.ContestID)
		if err != nil {
			continue
		}
		u, err := repo.GetUserByID(app.DB, b.UserID)
		if err != nil || u.Email == "" {
			done = append(done, b.ID) // 无法投递也标记，避免每分钟重试
			continue
		}
		minutes := (c.StartTime - time.Now().UnixMilli()) / 60000
		if minutes < 0 {
			minutes = 0
		}
		if err := emailkit.SendBookingNotice(u.Email, c.Url, c.Title, minutes); err != nil {
			slog.Error("预约提醒发送失败", "user", u.ID, "err", err)
		}
		done = append(done, b.ID)
	}
	if err := repo.MarkBookingsNotified(app.DB, done); err != nil {
		slog.Error("预约提醒标记失败", "err", err)
	}
}
