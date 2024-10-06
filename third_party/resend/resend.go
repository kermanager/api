package resend

import (
	"fmt"

	resendGo "github.com/resend/resend-go/v2"
)

type ResendService interface {
	SendInvitationEmail(to string, email string, password string) (*resendGo.SendEmailResponse, error)
}

type Resend struct {
	Client    *resendGo.Client
	FromEmail string
}

func NewResendService(apiKey string, fromEmail string) *Resend {
	return &Resend{
		Client:    resendGo.NewClient(apiKey),
		FromEmail: fromEmail,
	}
}

func (t *Resend) sendEmail(to []string, subject string, content string) (*resendGo.SendEmailResponse, error) {
	params := &resendGo.SendEmailRequest{
		From:    t.FromEmail,
		To:      to,
		Subject: subject,
		Html:    content,
	}

	return t.Client.Emails.Send(params)
}

func (t *Resend) SendInvitationEmail(to string, email string, password string) (*resendGo.SendEmailResponse, error) {
	content := fmt.Sprintf(`
    <p>You have been invited to join our platform. Your credentials are as follows :</p>
    <p>Email : %s</p>
    <p>Password : %s</p>
  `, email, password)

	return t.sendEmail([]string{to}, "Invitation to join our platform", content)
}
