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
const defaultAPIBaseURL = "https://api.mailgun.net"

// SendEmail sends a transactional email via the Mailgun HTTP API.
// Requires MAILGUN_API_KEY and MAILGUN_DOMAIN to be set.
//
// MAILGUN_API_BASE_URL optionally overrides the API host (defaults to
// Mailgun's own https://api.mailgun.net). This is for routing through a
// relay - e.g. a Cloudflare Worker - for hosts with no outbound IPv4
// connectivity, since Mailgun's API has no IPv6 address. When set,
// MAILGUN_RELAY_SECRET (if also set) is sent as an X-Relay-Secret header so
// the relay can reject requests that don't know the shared secret.
func SendEmail(toEmail string, subject string, htmlBody string, fromEmail ...string) {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")
	if apiKey == "" || domain == "" {
		logger.StdErr.Println("MAILGUN_API_KEY or MAILGUN_DOMAIN is not set, skipping email send")
		return
	}

	apiBaseURL := os.Getenv("MAILGUN_API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/%s/messages", apiBaseURL, domain), body)
	if err != nil {
		logger.StdErr.Println(err)
		return
	}
	req.SetBasicAuth("api", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if relaySecret := os.Getenv("MAILGUN_RELAY_SECRET"); relaySecret != "" {
		req.Header.Set("X-Relay-Secret", relaySecret)
	}

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
