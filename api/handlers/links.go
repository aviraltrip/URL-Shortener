package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/aviraltrip/urlshortener/helpers"
	"github.com/aviraltrip/urlshortener/services"
	"github.com/gofiber/fiber/v2"
)

type ShortenRequest struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
	Expiry      int    `json:"expiry"`
}

type ShortenResponse struct {
	URL       string    `json:"url"`
	Short     string    `json:"short"`
	Expiry    int       `json:"expiry"`
	ExpiresAt time.Time `json:"expires_at"`
}

func ShortenURL(svc *services.LinkService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body ShortenRequest
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid JSON payload",
			})
		}
		if body.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "url is required",
			})
		}
		res, err := svc.Shorten(c.Context(), body.URL, body.CustomShort, body.Expiry)
		if err != nil {
			if errors.Is(err, helpers.ErrInvalidURL) || errors.Is(err, helpers.ErrSelfDomain) ||
				errors.Is(err, helpers.ErrInvalidAlias) || errors.Is(err, helpers.ErrAliasKeyword) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			if errors.Is(err, services.ErrAliasConflict) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to shorten URL",
			})
		}
		return c.Status(fiber.StatusOK).JSON(ShortenResponse{
			URL:       res.Link.OriginalUrl,
			Short:     res.ShortURL,
			Expiry:    body.Expiry,
			ExpiresAt: res.Link.ExpiresAt,
		})
	}
}

func GetLinkStats(svc *services.LinkService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("code")
		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "code is required"})
		}
		link, err := svc.GetStats(c.Context(), code)
		if err != nil {
			if errors.Is(err, services.ErrNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "link not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get stats"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"short":      link.ShortCode,
			"url":        link.OriginalUrl,
			"clicks":     link.ClickCount,
			"created_at": link.CreatedAt,
			"expires_at": link.ExpiresAt,
		})
	}
}

func DeleteLink(svc *services.LinkService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("code")
		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "code is required"})
		}
		if err := svc.Delete(c.Context(), code); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete link"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "link deleted successfully",
		})
	}
}

func ListLinks(svc *services.LinkService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))
		links, err := svc.List(c.Context(), int32(page), int32(limit))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list links"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"links": links,
			"page":  page,
			"limit": limit,
		})
	}
}
