//go:build wireinject
// +build wireinject

package internal

import (
	"context"

	"github.com/google/wire"
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/courier/user-service/internal/api"
	"github.com/vucongthanh92/courier/user-service/internal/api/cron"
	"github.com/vucongthanh92/courier/user-service/internal/api/http"
	"github.com/vucongthanh92/courier/user-service/internal/usecase/cronjob"
	"github.com/vucongthanh92/courier/user-service/redis"

	auditLogUc "github.com/vucongthanh92/courier/user-service/internal/usecase/audit_log"
	authUc "github.com/vucongthanh92/courier/user-service/internal/usecase/auth"
	identityUc "github.com/vucongthanh92/courier/user-service/internal/usecase/identity"
	outboxUc "github.com/vucongthanh92/courier/user-service/internal/usecase/outbox"

	auditLogRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/audit_log"
	authCredWriteRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/auth_credential"
	emailVerificationRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/email_verification"
	identityRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/identity"
	outboxRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/outbox"
	userRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/user"

	EmailSender "github.com/vucongthanh92/courier/user-service/internal/repository/external/email_sender"
	jwtSigner "github.com/vucongthanh92/courier/user-service/internal/repository/external/jwt"

	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	grpcserver "github.com/vucongthanh92/courier/user-service/internal/api/grpc"
	v1 "github.com/vucongthanh92/courier/user-service/internal/api/http/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vucongthanh92/courier/user-service/internal/worker"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

var workerSet = wire.NewSet(
	worker.InitOutboxWorker,
	newPgxPool,
)

var container = wire.NewSet(
	api.NewApiContainer,
)

var apiSet = wire.NewSet(
	cron.NewServer,
	grpcserver.NewServer,
	http.NewServer,
)

var handlerSet = wire.NewSet(
	v1.InitIdentityHandler,
	v1.InitAuthHandler,
)

var serviceSet = wire.NewSet(
	cronjob.NewCronJobService,
	auditLogUc.InitAuditLogUsecase,
	authUc.InitAuthUsecase,
	identityUc.InitIdentityService,
	outboxUc.InitOutboxUsecase,
)

var repoSet = wire.NewSet(
	// internal repo
	transaction.InitManagerTxn,
	userRepo.InitUserCmdRepository,
	userRepo.InitUserQueryRepository,
	identityRepo.InitIdentityCmdRepository,
	identityRepo.InitIdentityQueryRepository,
	auditLogRepo.InitAuditLogCmdRepository,
	authCredWriteRepo.InitAuthCredentialCmdRepository,
	emailVerificationRepo.InitEmailVerificationCmdRepository,
	emailVerificationRepo.InitEmailVerificationQueryRepository,
	outboxRepo.InitOutboxCmdRepository,
	outboxRepo.InitOutboxQueryRepository,

	// external repo
	EmailSender.InitSMTPSender,
	jwtSigner.InitJWTSigner,

	// shared dependencies
	provideEmailConfig,
	provideLogger,
	provideJwtConfig,
)

func newPgxPool(cfg *config.AppConfig) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), cfg.Database.WriteDbCfg.ConnectionString)
	if err != nil {
		logger.Fatal("cannot init pgxpool", zap.Error(err))
	}
	return pool
}

func InitializeContainer(
	appCfg *config.AppConfig,
	readDb *database.GormReadDb,
	writeDb *database.GormWriteDb,
	redisClient redis.Client,
) *api.ApiContainer {
	wire.Build(repoSet, serviceSet, handlerSet, workerSet, apiSet, container)
	return &api.ApiContainer{}
}

// provideEmailConfig returns the nested email config for DI.
func provideEmailConfig(cfg *config.AppConfig) *config.EmailConfig {
	return cfg.Email
}

// provideLogger initializes a zap logger with configured level.
func provideLogger(cfg *config.AppConfig) logger.Logger {
	return logger.NewZapLogger(cfg.Logger.LogLevel)
}

// provideJWTSigner initializes the JWT signer with RSA keys from config.
func provideJwtConfig(cfg *config.AppConfig) *config.JWTConfig {
	return cfg.JWT
}
