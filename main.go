package main

import (
	"log"

	"job-tracker/internal/config"
	"job-tracker/internal/database"
	"job-tracker/internal/handler"
	"job-tracker/internal/repository"
	"job-tracker/internal/server"
	"job-tracker/internal/service"

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

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	appSvc := service.NewApplicationService(appRepo, historyRepo)
	jobSvc := service.NewJobService(jobRepo)
	resumeSvc := service.NewResumeService(resumeRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	appHandler := handler.NewApplicationHandler(appSvc)
	jobHandler := handler.NewJobHandler(jobSvc)
	resumeHandler := handler.NewResumeHandler(resumeSvc)
	statusHandler := handler.NewStatusHandler(statusRepo)

	// Server
	r := server.NewEngine()
	server.RegisterRoutes(r, authHandler, appHandler, jobHandler, resumeHandler, statusHandler, cfg.JWTSecret)

	log.Printf("server running on port %s", cfg.ServerPort)
	if err := server.Run(r, cfg.ServerPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
