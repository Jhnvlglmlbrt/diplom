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
			Status:     "Ok",
		},
		{
			ID:         2,
			DomainName: "amazon.com",
			Issuer:     "Let's Encrypt!",
			Expires:    time.Now().AddDate(0, 0, 14),
			Status:     "Ok",
		},
		{
			ID:         3,
			DomainName: "golangforall.com",
			Issuer:     "Let's Encrypt!",
			Expires:    time.Now().AddDate(0, 1, 14),
			Status:     "Ok",
		},
	}

	data := fiber.Map{
		"trackings": sslTrackings,
	}
	return c.Render("dashboard/index", data)
}
