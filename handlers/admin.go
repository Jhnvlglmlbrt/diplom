package handlers

import "github.com/gofiber/fiber/v2"

func HandleGetAdmin(c *fiber.Ctx) error {
	// user := getAuthenticatedUser(c)
	// if user != nil {
	// 	if !user.IsAdmin {
	// 		return c.Status(403).SendString("Forbidden")
	// 	}
	// 	return c.Render("home/admin", fiber.Map{})
	// }
	return c.Render("home/index", fiber.Map{})
}

// TODO: admin panel, if user is admin -> admin route
// mb can be made in the same route as just user
