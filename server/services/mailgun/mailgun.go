package mailgun

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"schej.it/server/logger"
)

const DefaultFromEmail = "Timeful <noreply@timeful.app>"

// SendEmail sends a transactional email via the Mailgun HTTP API.
// Requires MAILGUN_API_KEY and MAILGUN_DOMAIN to be set.
func SendEmail(toEmail string, subject string, htmlBody string, fromEmail ...string) {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")
	if apiKey == "" || domain == "" {
		logger.StdErr.Println("MAILGUN_API_KEY or MAILGUN_DOMAIN is not set, skipping email send")
		return
	}

	from := DefaultFromEmail
	if len(fromEmail) > 0 && fromEmail[0] != "" {
		from = fromEmail[0]
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("from", from)
	writer.WriteField("to", toEmail)
	writer.WriteField("subject", subject)
	writer.WriteField("html", htmlBody)
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", domain), body)
	if err != nil {
		logger.StdErr.Println(err)
		return
	}
	req.SetBasicAuth("api", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.StdErr.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		logger.StdErr.Printf("mailgun: received status %d sending email to %s\n", resp.StatusCode, toEmail)
	}
}
