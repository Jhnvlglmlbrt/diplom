package notify

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
)

type Notifier interface {
	NotifyStatus(context.Context, data.TrackingAndAccount) error
	NotifyExpires(context.Context, data.TrackingAndAccount) error
}

type EmailNotifier struct {
	from       string
	password   string
	smtpServer string
	smtpPort   string
}

func NewEmailNotifier(from, password, smtpServer, smtpPort string) *EmailNotifier {
	return &EmailNotifier{
		from:       from,
		password:   password,
		smtpServer: smtpServer,
		smtpPort:   smtpPort,
	}
}

// если домен не здоров (т.е. истекает, но ещё не подошёл ко времени, указанному в профиле)
func (n *EmailNotifier) NotifyStatus(ctx context.Context, tracking data.TrackingAndAccount) error {
	// logger.Log("time", tracking.Expires)
	daysUntilExpiration := int(time.Until(tracking.Expires).Hours() / 24)
	body := fmt.Sprintf("Ваш SSL-сертификат домена - %s. Дней до истечения срока действия SSL сертификата: %d.", tracking.DomainName, daysUntilExpiration)
	logger.Log("EMAIL SENT TO =>", tracking.Account.NotifyDefaultEmail, "Домен - ", tracking.DomainName)
	return n.sendEmail(tracking.Account.NotifyDefaultEmail, "Статус SSL-сертификата", body)
}

// метод для автоматической отправки email о домене (если оставшееся время до истечения < или = времени, указанному в аккаунте)
func (n *EmailNotifier) NotifyExpires(ctx context.Context, tracking data.TrackingAndAccount) error {
	daysUntilExpiration := int(time.Until(tracking.Expires).Hours() / 24)
	body := fmt.Sprintf("Ваш SSL-сертификат домена - %s скоро истечёт. Дней до истечения срока действия: %d.", tracking.DomainName, daysUntilExpiration)
	logger.Log("EMAIL SENT TO =>", tracking.Account.NotifyDefaultEmail, "Домен - ", tracking.DomainName)
	return n.sendEmail(tracking.Account.NotifyDefaultEmail, "Истечение срока действия SSL-сертификата", body)
}

// метод для отправки email о состоянии домена
func (n *EmailNotifier) JustNotify(ctx context.Context, email, subject, body string) error {
	return n.sendEmail(email, subject, body)
}

// Метод для отправки сообщения
func (n *EmailNotifier) sendEmail(to, subject, body string) error {
	htmlBody := fmt.Sprintf(`
        <html>
        <head>
            <style>
                p {
                    font-size: 26px;
                }
            </style>
        </head>
        <body>
            <p>%s</p>
        </body>
        </html>
    `, body)

	//  заголовок сообщения
	header := map[string]string{
		"From":         fmt.Sprintf("SafeCert Monitor <%s>", n.from),
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"utf-8\"",
	}

	//  тело сообщения
	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + htmlBody

	// Аутентификация и отправка почты
	auth := smtp.PlainAuth("", n.from, n.password, n.smtpServer)
	addr := fmt.Sprintf("%s:%s", n.smtpServer, n.smtpPort)
	if err := smtp.SendMail(addr, auth, n.from, []string{to}, []byte(message)); err != nil {
		return err
	}
	return nil
}
