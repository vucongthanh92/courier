package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"gorm.io/gorm"
)

// Error carries the canonical Courier error through GORM's error-only
// transaction callback without losing its HTTP/domain semantics.
type Error struct {
	Builder *errHandler.ErrorBuilder
}

func (e *Error) Error() string {
	if e.Builder != nil && e.Builder.LogError != nil {
		return e.Builder.LogError.Error()
	}
	return "transaction aborted by domain error"
}

var runnerKey = struct{}{}

func RunnerFromCtx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if v := ctx.Value(runnerKey); v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			return tx
		}
	}
	return db
}

type Options struct {
	Isolation string        // e.g., "READ COMMITTED", "SERIALIZABLE" (Postgres)
	Timeout   time.Duration // optional context timeout
}

type ManagerTxn struct {
	db *gorm.DB
}

func InitManagerTxn(writeDb *database.GormWriteDb) *ManagerTxn {
	return &ManagerTxn{db: *writeDb}
}

func (m *ManagerTxn) Do(ctx context.Context, fn func(ctx context.Context) *errHandler.ErrorBuilder, opts ...Options) error {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	if v := ctx.Value(runnerKey); v != nil {
		tx, ok := v.(*gorm.DB)
		if !ok {
			return fmt.Errorf("transaction context value has unexpected type %T", v)
		}

		sp := fmt.Sprintf("sp_%d", time.Now().UnixNano())
		if err := tx.SavePoint(sp).Error; err != nil {
			return err
		}

		if commonErr := fn(ctx); commonErr != nil {
			_ = tx.RollbackTo(sp).Error
			return &Error{Builder: commonErr}
		}

		return nil
	}

	db := m.db
	if opt.Isolation != "" {
		db = db.Session(&gorm.Session{
			/* Postgres isolation via Set(tx_opts) */
		})
		// For postgres with gorm: use Exec to set isolation if needed
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, runnerKey, tx)
		commonErr := fn(txCtx)
		if commonErr != nil {
			return &Error{Builder: commonErr}
		}
		return nil
	})
}
