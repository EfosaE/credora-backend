// email/domain.go
package email

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
)

type EmailSender interface {
	SendEmail(ctx context.Context, req SendEmailRequest)  error
}

// email/domain.go continued
type SendEmailRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Html     string `json:"html,omitempty"`     // raw HTML if already built
	Template string `json:"template,omitempty"` // template name, if using embedded templates
	Data     any    `json:"data,omitempty"`     // template data context
	ReplyTo  string `json:"reply_to,omitempty"`
}

//go:embed templates/account_email.html
var accountTemplate string

//go:embed templates/welcome_email.html
var welcomeTemplate string

// //go:embed templates/password_reset.html
// var passwordResetTemplate string

var templates = map[string]string{
	"account_email": accountTemplate,
	"welcome_email": welcomeTemplate,
	// "password_reset": passwordResetTemplate,
}

func GetTemplateContent(name string) (string, error) {
	tmpl, ok := templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}
	return tmpl, nil
}

func RenderTemplate(name string, data any) (string, error) {
	raw, err := GetTemplateContent(name)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(name).Parse(raw)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
