package handlers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/nedpals/supabase-go"
)

func HandleGetSignout(c *fiber.Ctx) error {
	client := createSupabaseClient()
	if err := client.Auth.SignOut(c.Context(), c.Cookies("accessToken")); err != nil {
		return err
	}
	c.ClearCookie("accessToken")
	return c.Redirect("/")
}

func HandleGetSignin(c *fiber.Ctx) error {
	client := createSupabaseClient()
	resp, err := client.Auth.SignInWithProvider(supabase.ProviderSignInOptions{
		Provider:   "google",
		RedirectTo: "http://localhost:3000/auth/callback/google",
		FlowType:   supabase.PKCE,
	})

	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Secure:  true,
		Expires: time.Now().Add(time.Minute * 1),
		Value:   resp.CodeVerifier,
		Name:    "verifier",
	})

	return c.Redirect(resp.URL)
}

func HandleGetCallbackGoogle(c *fiber.Ctx) error {
	var (
		client = createSupabaseClient()
		code   = c.Query("code")
	)

	if code == "" {
		return fmt.Errorf("oauth exchange code not valid")
	}
	verifier := c.Cookies("verifier")
	resp, err := client.Auth.ExchangeCode(context.Background(), supabase.ExchangeCodeOpts{
		AuthCode:     code,
		CodeVerifier: verifier,
	})
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Secure:   true,
		HTTPOnly: true,
		Name:     "accessToken",
		Value:    resp.AccessToken,
	})

	user := &supabase.User{
		ID:    resp.User.ID,
		Email: resp.User.Email,
	}

	acc, err := data.CreateAccountForUserIfNotExists(user)
	if err != nil {
		return err
	}

	logger.Log("info", "user sigin", "userID", resp.User.ID, "accountID", acc.ID)
	return c.Redirect("/domains")
}

func createSupabaseClient() *supabase.Client {
	return supabase.CreateClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"), false)
}
