package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to,subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	from := os.Getenv("EMAIL_FROM")
	addr:=host+":"+port
	auth:=smtp.PlainAuth("",user,pass,host)
	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
			"\r\n" +
			body,
	)
	err:=smtp.SendMail(addr,auth,from,[]string{to},msg)
		if err != nil {
		return fmt.Errorf("send mail failed: %w", err)
	}

	return nil


}