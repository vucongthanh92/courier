package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
)

type EmailSenderI interface {
	SendVerificationEmail(ctx context.Context, toEmail, token string) *errHandler.ErrorBuilder
}
