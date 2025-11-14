package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pujiutomo/cmsbackend/controller"
	"github.com/pujiutomo/cmsbackend/middleware"
)

func Setup(app *fiber.App) {
	app.Post("/api/login", controller.Login)
	app.Use(middleware.IsAuthenticate)
	app.Post("/api/register", controller.Register)
	app.Post("/api/domain/post", controller.PostDomain)
	app.Get("/api/domain", controller.GetDomain)
	app.Get("/api/dashboard/:dmn<int>?", controller.Dashboard)
}
