package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aviraltrip/urlshortener/config"
	"github.com/aviraltrip/urlshortener/database"
	"github.com/aviraltrip/urlshortener/handlers"
	"github.com/aviraltrip/urlshortener/middleware"
	"github.com/aviraltrip/urlshortener/services"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pgPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] PostgreSQL connection: %v", err)
	} else {
		defer pgPool.Close()
		if err := database.RunMigrations(ctx, pgPool); err != nil {
			log.Fatalf("[FATAL] Migration error: %v", err)
		}
		log.Println("[INFO] PostgreSQL connected and migrations applied")
	}
	redisClients, err := database.NewRedisClients(cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		log.Printf("[WARN] Redis connection: %v", err)
	} else {
		defer redisClients.Close()
		log.Println("[INFO] Redis connected")
	}
	var redisCache *redis.Client
	var redisRateLimit *redis.Client
	if redisClients != nil {
		redisCache = redisClients.Cache
		redisRateLimit = redisClients.RateLimit
	}
	linkService := services.NewLinkService(cfg, pgPool, redisCache)
	app := fiber.New(fiber.Config{
		AppName: "Go Fiber URL Shortener",
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Get("/health", handlers.HealthCheck)
	app.Get("/ready", handlers.ReadyCheck(pgPool, redisClients))
	app.Get("/:url", handlers.ResolveURL(linkService))

	api := app.Group("/api/v1")
	if redisRateLimit != nil {
		api.Use(middleware.NewRateLimiter(redisRateLimit, cfg.ApiQuota))
	}
	api.Post("/", handlers.ShortenURL(linkService))
	api.Get("/links", handlers.ListLinks(linkService))
	api.Get("/links/:code", handlers.GetLinkStats(linkService))
	api.Get("/links/:code/stats", handlers.GetLinkStats(linkService))
	api.Delete("/links/:code", handlers.DeleteLink(linkService))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[INFO] Shutting down server gracefully...")
		_ = app.Shutdown()
	}()
	log.Printf("[INFO] Server starting on port %s", cfg.AppPort)
	if err := app.Listen(cfg.AppPort); err != nil {
		log.Printf("[INFO] Server stopped: %v", err)
	}
}
