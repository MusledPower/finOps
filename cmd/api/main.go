package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "task-decomposer-backend/docs"
	"task-decomposer-backend/internal/config"
	"task-decomposer-backend/internal/database"
	"task-decomposer-backend/internal/handler"
	appmiddleware "task-decomposer-backend/internal/middleware"
	"task-decomposer-backend/internal/provider"
	"task-decomposer-backend/internal/repository"
	"task-decomposer-backend/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Task Decomposer API
// @version         1.0
// @description     API for AI-powered task decomposition and plan generation.
// @BasePath        /
// @schemes         http https
//
// @tag.name        System
// @tag.description System health and capabilities
//
// @tag.name        Plans
// @tag.description Plan generation and retrieval
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token with the Bearer prefix. Example: Bearer your-token

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.New(pool)
	llm := provider.NewDeepSeek(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel, cfg.LLMTimeout)
	plans := service.NewPlanService(repo, llm, cfg.MaxRequestBytes)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(appmiddleware.RequestID)

	r.Get("/health", handler.Health)
	r.Get(
		"/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/docs/doc.json"),
		),
	)

	r.Group(
		func(api chi.Router) {
			api.Use(appmiddleware.BearerAuth(cfg.BearerToken))
			api.Get("/v1/capabilities", handler.Capabilities)
			api.Post("/v1/plans", plans.Create)
			api.Get("/v1/plans/{id}", plans.Get)
		},
	)

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("server started", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-stop.Done()
	shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	_ = srv.Shutdown(shutdown)
}
