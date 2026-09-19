package handlers

import (
	"errors"

	"github.com/aviraltrip/urlshortener/services"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(svc *services.LinkService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("url")
		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "short code is required"})
		}
		destURL, err := svc.Resolve(c.Context(), code)
		if err != nil {
			if errors.Is(err, services.ErrNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "short URL not found or expired",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to resolve URL",
			})
		}
		return c.Redirect(destURL, fiber.StatusTemporaryRedirect)
	}
}
