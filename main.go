package main

import (
	"context"
	"log"

	"job-tracker/internal/config"
	"job-tracker/internal/database"
	"job-tracker/internal/handler"
	"job-tracker/internal/openrouter"
	"job-tracker/internal/repository"
	"job-tracker/internal/server"
	"job-tracker/internal/service"
	"job-tracker/internal/worker"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	appRepo := repository.NewApplicationRepository(db)
	historyRepo := repository.NewStatusHistoryRepository(db)
	jobRepo := repository.NewJobRepository(db)
	resumeRepo := repository.NewResumeRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	notionIntegrationRepo := repository.NewNotionIntegrationRepository(db)
	rajRepo := repository.NewResumeAnalyzerJobRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	appSvc := service.NewApplicationService(appRepo, historyRepo)
	jobSvc := service.NewJobService(jobRepo)
	resumeSvc := service.NewResumeService(resumeRepo)
	notionSvc := service.NewNotionService(cfg, notionIntegrationRepo, appRepo, jobRepo, statusRepo, historyRepo)
	orClient := openrouter.New(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)

	jobCh := make(chan int, 30)
	rajSvc := service.NewResumeAnalyzerJobService(rajRepo, resumeRepo, jobCh)

	workerPool := worker.NewPool(context.Background(), jobCh, worker.ProcessorDeps{
		ResumeRepo: resumeRepo,
		JobRepo:    rajRepo,
		ORClient:   orClient,
	})
	workerPool.Start()

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	appHandler := handler.NewApplicationHandler(appSvc)
	jobHandler := handler.NewJobHandler(jobSvc)
	resumeHandler := handler.NewResumeHandler(resumeSvc)
	statusHandler := handler.NewStatusHandler(statusRepo)
	notionHandler := handler.NewNotionHandler(notionSvc)
	rajHandler := handler.NewResumeAnalyzerJobHandler(rajSvc)

	// Server
	r := server.NewEngine()
	server.RegisterRoutes(r, authHandler, appHandler, jobHandler, resumeHandler, statusHandler, notionHandler, rajHandler, cfg.JWTSecret)

	log.Printf("server running on port %s", cfg.ServerPort)
	if err := server.Run(r, cfg.ServerPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
