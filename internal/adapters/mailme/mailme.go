package mailme

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// Mailer sends templated emails using gomail.
type Mailer struct {
	From   string
	Host   string
	Port   int
	User   string
	Pass   string
	Logger *logrus.Logger
}

// NewMailer creates a new instance of the Mailer.
func NewMailer(host string, port int, user, pass, from string, logger *logrus.Logger) *Mailer {
	return &Mailer{
		Host:   host,
		Port:   port,
		User:   user,
		Pass:   pass,
		From:   from,
		Logger: logger,
	}
}

// Mail parses and sends a multipart email with both HTML and plain text parts.
func (m *Mailer) Mail(to, subjectTemplate, htmlTemplate, textTemplate string, data map[string]interface{}) error {
	log := m.Logger.WithFields(logrus.Fields{"recipient": to})
	log.Info("preparing to send email")

	// 1. Render Subject
	subject, err := m.render("subject", subjectTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render subject template: %w", err)
	}

	// 2. Render HTML Body
	htmlBody, err := m.render("html", htmlTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render html template: %w", err)
	}

	// 3. Create Message
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", htmlBody)

	// 4. Render and add Plain Text Body (optional)
	if textTemplate != "" {
		textBody, err := m.render("text", textTemplate, data)
		if err != nil {
			// Log a warning but don't fail the entire send, as HTML is the primary part.
			log.WithError(err).Warn("failed to render plain text template, sending HTML only")
		} else {
			msg.AddAlternative("text/plain", textBody)
		}
	}

	// 5. Dial and Send
	dialer := gomail.NewDialer(m.Host, m.Port, m.User, m.Pass)
	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	log.Info("email sent successfully")
	return nil
}

// render is a helper function to parse and execute a template string.
func (m *Mailer) render(templateName, templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New(templateName).Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
