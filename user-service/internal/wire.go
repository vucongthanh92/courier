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

	// internal usecases
	auditLogUc "github.com/vucongthanh92/courier/user-service/internal/usecase/audit_log"
	authUc "github.com/vucongthanh92/courier/user-service/internal/usecase/authen"
	credentialUc "github.com/vucongthanh92/courier/user-service/internal/usecase/credential"
	identityUc "github.com/vucongthanh92/courier/user-service/internal/usecase/identity"
	outboxUc "github.com/vucongthanh92/courier/user-service/internal/usecase/outbox"
	tokenUc "github.com/vucongthanh92/courier/user-service/internal/usecase/token"

	// internal repositories
	auditLogRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/audit_log"
	authCredentialRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/auth_credential"
	emailVerificationRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/email_verification"
	identityRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/identity"
	jwkRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/jwk"
	outboxRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/outbox"
	refreshTokenRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/refresh_token"
	userRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/user"

	// external repositories
	emailSender "github.com/vucongthanh92/courier/user-service/internal/repository/external/email_sender"
	jwtSigner "github.com/vucongthanh92/courier/user-service/internal/repository/external/jwt"
	"github.com/vucongthanh92/courier/user-service/internal/repository/external/oauth"
	oauthRepo "github.com/vucongthanh92/courier/user-service/internal/repository/external/oauth"
	redisRepo "github.com/vucongthanh92/courier/user-service/internal/repository/external/redis"

	// shared interfaces
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"

	// helper
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	grpcserver "github.com/vucongthanh92/courier/user-service/internal/api/grpc"
	v1 "github.com/vucongthanh92/courier/user-service/internal/api/http/v1"

	// third-party
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
	v1.InitCredentialHandler,
)

var serviceSet = wire.NewSet(
	cronjob.NewCronJobService,
	auditLogUc.InitAuditLogUsecase,
	authUc.InitAuthUseCase,
	identityUc.InitIdentityUseCase,
	outboxUc.InitOutboxUsecase,
	tokenUc.InitTokenUseCase,
	credentialUc.InitCredentialUseCase,
)

var repoSet = wire.NewSet(

	// internal repo
	transaction.InitManagerTxn,
	userRepo.InitUserCmdRepository,
	userRepo.InitUserQueryRepository,
	identityRepo.InitIdentityCmdRepository,
	identityRepo.InitIdentityQueryRepository,
	auditLogRepo.InitAuditLogCmdRepository,
	authCredentialRepo.InitAuthCredentialCmdRepository,
	authCredentialRepo.InitAuthCredentialQueryRepository,
	emailVerificationRepo.InitEmailVerificationCmdRepository,
	emailVerificationRepo.InitEmailVerificationQueryRepository,
	outboxRepo.InitOutboxCmdRepository,
	outboxRepo.InitOutboxQueryRepository,
	refreshTokenRepo.InitRefreshTokenCmdRepository,
	refreshTokenRepo.InitRefreshTokenQueryRepository,
	jwkRepo.InitJWKQueryRepository,

	// external repo
	emailSender.InitSMTPSender,
	redisRepo.InitRedisDenylist,

	// shared dependencies
	provideEmailConfig,
	provideLogger,
	provideJWTSigner,
	provideGoogleClient,
	provideGitHubClient,
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
func provideJWTSigner(jwkRepo interfaces.JWKQueryRepoI, log logger.Logger) interfaces.JWTSignerI {
	jwk, err := jwkRepo.GetActiveKey(context.Background())
	if err != nil {
		log.Fatal("load active jwk failed", zap.Any("error", err))
	}
	s, err2 := jwtSigner.InitJWTSigner(jwk, log)
	if err2 != nil {
		log.Fatal("init jwt signer failed", zap.Error(err2))
	}
	return s
}

// provideGoogleClient initializes the Google OAuth client with credentials from config.
func provideGoogleClient(cfg *config.AppConfig) interfaces.GoogleProviderClient {
	return oauth.NewGoogleClient(
		cfg.OAuth.Google.ClientID,
		cfg.OAuth.Google.ClientID,
		cfg.OAuth.Google.ClientSecret,
		cfg.OAuth.Google.RedirectURI,
	)
}

// provideGitHubClient initializes the GitHub OAuth client with API base URL from config.
func provideGitHubClient(cfg *config.AppConfig) interfaces.GithubProviderClient {
	return oauthRepo.NewGitHubClient(
		cfg.OAuth.Github.APIBase,
		cfg.OAuth.Github.ClientID,
		cfg.OAuth.Github.ClientSecret,
	)
}
