package resend

import (
	"fmt"

	resendGo "github.com/resend/resend-go/v2"
)

type ResendService interface {
	SendInvitationEmail(to string, email string, password string) (*resendGo.SendEmailResponse, error)
	SendKermesseInvitationEmail(to string, name string, description string) (*resendGo.SendEmailResponse, error)
	SendPaymentEmail(to string, credit int) (*resendGo.SendEmailResponse, error)
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
    <p>Votre parent vous a invité à rejoindre notre plateforme. Vos identifiants sont les suivants :</p>
    <p>Email : %s</p>
    <p>Mot de passe : %s</p>
  `, email, password)

	return t.sendEmail([]string{to}, "Invitation pour rejoindre notre plateforme", content)
}

func (t *Resend) SendKermesseInvitationEmail(to string, name string, description string) (*resendGo.SendEmailResponse, error) {
	content := fmt.Sprintf(`
		<p>Un organisateur vous a invité(e) à participer à la kermesse.</p>
    <p>Nom : %s</p>
    <p>Description : %s</p>
  `, name, description)

	return t.sendEmail([]string{to}, "Invitation pour participer à la kermesse", content)
}

func (t *Resend) SendPaymentEmail(to string, credit int) (*resendGo.SendEmailResponse, error) {
	content := fmt.Sprintf(`
		<p>Votre parent vous a envoyé(e) des jetons : %d</p>
  `, credit)

	return t.sendEmail([]string{to}, "Reception de jetons", content)
}
