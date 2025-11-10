package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/routes"
)

func main() {
	database.Connnect()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro loading .env files")
	}
	port := os.Getenv("PORT")
	app := fiber.New()
	routes.Setup(app)
	app.Listen(":" + port)
}
