package handlers

import (
	"context"
	"net/http"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/gofiber/fiber/v2"
)

const localUserKey = "user"

func WithViewHelpers(c *fiber.Ctx) error {
	c.Locals("activeFor", func(s string) (res string) {
		if c.Path() == s {
			return "active"
		}
		return ""
	})
	return c.Next()
}

func WithAuthenticatedUser(c *fiber.Ctx) error {
	c.Locals(localUserKey, nil)
	client := createSupabaseClient()

	token := c.Cookies("accessToken")

	if len(token) == 0 {
		return c.Next()
	}

	user, err := client.Auth.User(context.Background(), token)
	if err != nil {
		c.ClearCookie("accessToken")
		return c.Redirect("/")
	}

	ourUser := &data.User{ID: user.ID, Email: user.Email}
	c.Locals(localUserKey, ourUser)
	return c.Next()
}

func WithMustBeAuthenticated(c *fiber.Ctx) error {
	user := getAuthenticatedUser(c)
	if user == nil {
		return c.RedirectBack("/")
	}
	return c.Next()
}

func NotFoundMiddleware(c *fiber.Ctx) error {
	return c.Status(http.StatusNotFound).Render("error/404", nil)
}
