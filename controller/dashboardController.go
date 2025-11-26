package controller

import (
	"github.com/gofiber/fiber/v2"
)

func Dashboard(c *fiber.Ctx) error {
	//idDomain := c.Params("dmn")
	/*var dataDomain models.Domain
	idDomain := c.Params("dmn")
	database.DB.Where("id=?", idDomain).First(&dataDomain)
	if dataDomain.Id == 0 {
		c.Status(404)
		return c.JSON(fiber.Map{
			"message": "Page not Found",
		})
	}
	existModul := strings.Split(dataDomain.Modul, ",")
	fixModul := make(map[string]map[string]string)
	cookieModul := new(fiber.Cookie)
	if c.Cookies("modulExist") == "" {
		for _, key := range existModul {
			fixModul[key] = util.ModulDesc(key)
		}
		jsonData, err := json.Marshal(fixModul)
		if err != nil {
			return err
		}
		jsonString := string(jsonData)
		cookieModul.Name = "modulExist"
		cookieModul.Value = jsonString
		cookieModul.HTTPOnly = true

		c.Cookie(cookieModul)
	} else {
		c.Status(200)
		return c.JSON(fiber.Map{
			"message": "Cookie exists",
		})
	}
	c.Status(200)
	return c.JSON(fiber.Map{
		"message":    "Page ready to action",
		"data_modul": c.Cookies("modulExist"),
		"id_domain":  dataDomain,
	})*/
	return c.JSON(fiber.Map{
		"message": "Page ready to action",
	})
}
