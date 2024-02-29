package handlers

import (
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/gofiber/fiber/v2"
)

func HandleGetDashboard(c *fiber.Ctx) error {
	sslTrackings := []data.SSLTracking{
		{
			ID:         1,
			DomainName: "google.com",
			Issuer:     "Let's Encrypt!",
			Expires:    time.Now().AddDate(0, 0, 7),
			Status:     "healthy",
		},
		{
			ID:         2,
			DomainName: "amazon.com",
			Issuer:     "Let's Encrypt!",
			Expires:    time.Now().AddDate(0, 0, 14),
			Status:     "expires",
		},
		{
			ID:         3,
			DomainName: "golangforall.com",
			Issuer:     "Let's Encrypt!",
			Expires:    time.Now().AddDate(0, 1, 14),
			Status:     "expired",
		},
		{
			ID:         4,
			DomainName: "https://github.com/",
			Issuer:     "",
			Expires:    time.Now().AddDate(0, 1, 14),
			Status:     "invalid",
		},
		{
			ID:         5,
			DomainName: "https://tailwindflex.com/",
			Issuer:     "Something!",
			Expires:    time.Now().AddDate(0, 1, 14),
			Status:     "offline",
		},
	}

	data := fiber.Map{
		"trackings": sslTrackings,
	}
	return c.Render("domains/index", data)
}
