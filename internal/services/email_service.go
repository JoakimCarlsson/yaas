package services

import (
	"net/smtp"
	"strings"

	"github.com/joakimcarlsson/yaas/internal/config"
)

type EmailService interface {
	SendVerificationEmail(to, token string) error
}

type emailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{config: cfg}
}

func (s *emailService) SendVerificationEmail(to, token string) error {
	smtpHost := s.config.SmtpHost
	smtpPort := s.config.SmtpPort

	auth := smtp.PlainAuth("", s.config.SmtpUser, s.config.SmtpPassword, smtpHost)

	from := "noreply@jdaddy.net"
	subject := "Verify your email address"
	body := "Click the following link to verify your email address: http://localhost:8080/verify?token=" + token
	message := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n")

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, strings.Split(to, ","), message)
	return err
}
