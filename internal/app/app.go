package app

import (
	"backend-trainee-assignment-2024/internal/config"
	"backend-trainee-assignment-2024/internal/logger"
	"backend-trainee-assignment-2024/internal/repo"
	"backend-trainee-assignment-2024/internal/services/auth"
	"backend-trainee-assignment-2024/internal/services/banner"
	"backend-trainee-assignment-2024/internal/storage/postgres"
	"backend-trainee-assignment-2024/internal/storage/rabbitmq"
	"backend-trainee-assignment-2024/internal/storage/redis"
	"backend-trainee-assignment-2024/internal/transport"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	// Инициализация config
	cfg, err := config.Init()
	if err != nil {
		fmt.Printf("app.Run failed to init config: %s\n", err.Error())
		os.Exit(1)
	}

	// Инициализация логгера
	logger, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("app.Run failed to init logger: %s\n", err.Error())
		os.Exit(1)
	}

	// Инициализация хранилища
	logger.Info("initializing storages...")
	postgres, err := postgres.New(cfg.Postgres)
	if err != nil {
		logger.Error(fmt.Sprintf("app.Run failed to init database: %s\n", err.Error()))
		os.Exit(1)
	}

	// Инициализация репозиториев
	logger.Info("initializing repos...")
	repo := repo.NewBannerRepository(postgres)

	cache := redis.NewRedis(cfg.Redis)
	defer cache.Close()

	queue, err := rabbitmq.NewRabbitMQQueue(cfg.RabbitMQ)
	defer queue.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("app.Run failed to init queue: %s\n", err.Error()))
		os.Exit(1)
	}

	// Инициализация сервисов
	bannerService := banner.NewBannerService(cfg.BannerService, repo, cache, queue, logger)
	defer bannerService.Shutdown()

	authService := auth.NewAuthService(cfg.AuthService.SecretKey)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	server := transport.New(cfg.HTTPServer, logger, bannerService, authService)
	errCh := server.Start()

	select {
	case <-done:
		logger.Info("stopping server")
	case err := <-errCh:
		logger.Error("server stopped with error: " + err.Error())
	}

	err = server.Stop()
	if err != nil {
		logger.Error(fmt.Sprintf("app.Run failed to stop server: %s\n", err.Error()))
		os.Exit(1)
	}

	logger.Info("app.Run server ended")
}
