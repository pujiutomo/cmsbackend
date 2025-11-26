package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pujiutomo/cmsbackend/util"
)

var SecretKeyToken string = os.Getenv("SECRETKEY")
var RefreshSecretKey string = os.Getenv("REFRESHSCRETKEY")

func IsAuthenticate(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	if _, err := util.ParseJwt(cookie); err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Unauthenticated",
		})
	}
	return c.Next()
}

func AuthMiddleware(c *fiber.Ctx) error {
	//cek token dari cookie
	tokenString := c.Cookies("auth_token")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	//veryfy token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKeyToken), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	//set user data ke context
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("user", claims)

	return c.Next()
}

func CSRFMiddleware(c *fiber.Ctx) error {
	if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
		return c.Next()
	}

	//verify CSRF token
	clientToken := c.Get("X-CSRF-Token")
	cookieToken := c.Cookies("csrf_token")

	if clientToken == "" || clientToken != cookieToken {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid CSRF Token",
		})
	}

	return c.Next()
}
