package controller

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/models"
)

func validateDomain(domain string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(domain)
}

func PostDomain(c *fiber.Ctx) error {
	var data map[string]interface{}
	var domainPost models.Domain
	if err := c.BodyParser(&domainPost); err != nil {
		fmt.Println("Unable to parse Body")
	}
	if err := c.BodyParser(&data); err != nil {
		fmt.Println("Unable to parse Body")
	}
	database.DB.Where("name=?", strings.TrimSpace(data["name"].(string))).First(&domainPost)
	if domainPost.Id != 0 {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "Domain already exists",
		})
	}
	if !validateDomain(strings.TrimSpace(data["name"].(string))) {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "Invalid Domain Name",
		})
	}
	if err := database.DB.Create(&domainPost).Error; err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "Invalid payload",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Successfully Post Domain Setting",
	})
}
func GetDomain(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit := 5
	offset := (page - 1) * limit
	var total int64
	var getDomain []models.Domain
	database.DB.Offset(offset).Limit(limit).Find(&getDomain)
	database.DB.Model(&models.Domain{}).Count(&total)
	return c.JSON(fiber.Map{
		"data": getDomain,
		"meta": fiber.Map{
			"total":     total,
			"page":      page,
			"last_page": math.Ceil(float64(int(total) / limit)),
		},
	})
}
