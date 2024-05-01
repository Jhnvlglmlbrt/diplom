package handlers

import (
	"fmt"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/gofiber/fiber/v2"
)

func HandleGetHome(c *fiber.Ctx) error {
	accessToken := c.Query("access_token")
	// logger.Log("accessToken", accessToken)
	if len(accessToken) > 0 {
		return c.Redirect("/auth/callback/" + accessToken)
	}
	return c.Render("home/index", fiber.Map{})
}

// HandleGetReset(c *fiber.Ctx) error {
// 	token := c.Query("token")
// 	if len(accessToken) > 0 {
// 		return c.Redirect("/auth/callback/" + accessToken)
// 	}
// 	return c.Render("home/index", fiber.Map{})
// }

func HandleGetPricing(c *fiber.Ctx) error {
	context := fiber.Map{
		"starterDomains":    2,
		"businessDomains":   50,
		"enterpriseDomains": 200,
	}
	return c.Render("home/pricing", context)
}

func HandleGetPlans(c *fiber.Ctx) error {
	plans, err := data.GetAllPlans()
	if err != nil {
		return err
	}

	return c.Render("home/pricing1", fiber.Map{
		"plans":       plans,
		"formatPrice": formatPrice,
	})
}

func formatPrice(price float64) string {
	return fmt.Sprintf("%.f", price)
}
