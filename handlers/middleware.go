package handlers

import (
	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/gofiber/fiber/v2"
)

const localUserKey = "user"

func WithAuthenticatedUser(c *fiber.Ctx) error {
	// c.Locals(localUserKey, nil)

	// authentication here
	// 1. user authenticated
	// 2. user not authenticated
	// c.Locals(LocalsUserKey, nil)

	c.Locals(localUserKey, &data.User{ID: 1, Email: "user@example.com"})
	return c.Next()
}
