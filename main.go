package main

import (
	"log"

	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/engine"
	"github.com/Jhnvlglmlbrt/monitoring-certs/handlers"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/joho/godotenv"
)

func main() {
	app, err := initApp()
	if err != nil {
		log.Fatal(err)
	}

	db.Init()
	logger.Init()

	app.Static("/static", "./static", fiber.Static{
		CacheDuration: 0,
	})

	app.Static("/static", "./static")

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Surrogate-Control", "no-store")
		return c.Next()
	})

	app.Use(favicon.New(favicon.ConfigDefault))
	app.Use(recover.New())
	app.Use(handlers.WithFlash)
	app.Use(handlers.WithAuthenticatedUser)
	app.Use(handlers.WithViewHelpers)
	app.Get("/", handlers.HandleGetHome)

	app.Get("/pricing", handlers.HandleGetPlans)
	app.Get("/about", handlers.HandleGetAbout)

	app.Get("/check", handlers.HandleGetCheck)
	app.Post("/check_domain", handlers.HandleCheckDomainStatus)

	app.Get("/signup", handlers.HandleGetSignup)
	app.Post("/signup", handlers.HandleSignupWithEmail)
	app.Get("/signin", handlers.HandleGetSignin)
	app.Post("/signin", handlers.HandleSigninWithEmail)
	app.Get("/signout", handlers.HandleGetSignout)
	app.Get("/auth/callback/:accessToken", handlers.HandleAuthCallback)

	app.Get("/confirmation", handlers.HandleResendConfirmationEmail)

	app.Get("/reset", handlers.HandleGetReset)
	app.Post("/reset", handlers.HandleReset)

	app.Get("/update_pass", handlers.HandleGetUpdatePassword)
	app.Post("/update_pass", handlers.HandleUpdatePassword)

	// Вкладка с аккаунтом
	account := app.Group("/account", handlers.WithMustBeAuthenticated)
	account.Get("/", handlers.HandleAccountShow)
	account.Post("/", handlers.HandleAccountUpdate)

	// Вкладка с избранными доменами
	favorites := app.Group("/favorites", handlers.WithMustBeAuthenticated)
	favorites.Get("/", handlers.HandleFavoritesList)
	favorites.Post("/add_favorite", handlers.HandleAddFavorite)
	favorites.Post("/delete_favorite", handlers.HandleRemoveFavorite)

	// Вкладка с добавлением доменов, с просмотром доменов
	domains := app.Group("/domains", handlers.WithMustBeAuthenticated)
	domains.Get("/", handlers.HandleDomainList)
	domains.Post("/", handlers.HandleDomainCreate)
	domains.Post("/delete", handlers.HandleDomainsDelete)
	domains.Get("/new", handlers.HandleDomainNew)
	domains.Get("/:id", handlers.HandleDomainShow)
	domains.Get("/:id/raw", handlers.HandleDomainShowRaw)
	domains.Post("/:id/delete", handlers.HandleDomainDelete)
	domains.Get("/:id/test_notification", handlers.HandleSendTestNotification)

	// Admin panel (users, accounts, domains, plans)
	app.Get("/admin/auth/callback/:accessToken", handlers.HandleAuthCallbackWithAdmin)
	admin := app.Group("/admin", handlers.WithMustBeAuthenticated)
	admin.Get("/users", handlers.HandleUsersList)
	admin.Post("/users/delete", handlers.HandleUsersDelete)
	admin.Post("/users/create", handlers.HandleUsersCreate)

	admin.Get("/accounts", handlers.HandleAccountsList)
	admin.Post("/accounts/delete", handlers.HandleAccountsDelete)
	admin.Post("/accounts/update", handlers.HandleAccountsUpdate)

	admin.Get("/domains", handlers.HandleAdminDomainList)
	admin.Post("/domains", handlers.HandleDomainCreate)
	admin.Post("/domains/delete", handlers.HandleAdminDomainsDelete)
	admin.Get("/domains/:id", handlers.HandleAdminDomainShow)
	admin.Get("/domains/:id/raw", handlers.HandleAdminDomainShowRaw)
	admin.Post("/domains/:id/delete", handlers.HandleAdminDomainsDelete)
	admin.Get("/domains/:id/test_notification", handlers.HandleSendTestNotification)

	admin.Get("/plans", handlers.HandlePlansList)
	admin.Post("/plans/create", handlers.HandlePlansCreate)
	admin.Post("/plans/delete", handlers.HandlePlansDelete)
	admin.Post("/plans/update", handlers.HandlePlansUpdate)

	// app.Use(handlers.NotFoundMiddleware)

	log.Fatal(app.Listen(":3000"))
}

func initApp() (*fiber.App, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler:          handlers.ErrorHandler,
		DisableStartupMessage: true,
		PassLocalsToViews:     true,
		Views:                 engine.CreateEngine(),
	})
	return app, nil

}
