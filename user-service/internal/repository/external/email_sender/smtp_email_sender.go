package emailsender

import (
	"context"
	"fmt"
	"net/smtp"
	"net/url"

	"github.com/vucongthanh92/courier/user-service/config"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

type smtpSender struct {
	cfg    *config.EmailConfig
	logger logger.Logger
}

func InitSMTPSender(cfg *config.EmailConfig, logger logger.Logger) interfaces.EmailSenderI {
	return &smtpSender{cfg: cfg, logger: logger}
}

// SendVerificationEmail implements interfaces.EmailSenderI
func (s *smtpSender) SendVerificationEmail(ctx context.Context, toEmail, token string) *errHandler.ErrorBuilder {
	if s.cfg == nil || !s.cfg.Enabled {
		s.logger.Info("Skip send email (disabled)", zap.String("to", toEmail))
		return nil
	}

	link := fmt.Sprintf("%s?email=%s&token=%s",
		s.cfg.VerifyURL,
		url.QueryEscape(toEmail),
		url.QueryEscape(token),
	)

	subject := "Verify your Courier account"
	body := fmt.Sprintf("Hi,\n\nUse this code: %s\nOr click: %s\n\nThanks!", token, link)
	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTP.Host, s.cfg.SMTP.Port)
	auth := smtp.PlainAuth("", s.cfg.SMTP.Username, s.cfg.SMTP.Password, s.cfg.SMTP.Host)

	if err := smtp.SendMail(addr, auth, s.cfg.From, []string{toEmail}, msg); err != nil {
		return errHandler.InitErrorBuilder(ctx).SetLogError(err).SetStatus(500)
	}

	return nil
}
