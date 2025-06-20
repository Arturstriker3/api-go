package email

import (
	"fmt"
	"github.com/Arturstriker3/api-go/config"
	"gopkg.in/gomail.v2"
)

type EmailData struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

type Service struct {
	config *config.Config
	dialer *gomail.Dialer
}

func NewEmailService(cfg *config.Config) *Service {
	dialer := gomail.NewDialer(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.User,
		cfg.SMTP.Password,
	)

	return &Service{
		config: cfg,
		dialer: dialer,
	}
}

func (s *Service) SendEmail(data *EmailData) error {
	if len(data.To) == 0 {
		return fmt.Errorf("recipient list is empty")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.SMTP.From)
	m.SetHeader("To", data.To...)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", data.Body)

	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
} 