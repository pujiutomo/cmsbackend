package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/routes"
)

func main() {
	database.Connnect()
	database.RedisClient()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro loading .env files")
	}
	port := os.Getenv("PORT")
	app := fiber.New()
	app.Use(cors.New(getCORSConfig()))
	routes.Setup(app)
	app.Listen(":" + port)
}

func getCORSConfig() cors.Config {
	envOrigins := os.Getenv("ALLOWED_ORIGINS")

	if envOrigins == "" {
		envOrigins = "http://localhost:3000,http://127.0.0.1:3000"
	}

	allowedOrigins := strings.Split(envOrigins, ",")

	return cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// Allow in development
			if os.Getenv("APP_ENV") == "development" {
				return true
			}

			// Check against allowed list
			for _, allowed := range allowedOrigins {
				if strings.TrimSpace(origin) == strings.TrimSpace(allowed) {
					return true
				}
			}

			return false
		},
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowCredentials: true,
		MaxAge:           300,
	}
}
