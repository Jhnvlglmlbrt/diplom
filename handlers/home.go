package handlers

import (
	"fmt"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/fiber/v2"
)

func HandleGetHome(c *fiber.Ctx) error {
	accessToken := c.Query("access_token")

	if len(accessToken) > 0 {
		return c.Redirect("/auth/callback/" + accessToken)
	}
	return c.Render("home/index", fiber.Map{})
}

func HandleGetAbout(c *fiber.Ctx) error {
	return c.Render("home/about", fiber.Map{})
}

func HandleGetCheck(c *fiber.Ctx) error {
	return c.Render("home/check", fiber.Map{})
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

	spew.Dump(plans)

	return c.Render("home/pricing1", fiber.Map{
		"plans": plans,
	})
}

func formatPrice(price float64) string {
	return fmt.Sprintf("%.f", price)
}

// func GetPlans(limit int, page int, sortField string, ascending bool) ([]data.Plans, error) {
// 	if limit == 0 {
// 		limit = data.DefaultLimit
// 	}

// 	var plans []data.Plans

// 	builder := db.Bun.NewSelect().Model(&plans).Limit(limit)
// 	for k, v := range filter {
// 		if v != "" {
// 			builder.Where("? = ?", bun.Ident(k), v)
// 		}
// 	}
// 	offset := (page - 1) * limit
// 	builder.Offset(offset)
// 	if ascending {
// 		builder.OrderExpr("? ASC", bun.Ident(sortField))
// 	} else {
// 		builder.OrderExpr("? DESC", bun.Ident(sortField))
// 	}
// 	err := builder.Scan(context.Background())
// 	return accounts, err
// }
