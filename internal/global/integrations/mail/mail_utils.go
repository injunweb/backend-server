package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/injunweb/backend-server/env"
)

var (
	host     string
	port     string
	username string
	password string
)

func init() {
	host = env.SMTP_HOST
	port = env.SMTP_PORT
	username = env.SMTP_USERNAME
	password = env.SMTP_PASSWORD

	if host == "" || port == "" || username == "" || password == "" {
		log.Fatal("SMTP environment variables not set")
	}
}

func SendEmail(to []string, subject, body string) error {
	auth := smtp.PlainAuth("", username, password, host)

	header := make(map[string]string)
	header["From"] = username
	header["To"] = strings.Join(to, ",")
	header["Subject"] = subject
	header["MIME-version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", host+":"+port, tlsconfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(username); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
