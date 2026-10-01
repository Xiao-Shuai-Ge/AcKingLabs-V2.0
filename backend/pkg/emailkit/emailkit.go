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

// SendResumeApproved 简历通过：账号已自动开通，初始密码随邮件发放（登录后请修改）
func SendResumeApproved(to, username, initialPassword string) error {
	body := fmt.Sprintf(`
<div>
    <p style="text-indent:2em;">恭喜！您的简历已通过审核，AcKing 学习分享平台账号已自动为您开通。</p>
    <p style="text-indent:2em;">用户名：<strong style="color:#007bff;font-size:18px;">%s</strong></p>
    <p style="text-indent:2em;">初始密码：<strong style="color:#007bff;font-size:18px;">%s</strong></p>
    <p style="text-indent:2em;">请使用本邮箱与初始密码登录，并在「个人设置」中修改密码：</p>
    <a href="%s/login" style="margin:2px;">%s/login</a>
    <br>
    <p style="text-indent:2em;">如有疑问，请联系管理员。</p>
</div>`, username, initialPassword, app.Cfg.App.BaseURL, app.Cfg.App.BaseURL)
	return send([]string{to}, "[AcKing学习分享平台] 简历通过通知", body, app.Cfg.App.StaticDir+"/images/qr-code.png")
}

// SendNewResumeNotify 新简历投递提醒（发给开启了提醒的管理员）
func SendNewResumeNotify(to, realName, email string) error {
	link := app.Cfg.App.BaseURL + "/admin/resumes"
	body := fmt.Sprintf(`
<div>
    <p style="text-indent:2em;">收到一份新的简历投递，等待审核：</p>
    <p style="text-indent:2em;">姓名：<strong>%s</strong>（%s）</p>
    <p style="text-indent:2em;"><a href="%s">点击前往后台审核</a></p>
</div>`, realName, email, link)
	return send([]string{to}, "[AcKing学习分享平台] 新简历待审核", body)
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
