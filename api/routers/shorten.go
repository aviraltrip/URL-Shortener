package routers

import "github.com/aviraltrip/urlshortener/helpers"

type request struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
	Expiry      string `json:"expiry"`
}

type response struct {
	URL            string `json:"url"`
	CustomShort    string `json:"short"`
	Expiry         string `json:"expiry"`
	XRateRemaining int    `json:"rate_limit"`
	XRateLimitRest int    `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "can't parse json"})
	}

	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error", "you cant hack the system"})
	}

	body.URL = helpers.EnforceHTTP(body.URL)
}
