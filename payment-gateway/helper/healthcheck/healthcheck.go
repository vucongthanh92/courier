package healthcheck

import (
	"context"
	"net/http"

	"github.com/vucongthanh92/courier/payment-gateway/config"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

func RunHealthCheck(
	_ context.Context,
	cfg *config.AppConfig,
	_ database.GormReadDb,
	_ database.GormWriteDb,
) func() {
	return func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/live", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		logger.Info("Healthcheck server listening on port", zap.String("Port", cfg.Healthcheck.Port))
		if err := http.ListenAndServe(cfg.Healthcheck.Port, mux); err != nil {
			logger.Warn("Healthcheck server", zap.Error(err))
		}
	}
}
