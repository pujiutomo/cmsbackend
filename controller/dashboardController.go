package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/models"
)

func Dashboard(c *fiber.Ctx) error {
	var dataDomain models.Domain
	idDomain := c.Params("dmn")
	database.DB.Where("id=?", idDomain).First(&dataDomain)
	if dataDomain.Id == 0 {
		c.Status(404)
		return c.JSON(fiber.Map{
			"message": "Page not Found",
		})
	}
	c.Status(200)
	return c.JSON(fiber.Map{
		"message": "Page ready to action",
	})
}
