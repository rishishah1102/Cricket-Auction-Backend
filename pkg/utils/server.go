package utils

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StartServer(ctx context.Context, router *gin.Engine, logger *zap.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	srv := &http.Server{
		Addr:    ":" + "5000",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start the server", zap.Error(err))
		}
	}()
	logger.Info("server started")

	<-quit
	logger.Info("shutting down the server...")

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdwon", zap.Error(err))
	}

	logger.Info("server exited gracefully")
}
