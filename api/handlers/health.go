package handlers

import (
	"context"
	"time"

	"github.com/aviraltrip/urlshortener/database"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}
func ReadyCheck(pool *pgxpool.Pool, rdb *database.RedisClients) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		status := fiber.Map{
			"status":   "ok",
			"postgres": "up",
			"redis":    "up",
		}
		isReady := true
		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				status["postgres"] = "down"
				isReady = false
			}
		} else {
			status["postgres"] = "not_configured"
			isReady = false
		}
		if rdb != nil && rdb.Cache != nil {
			if err := rdb.Cache.Ping(ctx).Err(); err != nil {
				status["redis"] = "down"
				isReady = false
			}
		} else {
			status["redis"] = "not_configured"
			isReady = false
		}
		if !isReady {
			status["status"] = "degraded"
			return c.Status(fiber.StatusServiceUnavailable).JSON(status)
		}
		return c.Status(fiber.StatusOK).JSON(status)
	}
}
