package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-utils/logger"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"go.uber.org/zap"
)

type OutboxWorker struct {
	pgxPool           *pgxpool.Pool
	outboxQueryRepo   interfaces.OutboxQueryRepoI
	outboxCommandRepo interfaces.OutboxCommandRepoI
	auditLogService   interfaces.AuditLogServiceI
	emailSender       interfaces.EmailSenderI
	logger            logger.Logger
}

func InitOutboxWorker(
	pgxPool *pgxpool.Pool,
	outboxQueryRepo interfaces.OutboxQueryRepoI,
	outboxCommandRepo interfaces.OutboxCommandRepoI,
	auditLogService interfaces.AuditLogServiceI,
	emailSender interfaces.EmailSenderI,
	logger logger.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		pgxPool:           pgxPool,
		outboxQueryRepo:   outboxQueryRepo,
		outboxCommandRepo: outboxCommandRepo,
		auditLogService:   auditLogService,
		emailSender:       emailSender,
		logger:            logger,
	}
}

// Start starts the outbox worker with LISTEN/NOTIFY pattern
func (w *OutboxWorker) Start(ctx context.Context) error {
	ctx, span := tracing.StartSpanFromContext(ctx, "OutboxWorker.Start")
	defer span.End()

	w.logger.Info("Starting outbox worker with LISTEN/NOTIFY pattern")

	// Acquire connection from pool for LISTEN
	conn, err := w.pgxPool.Acquire(ctx)
	if err != nil {
		w.logger.Error("Failed to acquire connection from pool", zap.Error(err))
		return err
	}
	defer conn.Release()

	// Execute LISTEN command
	_, err = conn.Exec(ctx, "LISTEN outbox_events")
	if err != nil {
		w.logger.Error("Failed to execute LISTEN command", zap.Error(err))
		return err
	}

	w.logger.Info("Successfully subscribed to outbox_events channel")

	// Listen for notifications
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Outbox worker stopped")
			return ctx.Err()
		default:
			// Wait for notification with timeout
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				// Check if context is cancelled
				if ctx.Err() != nil {
					return ctx.Err()
				}
				w.logger.Error("Error waiting for notification", zap.Error(err))
				// Sleep before retry
				time.Sleep(5 * time.Second)
				continue
			}

			// Process notification
			w.logger.Info("Received notification", zap.String("payload", notification.Payload))
			if err := w.processNotification(ctx, notification.Payload); err != nil {
				w.logger.Error("Failed to process notification", zap.Error(err), zap.String("payload", notification.Payload))
			}
		}
	}
}

// processNotification processes a single notification
func (w *OutboxWorker) processNotification(ctx context.Context, payload string) error {
	ctx, span := tracing.StartSpanFromContext(ctx, "OutboxWorker.processNotification")
	defer span.End()

	// Parse outbox event ID from payload
	outboxID, err := strconv.ParseUint(payload, 10, 64)
	if err != nil {
		w.logger.Error("Failed to parse outbox ID", zap.Error(err), zap.String("payload", payload))
		return err
	}

	// Fetch outbox event from database
	outboxEvent, txnErr := w.outboxQueryRepo.GetOutboxByID(ctx, outboxID)
	if txnErr != nil {
		w.logger.Error("Failed to fetch outbox event", zap.Any("error", txnErr), zap.Uint64("outbox_id", outboxID))
		return fmt.Errorf("failed to fetch outbox event: %w", txnErr)
	}

	// Check if already published
	if outboxEvent.PublishedAt != nil {
		w.logger.Info("Outbox event already published, skipping", zap.Uint64("outbox_id", outboxID))
		return nil
	}

	// Process based on event type
	switch outboxEvent.EventType {
	case "USER_CREATED":
		if err := w.processUserCreatedEvent(ctx, outboxEvent); err != nil {
			return err
		}
	case "EMAIL_VERIFICATION_SEND":
		if err := w.processSendVerifyEmail(ctx, outboxEvent); err != nil {
			return err
		}
	default:
		w.logger.Warn("Unknown event type", zap.String("event_type", outboxEvent.EventType), zap.Uint64("outbox_id", outboxID))
		return nil
	}

	// Mark event as published
	now := time.Now()
	outboxEvent.PublishedAt = &now
	txnErr = w.outboxCommandRepo.UpdateOutboxPublished(ctx, outboxEvent)
	if txnErr != nil {
		w.logger.Error("Failed to mark outbox event as published", zap.Any("error", txnErr), zap.Uint64("outbox_id", outboxID))
		return fmt.Errorf("failed to mark outbox event as published: %w", txnErr)
	}

	w.logger.Info("Successfully processed outbox event", zap.Uint64("outbox_id", outboxID), zap.String("event_type", outboxEvent.EventType))
	return nil
}

// processUserCreatedEvent processes USER_CREATED event
func (w *OutboxWorker) processUserCreatedEvent(ctx context.Context, outboxEvent *entities.Outbox) error {
	ctx, span := tracing.StartSpanFromContext(ctx, "OutboxWorker.processUserCreatedEvent")
	defer span.End()

	// Deserialize user entity from payload
	var user entities.User
	if err := json.Unmarshal(outboxEvent.Payload, &user); err != nil {
		w.logger.Error("Failed to deserialize user entity", zap.Error(err), zap.Uint64("outbox_id", outboxEvent.ID))
		return err
	}

	// Log user created event to audit log
	// For now, we use "unknown" for IP and User-Agent since they are not in the payload
	// This can be enhanced in Phase 3
	txnErr := w.auditLogService.LogUserCreated(ctx, &user, "unknown", "unknown")
	if txnErr != nil {
		w.logger.Error("Failed to log user created event", zap.Any("error", txnErr), zap.Uint64("user_id", user.ID))
		return fmt.Errorf("failed to log user created event: %w", txnErr)
	}

	w.logger.Info("Successfully logged user created event", zap.Uint64("user_id", user.ID))
	return nil
}

func (w *OutboxWorker) processSendVerifyEmail(ctx context.Context, outboxEvent *entities.Outbox) error {
	ctx, span := tracing.StartSpanFromContext(ctx, "OutboxWorker.processSendVerifyEmail")
	defer span.End()

	var payload struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}

	//
	if err := json.Unmarshal(outboxEvent.Payload, &payload); err != nil {
		w.logger.Error("Failed to deserialize payload", zap.Error(err))
		return err
	}

	// Call email sender to send verification email
	if err := w.emailSender.SendVerificationEmail(ctx, payload.Email, payload.Token); err != nil {
		w.logger.Error("Send verification email failed", zap.Any("error", err), zap.String("email", payload.Email))
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}
