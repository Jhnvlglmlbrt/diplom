package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func HandleGetHome(c *fiber.Ctx) error {
	accessToken := c.Query("access_token")
	if len(accessToken) > 0 {
		return c.Redirect("/auth/callback/" + accessToken)
	}
	return c.Render("home/index", fiber.Map{})
}

func HandleGetPricing(c *fiber.Ctx) error {
	context := fiber.Map{
		"starterDomains":    2,
		"businessDomains":   50,
		"enterpriseDomains": 500,
	}
	return c.Render("home/pricing", context)
}
