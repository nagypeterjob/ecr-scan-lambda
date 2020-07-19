package exporters

import (
	"strings"

	"github.com/mailgun/mailgun-go"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
)

// MailgunExporter lets you send reports via email
type MailgunExporter struct {
	client     *mailgun.MailgunImpl
	from       string
	name       string
	recipients string
}

// NewMailgunExporter .
func NewMailgunExporter(name string, recipients string, from string, apiKey string) *MailgunExporter {
	domain := domain(from)

	return &MailgunExporter{
		client:     mailgun.NewMailgun(domain, apiKey),
		from:       from,
		recipients: recipients,
		name:       name,
	}
}

// Name .
func (m MailgunExporter) Name() string {
	return m.name
}

// Format clousure formats scan results and returns a function that sends report on invocation
func (m MailgunExporter) Format(filtered []*api.RepositoryInfo, failed []*api.RepositoryInfo) (func() error, error) {

	filteredMsg, err := format(filtered)
	if err != nil {
		return nil, err
	}

	failedMsg, err := formatFailed(failed)
	if err != nil {
		return nil, err
	}

	msg := m.client.NewMessage(
		m.from,
		"Daily ECR scan report",
		filteredMsg+failedMsg,
	)

	recipientList := strings.Split(m.recipients, ",")

	for _, user := range recipientList {
		if err := msg.AddRecipient(user); err != nil {
			return nil, err
		}
	}

	return func() error {

		if _, _, err = m.client.Send(msg); err != nil {
			return err
		}

		return nil
	}, nil
}

// extracts domain from email address
func domain(email string) string {
	if email == "" {
		return ""
	}

	if strings.Count(email, "@") > 1 {
		return ""
	}

	atPosition := strings.Index(email, "@")
	return email[atPosition+1:]
}
