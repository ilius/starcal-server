package scal

import (
	"os/exec"
	"strings"
)

func SendEmail(
	to string,
	subject string,
	isHtml bool,
	body string,
) error {
	cmd := exec.Command(
		"/usr/sbin/exim4",
		"-i",
		to,
	)
	var contentType string
	if isHtml {
		contentType = "text/html"
	} else {
		contentType = "text/plain"
	}
	cmd.Stdin = strings.NewReader(
		"From: StarCalendar <noreply@starcalendar.net>\n" +
			"Content-Type: " + contentType + "\n" +
			"To: " + to + "\n" +
			"Subject: " + subject + "\n" +
			body)
	err := cmd.Run()
	return err
}
