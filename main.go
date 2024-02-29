package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Jhnvlglmlbrt/monitoring-certs/cmd/app"
	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler:          handlers.ErrorHandler,
		DisableStartupMessage: true,
		PassLocalsToViews:     true,
		Views:                 app.CreateEngine(),
	})

	db.Init()

	initRoutes(app)
	listenAddr := os.Getenv("HTTP_LISTEN_ADDR")
	fmt.Printf("app listening on: http://127.0.0.1:%s\n", listenAddr)
	log.Fatal(app.Listen(listenAddr))
	// app.Get("/admin", handlers.HandleGetAdmin)

}

func initRoutes(app *fiber.App) {
	app.Static("/static", "./static")
	app.Use(favicon.New(favicon.ConfigDefault))
	app.Use(handlers.WithAuthenticatedUser)
	app.Use(handlers.WithViewHelpers)
	app.Get("/", handlers.HandleGetHome)
	app.Get("/signin", handlers.HandleGetSignin)
	app.Get("/signout", handlers.HandleGetSignout)
	app.Get("/auth/callback/google", handlers.HandleGetCallbackGoogle)

	app.Get("/domains", handlers.HandleGetDashboard)
	// app.Post("/domains", handlers.WithMustBeAuthenticated, handlers.HandlePostDomain)
	// app.Post("/domains/new", handlers.WithMustBeAuthenticated, handlers.HandleGetDomainNew)

	app.Use(handlers.NotFoundMiddleware)
}
