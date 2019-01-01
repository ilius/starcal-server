package scal

import (
	"os/exec"
	"strings"
)

type SendEmailInput struct {
	To      string
	Subject string
	IsHtml  bool
	Body    string
}

func SendEmail(in *SendEmailInput) error {
	cmd := exec.Command(
		"/usr/sbin/exim4",
		"-i",
		in.To,
	)
	var contentType string
	if in.IsHtml {
		contentType = "text/html"
	} else {
		contentType = "text/plain"
	}
	cmd.Stdin = strings.NewReader(
		"From: StarCalendar <noreply@starcalendar.net>\n" +
			"Content-Type: " + contentType + "\n" +
			"To: " + in.To + "\n" +
			"Subject: " + in.Subject + "\n" +
			in.Body)
	err := cmd.Run()
	return err
}
