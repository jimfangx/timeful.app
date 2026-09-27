package mailgun

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"schej.it/server/logger"
)

// sanitizeHeaderValue strips CR/LF characters from user-controlled strings
// before they're used in an email subject line, to prevent header injection.
func sanitizeHeaderValue(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}

const wrapperTemplateSrc = `<!DOCTYPE html>
<html>
  <body style="margin:0;padding:32px 16px;background-color:#f2f2f2;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
    <div style="max-width:480px;margin:0 auto;background-color:#ffffff;border-radius:8px;padding:32px;">
      <div style="font-size:20px;font-weight:600;color:#1C7D45;margin-bottom:24px;">timeful</div>
      {{.Body}}
    </div>
  </body>
</html>`

var wrapperTemplate = template.Must(template.New("wrapper").Parse(wrapperTemplateSrc))

const buttonStyle = "display:inline-block;background-color:#00994C;color:#ffffff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;"
const mutedTextStyle = "font-size:14px;color:#6B6B6B;"
const bodyTextStyle = "font-size:16px;color:#111111;line-height:1.5;"

// renderBody executes bodyTemplateSrc with data (escaping any user-controlled
// values in the process), then wraps the result in the shared email layout.
func renderBody(bodyTemplateSrc string, data any) (string, error) {
	bodyTemplate, err := template.New("body").Parse(bodyTemplateSrc)
	if err != nil {
		return "", err
	}

	var bodyBuf bytes.Buffer
	if err := bodyTemplate.Execute(&bodyBuf, data); err != nil {
		return "", err
	}

	// bodyBuf is already safely escaped by bodyTemplate above, so it's passed
	// through as template.HTML to avoid the wrapper template escaping it again.
	var out bytes.Buffer
	if err := wrapperTemplate.Execute(&out, struct{ Body template.HTML }{template.HTML(bodyBuf.String())}); err != nil {
		return "", err
	}

	return out.String(), nil
}

// send renders bodyTemplateSrc/data, then sends it as an email with subject
// to toEmail, logging (rather than returning) any rendering error, matching
// the fire-and-forget style of the rest of this package.
func send(toEmail string, subject string, bodyTemplateSrc string, data any) {
	html, err := renderBody(bodyTemplateSrc, data)
	if err != nil {
		logger.StdErr.Println(err)
		return
	}
	SendEmail(toEmail, subject, html)
}

// SendOtpEmail sends the sign-in verification code email.
func SendOtpEmail(toEmail, code string) {
	send(toEmail, "Your Timeful sign-in code", `
		<p style="`+bodyTextStyle+`">Your sign-in code is:</p>
		<p style="font-size:32px;font-weight:700;letter-spacing:6px;color:#1C7D45;margin:16px 0;">{{.Code}}</p>
		<p style="`+mutedTextStyle+`">This code expires in 10 minutes. If you didn't request this, you can safely ignore this email.</p>
	`, struct{ Code string }{code})
}

// SendGroupInviteEmail sends the "you've been invited to a group" email.
func SendGroupInviteEmail(toEmail, ownerName, groupName, groupUrl string) {
	subject := fmt.Sprintf(`%s invited you to "%s" on Timeful`, sanitizeHeaderValue(ownerName), sanitizeHeaderValue(groupName))
	send(toEmail, subject, `
		<p style="`+bodyTextStyle+`"><strong>{{.OwnerName}}</strong> invited you to join <strong>{{.GroupName}}</strong> on Timeful.</p>
		<p style="margin:24px 0;"><a href="{{.GroupUrl}}" style="`+buttonStyle+`">View group</a></p>
	`, struct {
		OwnerName string
		GroupName string
		GroupUrl  string
	}{ownerName, groupName, groupUrl})
}

// SendAddedAttendeeEmail sends the "new people were added to a group you're
// in" email.
func SendAddedAttendeeEmail(toEmail, ownerName, groupName, groupUrl string, emails []string) {
	subject := fmt.Sprintf(`New people were added to "%s" on Timeful`, sanitizeHeaderValue(groupName))
	send(toEmail, subject, `
		<p style="`+bodyTextStyle+`"><strong>{{.OwnerName}}</strong> added the following people to <strong>{{.GroupName}}</strong>:</p>
		<ul style="`+bodyTextStyle+`">
			{{range .Emails}}<li>{{.}}</li>{{end}}
		</ul>
		<p style="margin:24px 0;"><a href="{{.GroupUrl}}" style="`+buttonStyle+`">View group</a></p>
	`, struct {
		OwnerName string
		GroupName string
		GroupUrl  string
		Emails    []string
	}{ownerName, groupName, groupUrl, emails})
}

// SendSomeoneRespondedEmail sends the "X responded to your event/group"
// email.
func SendSomeoneRespondedEmail(toEmail, itemName, ownerName, respondentName, itemUrl string, isGroup bool) {
	itemNoun := "event"
	if isGroup {
		itemNoun = "group"
	}
	subject := fmt.Sprintf(`%s responded to "%s"`, sanitizeHeaderValue(respondentName), sanitizeHeaderValue(itemName))
	send(toEmail, subject, `
		<p style="`+bodyTextStyle+`">Hi {{.OwnerName}},</p>
		<p style="`+bodyTextStyle+`"><strong>{{.RespondentName}}</strong> just responded to your {{.ItemNoun}} <strong>{{.ItemName}}</strong>.</p>
		<p style="margin:24px 0;"><a href="{{.ItemUrl}}" style="`+buttonStyle+`">View responses</a></p>
	`, struct {
		OwnerName      string
		RespondentName string
		ItemNoun       string
		ItemName       string
		ItemUrl        string
	}{ownerName, respondentName, itemNoun, itemName, itemUrl})
}

// SendXResponsesEmail sends the "your event reached N responses" email.
func SendXResponsesEmail(toEmail, eventName, ownerName, eventUrl string, numResponses int) {
	subject := fmt.Sprintf(`"%s" has reached %d responses`, sanitizeHeaderValue(eventName), numResponses)
	send(toEmail, subject, `
		<p style="`+bodyTextStyle+`">Hi {{.OwnerName}},</p>
		<p style="`+bodyTextStyle+`">Your event <strong>{{.EventName}}</strong> has reached <strong>{{.NumResponses}}</strong> responses!</p>
		<p style="margin:24px 0;"><a href="{{.EventUrl}}" style="`+buttonStyle+`">View responses</a></p>
	`, struct {
		OwnerName    string
		EventName    string
		EventUrl     string
		NumResponses int
	}{ownerName, eventName, eventUrl, numResponses})
}

// SendEveryoneRespondedEmail sends the "everyone has responded to your
// event" email.
func SendEveryoneRespondedEmail(toEmail, eventName, eventUrl string) {
	subject := fmt.Sprintf(`Everyone has responded to "%s"`, sanitizeHeaderValue(eventName))
	send(toEmail, subject, `
		<p style="`+bodyTextStyle+`">Great news! Everyone has responded to your event <strong>{{.EventName}}</strong>.</p>
		<p style="margin:24px 0;"><a href="{{.EventUrl}}" style="`+buttonStyle+`">View responses</a></p>
	`, struct {
		EventName string
		EventUrl  string
	}{eventName, eventUrl})
}
