package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/handlers"
	"github.com/Jhnvlglmlbrt/monitoring-certs/util"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/django/v3"
	"github.com/joho/godotenv"
)

func main() {
	app, err := initApp()
	if err != nil {
		log.Fatal(err)
	}

	db.Init()

	app.Static("/static", "./static", fiber.Static{
		CacheDuration: 0,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Surrogate-Control", "no-store")
		return c.Next()
	})

	app.Static("/static", "./static")

	app.Use(favicon.New(favicon.ConfigDefault))
	app.Use(recover.New())
	app.Use(handlers.WithFlash)
	app.Use(handlers.WithAuthenticatedUser)
	app.Use(handlers.WithViewHelpers)
	app.Get("/", handlers.HandleGetHome)
	app.Get("/pricing", handlers.HandleGetPricing)

	app.Get("/signin", handlers.HandleGetSignin)
	// app.Post("/signin", handlers.HandleSigninWithEmail)
	app.Get("/signout", handlers.HandleGetSignout)
	// app.Get("/signup", handlers.HandleGetSignup)
	// app.Post("/signup", handlers.HandleGetSignupWithEmail)

	app.Get("/auth/callback/google", handlers.HandleGetCallbackGoogle)

	app.Get("/domains", handlers.HandleGetDashboard)
	// app.Post("/domains", handlers.WithMustBeAuthenticated, handlers.HandlePostDomain)
	// app.Post("/domains/new", handlers.WithMustBeAuthenticated, handlers.HandleGetDomainNew)

	// app.Get("/admin", handlers.HandleGetAdmin)

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
		Views:                 createEngine(),
	})
	return app, nil

}

func createEngine() *django.Engine {
	engine := django.New("./views", ".html")
	engine.Reload(true)
	engine.AddFunc("css", func(name string) (res template.HTML) {
		filepath.Walk("static/assets", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == name {
				res = template.HTML("<link rel=\"stylesheet\" href=\"/" + path + "\">")
			}
			return nil
		})
		return
	})

	engine.AddFunc("badgeForStatus", func(status string) (res string) {
		switch status {
		case data.StatusOffline:
			return fmt.Sprintf(`<div class="badge badge-accent">%s</div>`, status)
		case data.StatusHealthy:
			return fmt.Sprintf(`<div class="badge badge-success">%s</div>`, status)
		case data.StatusExpires:
			return fmt.Sprintf(`<div class="badge badge-warning">%s</div>`, status)
		case data.StatusExpired:
			return fmt.Sprintf(`<div class="badge badge-accent">%s</div>`, status)
		case data.StatusUnresponsive:
			return fmt.Sprintf(`<div class="badge badge-accent">%s</div>`, status)
		case data.StatusInvalid:
			return fmt.Sprintf(`<div class="badge badge-error">%s</div>`, status)
		}
		return ""
	})

	engine.AddFunc("formatTime", func(t time.Time) (res string) {
		timeZero := time.Time{}
		if t.Equal(timeZero) {
			return "n/a"
		}
		return t.Format(time.DateTime)

	})
	engine.AddFunc("timeAgo", func(t time.Time) (res string) {
		x := time.Since(t).Seconds()
		return fmt.Sprintf("%v seconds ago", math.Round(x))
	})
	engine.AddFunc("daysLeft", func(t time.Time) (res string) {
		return util.DaysLeft(t)
	})
	engine.AddFunc("pluralize", func(s string, n int) (res string) {
		return util.Pluralize(s, n)
	})
	return engine
}
