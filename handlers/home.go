package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func HandleGetHome(c *fiber.Ctx) error {
	user := getAuthenticatedUser(c)
	if user != nil {
		fmt.Println("user is authenticated")
	}

	return c.Render("home/index", fiber.Map{})
}
