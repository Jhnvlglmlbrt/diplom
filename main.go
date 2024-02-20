package main

import (
	"log"
	"os"

	"github.com/Jhnvlglmlbrt/monitoring-certs/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/template/django/v3"
	"github.com/joho/godotenv"
)

func main() {
	app, err := initApp()
	if err != nil {
		log.Fatal(err)
	}

	app.Static("/static", "./static")
	app.Use(favicon.New(favicon.ConfigDefault))
	app.Use(handlers.WithAuthenticatedUser)
	app.Get("/", handlers.HandleGetHome)

	log.Fatal(app.Listen(os.Getenv("HTTP_LISTEN_ADDR")))
}

func initApp() (*fiber.App, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	engine := django.New("./views", ".html")
	engine.Reload(true)
	app := fiber.New(
		fiber.Config{
			Views:                 engine,
			PassLocalsToViews:     true,
			DisableStartupMessage: true,
		})

	return app, nil
}
