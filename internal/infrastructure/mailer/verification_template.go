package mailer

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var verificationEmailTemplate = template.Must(template.ParseFS(templatesFS, "templates/verification_email.html.tmpl"))

type VerificationTemplateData struct {
	AppName   string
	VerifyURL string
}

func RenderVerificationEmail(data VerificationTemplateData) (string, error) {
	var rendered bytes.Buffer
	if err := verificationEmailTemplate.Execute(&rendered, data); err != nil {
		return "", err
	}

	return rendered.String(), nil
}
