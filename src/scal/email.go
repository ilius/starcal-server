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

func (in *SendEmailInput) GetContentType() string {
	if in.IsHtml {
		return "text/html"
	}
	return "text/plain"
}

func SendEmail(in *SendEmailInput) error {
	cmd := exec.Command(
		"/usr/sbin/exim4",
		"-i",
		in.To,
	)
	cmd.Stdin = strings.NewReader(
		"From: StarCalendar <noreply@starcalendar.net>\n" +
			"Content-Type: " + in.GetContentType() + "\n" +
			"To: " + in.To + "\n" +
			"Subject: " + in.Subject + "\n" +
			in.Body)
	err := cmd.Run()
	return err
}
