// Package emailkit SMTP 邮件发送与业务邮件模板。
// SMTP 配置来自 config；未配置时所有发送直接返回错误（由调用方决定降级策略）。
package emailkit

import (
	"crypto/tls"
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"

	"acking/internal/app"
)

func send(to []string, subject, body string, embeds ...string) error {
	cfg := app.Cfg.Email
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.Username)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	for _, p := range embeds {
		m.Embed(p)
	}
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	d.TLSConfig = &tls.Config{ServerName: cfg.Host}
	return d.DialAndSend(m)
}

// SendCode 邮箱验证码
func SendCode(to, code string) error {
	body := `
<div>
    <p style="text-indent:2em;">你的邮箱验证码为：<strong style="font-size:18px;">` + code + `</strong></p>
    <p style="text-indent:2em;">此验证码的有效期为 5 分钟，请尽快使用。</p>
</div>`
	return send([]string{to}, "[AcKing学习分享平台] 邮箱验证码", body)
}

// SendBookingNotice 比赛开赛提醒
func SendBookingNotice(to string, url, title string, minutes int64) error {
	body := fmt.Sprintf(`
<div>
    <span>你预约的比赛</span>
    <a href="%s" style="margin:2px;">%s</a>
    <span>将在 %d 分钟后开始，请注意准备。</span>
</div>`, url, title, minutes)
	return send([]string{to}, "[AcKing学习分享平台] 比赛预约提醒", body)
}

// SendInviteCode 简历通过 + 注册邀请码（附考核群二维码）
func SendInviteCode(to, code string) error {
	body := fmt.Sprintf(`
<div>
    <p style="text-indent:2em;">恭喜！您已获得 AcKing 内部平台注册资格。</p>
    <p style="text-indent:2em;">您的邀请码为：<strong style="color:#007bff;font-size:18px;">%s</strong></p>
    <p style="text-indent:2em;">请使用此邀请码注册账号，邀请码仅限该邮箱使用。</p>
    <br>
    <p style="text-indent:2em;">可以点击下方链接注册账号：</p>
    <a href="%s/login" style="margin:2px;">%s/login</a>
    <br>
    <p style="text-indent:2em;">如有疑问，请联系管理员。</p>
</div>`, code, app.Cfg.App.BaseURL, app.Cfg.App.BaseURL)
	return send([]string{to}, "[AcKing学习分享平台] 简历通过通知", body, app.Cfg.App.StaticDir+"/images/qr-code.png")
}

// SendResumePending 待考核通知（附考核群二维码）
func SendResumePending(to string) error {
	body := `
<div>
    <p style="text-indent:2em;">恭喜！您的简历已通过初审，等待进入下一轮考核！</p>
    <br>
    <p style="text-indent:2em;">请尽快扫描下方二维码加入考核通知群聊，等待考核通知：</p>
    <img src="cid:qr-code.png" alt="群聊二维码" style="width:100px;height:100px;display:block;margin:0 auto;">
    <br>
    <p style="text-indent:2em;">如有疑问，请联系管理员。</p>
</div>`
	return send([]string{to}, "[AcKing学习分享平台] 待考核通知", body, app.Cfg.App.StaticDir+"/images/qr-code.png")
}

// SendSystemNotice 系统消息邮件（站内消息的邮件同步）
func SendSystemNotice(to, content, url string) error {
	link := url
	if app.Cfg.App.BaseURL != "" {
		link = strings.TrimRight(app.Cfg.App.BaseURL, "/") + url
	}
	body := fmt.Sprintf(`
<div>
    <p style="text-indent:2em;">%s</p>
    <p style="text-indent:2em;"><a href="%s">点击查看详情</a></p>
</div>`, content, link)
	return send([]string{to}, "[AcKing学习分享平台] 站内通知", body)
}

// SendResumeReject 简历未通过
func SendResumeReject(to string) error {
	body := `
<div>
    <p style="text-indent:2em;">经过实验室评审团队的综合评估，很遗憾地通知您，您的简历未通过审核。</p>
    <p style="text-indent:2em;">非常感谢你对 AcKing 算法竞赛实验室的关注与认可🌹🌹🌹</p>
    <p style="text-indent:2em;">欢迎您继续关注我们的其他活动，祝你在算法学习之路上收获更多进步！</p>
    <br>
    <p style="text-indent:2em;">如有疑问，请联系管理员。</p>
</div>`
	return send([]string{to}, "[AcKing学习分享平台] 简历审核结果通知", body)
}
