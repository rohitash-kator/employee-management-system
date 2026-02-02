package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/rohitashk/golang-rest-api/internal/adapters/mongodb"
	"github.com/rohitashk/golang-rest-api/internal/config"
	"github.com/rohitashk/golang-rest-api/internal/delivery/httpapi"
	"github.com/rohitashk/golang-rest-api/internal/observability"
	employeeUC "github.com/rohitashk/golang-rest-api/internal/usecase/employee"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	fmt.Println("cfg", cfg)
	if err != nil {
		panic(err)
	}

	logger := observability.NewLogger(cfg.AppEnv)
	slog.SetDefault(logger)

	if cfg.AppEnv != "prod" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	mongoClient, err := mongodb.Connect(ctx, cfg.MongoURI, cfg.MongoConnectTimeout)
	if err != nil {
		logger.Error("mongo connect failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
	}()

	db := mongoClient.Database(cfg.MongoDB)
	employeeRepo := mongodb.NewEmployeeRepository(db)
	if err := employeeRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("mongo indexes failed", "err", err)
		os.Exit(1)
	}

	employeeSvc := employeeUC.NewService(employeeRepo)

	router := httpapi.NewRouter(httpapi.RouterDeps{
		Logger:         logger,
		RequestTimeout: cfg.RequestTimeout,
		EmployeeSvc:    employeeSvc,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("http server starting", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("http server shutting down")
	_ = srv.Shutdown(ctxShutdown)
}
