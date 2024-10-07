package email

import (
	"fmt"
	"strconv"

	"github.com/injunweb/backend-server/internal/config"

	"gopkg.in/gomail.v2"
)

func SendApprovalEmail(toEmail, appName, dbPassword string) error {
	msg := fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: Application Approved\r\n\r\n"+
			"Database Type: mysql\r\n"+
			"Your application has been approved.\r\n\r\n"+
			"Database Host: %s\r\n"+
			"Database Port: %s\r\n"+
			"Database Name: %s\r\n"+
			"Database User: %s\r\n"+
			"Database Password: %s\r\n",
		toEmail, config.AppConfig.DBHost, config.AppConfig.DBPort, appName, appName, dbPassword,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", config.AppConfig.SMTPSenderEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Application Approved")
	m.SetBody("text/plain", msg)

	port, err := strconv.Atoi(config.AppConfig.SMTPPort)
	if err != nil {
		return fmt.Errorf("failed to convert SMTP port: %v", err)
	}

	d := gomail.NewDialer(config.AppConfig.SMTPHost, port, config.AppConfig.SMTPUser, config.AppConfig.SMTPPass)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
