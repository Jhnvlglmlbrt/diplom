package handlers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"

	"github.com/Jhnvlglmlbrt/monitoring-certs/util"

	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/gofiber/fiber/v2"
	"github.com/nedpals/supabase-go"

	"github.com/sujit-baniya/flash"
)

var store = session.New()

type SignupParams struct {
	Email    string
	Fullname string
	Password string
}

func HandleGetSignup(c *fiber.Ctx) error {
	selectedPlan := c.Query("p")
	c.Cookie(&fiber.Cookie{
		Secure:   true,
		HTTPOnly: true,
		Name:     "p",
		Value:    selectedPlan,
	})

	// logger.Log("plan", selectedPlan)
	return c.Render("auth/signup", fiber.Map{})
}

func HandleSignupWithEmail(c *fiber.Ctx) error {
	session, err := store.Get(c)
	if err != nil {
		return err
	}

	selectedPlan := c.Cookies("p")
	// logger.Log("selectedPlan", selectedPlan)

	var params SignupParams
	if err := c.BodyParser(&params); err != nil {
		return err
	}
	if errors := params.Validate(); len(errors) > 0 {
		errors["email"] = params.Email
		errors["fullname"] = params.Fullname
		return flash.WithData(c, errors).Redirect("/signup?p=" + selectedPlan)
	}
	client := createSupabaseClient()
	resp, err := client.Auth.SignUp(context.Background(), supabase.UserCredentials{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		return err
	}

	session.Set("pass", params.Password)
	session.Set("email", params.Email)
	if err := session.Save(); err != nil {
		return err
	}

	pageLoadTime := time.Now().UnixNano() / int64(time.Millisecond)

	logger.Log("msg", "user signup with email", "id", resp.ID)
	return c.Render("auth/confirmation", fiber.Map{"email": params.Email, "pageLoadTime": pageLoadTime})
}

func (p SignupParams) Validate() fiber.Map {
	data := fiber.Map{}
	if !util.IsValidEmail(p.Email) {
		data["emailError"] = "Пожалуйста введите правильный email"
	}
	if !util.IsValidPassword(p.Password) {
		data["passwordError"] = "Пожалуйста используйте более сложный пароль"
	}
	if len(p.Fullname) < 3 {
		data["fullnameError"] = "Пожалуйста используйте реальное имя"
	}
	return data
}

func HandleGetSignin(c *fiber.Ctx) error {
	return c.Render("auth/signin", fiber.Map{})
}

func HandleSigninWithEmail(c *fiber.Ctx) error {
	var credentials supabase.UserCredentials

	if err := c.BodyParser(&credentials); err != nil {
		return err
	}

	client := createSupabaseClient()
	errors := fiber.Map{}
	resp, err := client.Auth.SignIn(context.Background(), credentials)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid login credentials") {
			errors["authError"] = "Неправильные данные, попробуйте снова"
		}
		return flash.WithData(c, errors).Redirect("/signin")
	}

	return c.Redirect("/auth/callback/" + resp.AccessToken)
}

func HandleGetSignout(c *fiber.Ctx) error {
	client := createSupabaseClient()
	if err := client.Auth.SignOut(c.Context(), c.Cookies("accessToken")); err != nil {
		return err
	}
	c.ClearCookie("accessToken")
	return c.Redirect("/")
}

// This is the main callback that will be triggered after each authentication.
func HandleAuthCallback(c *fiber.Ctx) error {
	accessToken := c.Params("accessToken")
	if len(accessToken) == 0 {
		return fmt.Errorf("invalid access token")
	}
	c.Cookie(&fiber.Cookie{
		Secure:   true,
		HTTPOnly: true,
		Name:     "accessToken",
		Value:    accessToken,
	})

	client := createSupabaseClient()
	user, err := client.Auth.User(context.Background(), accessToken)
	if err != nil {
		return err
	}

	selectedPlan := c.Cookies("pn")

	acc, err := data.CreateAccountForUserIfNotExist(user, selectedPlan, data.SubscriptionStatusActive)
	if err != nil {
		return err
	}

	logger.Log("info", "user signin", "userID", user.ID, "accountID", acc.ID)

	var redirectTo = "/domains"
	return c.Redirect(redirectTo)
}

func HandleResendConfirmationEmail(c *fiber.Ctx) error {
	session, err := store.Get(c)
	if err != nil {
		return err
	}

	signupParams, err := getSignupParams(session)
	if err != nil {
		return err
	}

	client := createSupabaseClient()
	resp, err := client.Auth.SignUp(context.Background(), supabase.UserCredentials{
		Email:    signupParams.Email,
		Password: signupParams.Password,
	})
	if err != nil {
		return err
	}

	logger.Log("msg", "user signup with email", "id", resp.ID)

	return nil
}

func createSupabaseClient() *supabase.Client {
	return supabase.CreateClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"), false)
}

func getSignupParams(session *session.Session) (SignupParams, error) {
	email := session.Get("email")
	if email == nil {
		return SignupParams{}, fmt.Errorf("email not found in session")
	}

	pass := session.Get("pass")
	if pass == nil {
		return SignupParams{}, fmt.Errorf("password not found in session")
	}

	emailStr, ok := email.(string)
	if !ok {
		return SignupParams{}, fmt.Errorf("failed to assert email as string")
	}

	passStr, ok := pass.(string)
	if !ok {
		return SignupParams{}, fmt.Errorf("failed to assert password as string")
	}

	return SignupParams{
		Email:    emailStr,
		Password: passStr,
	}, nil
}
