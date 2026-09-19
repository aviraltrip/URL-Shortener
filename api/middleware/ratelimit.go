package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func NewRateLimiter(rdb *redis.Client, quota int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rdb == nil {
			return c.Next()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		ip := c.IP()
		key := "ratelimit:" + ip
		window := 30 * time.Minute
		val, err := rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			_ = rdb.Set(ctx, key, quota-1, window).Err()
			c.Set("X-RateLimit-Limit", strconv.Itoa(quota))
			c.Set("X-RateLimit-Remaining", strconv.Itoa(quota-1))
			c.Set("X-RateLimit-Reset", "30")
			return c.Next()
		} else if err != nil {
			return c.Next()
		}
		remaining, _ := strconv.Atoi(val)
		if remaining <= 0 {
			ttl, _ := rdb.TTL(ctx, key).Result()
			resetMin := int(ttl / time.Minute)
			if resetMin < 1 {
				resetMin = 1
			}
			c.Set("X-RateLimit-Limit", strconv.Itoa(quota))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", strconv.Itoa(resetMin))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":            "rate limit exceeded",
				"rate_limit_reset": resetMin,
			})
		}
		newRemaining, _ := rdb.Decr(ctx, key).Result()
		ttl, _ := rdb.TTL(ctx, key).Result()
		resetMin := int(ttl / time.Minute)
		c.Set("X-RateLimit-Limit", strconv.Itoa(quota))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(newRemaining, 10))
		c.Set("X-RateLimit-Reset", strconv.Itoa(resetMin))
		return c.Next()
	}
}
