package controller

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/models"
	"github.com/pujiutomo/cmsbackend/util"
	"gorm.io/gorm"
)

// Constants
const (
	UserKeyPattern    = "user:%d"
	MinPasswordLength = 6
)

// Request structs
type RegisterRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Password    string `json:"password"`
	DomainsId   string `json:"domains_id"`
	AccessRight string `json:"access_right"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// get redis domain
func getUserDomainsFromRedis(domainIDs string) ([]map[string]interface{}, error) {
	if strings.TrimSpace(domainIDs) == "" {
		return nil, fmt.Errorf("domain IDs cannot be empty")
	}

	domains := strings.Split(domainIDs, ",")
	var dataDomain []map[string]interface{}

	for _, domainID := range domains {
		trimmedID := strings.TrimSpace(domainID)
		if trimmedID == "" {
			continue
		}

		domainKey := "domain:" + trimmedID
		val, err := GetFromRedis(domainKey)
		fmt.Println(val)
		if err != nil {
			log.Printf("Error getting domain %s from Redis: %v", domainKey, err)
			continue
		}
		dataDomain = append(dataDomain, val...)
	}
	if len(dataDomain) == 0 {
		return nil, fmt.Errorf("no valid domains found")
	}
	return dataDomain, nil
}

func Register(c *fiber.Ctx) error {
	var request RegisterRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unable to parse request body",
			"error":   err.Error(),
		})
	}

	//validasi
	validator := util.NewValidator().
		Required(request.FirstName, "first_name").
		Required(request.LastName, "last_name").
		Required(request.Email, "email").
		Email(request.Email, "email").
		Required(request.Password, "password").
		MinLength(request.Password, 6, "password").
		Required(request.DomainsId, "domain")

	if validator.HasErrors() {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation field",
			"errors":  validator.Errors(),
		})
	}

	// Check if email already exists
	var existingUser models.User
	result := database.DB.Where("email = ?", strings.TrimSpace(request.Email)).First(&existingUser)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Printf("Error checking email existence: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error checking email availability",
			"error":   result.Error.Error(),
		})
	}

	if existingUser.Id != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email already exists",
		})
	}

	//create user
	user := models.User{
		FirstName:   strings.TrimSpace(request.FirstName),
		LastName:    strings.TrimSpace(request.LastName),
		Email:       strings.TrimSpace(request.Email),
		Phone:       strings.TrimSpace(request.Phone),
		DomainsId:   strings.TrimSpace(request.DomainsId),
		AccessRight: strings.TrimSpace(request.AccessRight),
		Su:          "n",
	}

	//set password
	user.SetPassword(request.Password)

	//save to database
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create user account",
			"errors":  err.Error(),
		})
	}

	//user response no sensitive data
	userResponse := fiber.Map{
		"id":         user.Id,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.Phone,
		"su":         user.Su,
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account create successfully",
		"user":    userResponse,
	})
}

func Login(c *fiber.Ctx) error {
	var request LoginRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "unable to parse request body",
			"errors":  err.Error(),
		})
	}

	//validation required
	validator := util.NewValidator().
		Required(request.Email, "email").
		Email(request.Email, "email").
		Required(request.Password, "password")

	if validator.HasErrors() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation faild",
			"errors":  validator.Errors(),
		})
	}
	dataEmail := strings.TrimSpace(request.Email)

	var user models.User
	//cek data to database
	result := database.DB.Where("email = ?", dataEmail).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Email address not found",
			})
		}
		log.Printf("Error finding user: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error Login",
			"errors":  result.Error.Error(),
		})
	}

	//check password
	if err := user.ComparePassword(request.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Incorrect Password",
		})
	}

	//geneter JWT token
	//accessToken, accessExpires, err := util.GenerateAccessToken(&user)
	token, err := util.GenerateJwt(strconv.Itoa(int(user.Id)))
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error generating authentication token",
			"error":   err.Error(),
		})
	}

	//set cookie
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	}
	c.Cookie(&cookie)

	//get data domains from redis
	domains, err := getUserDomainsFromRedis(user.DomainsId)
	if err != nil {
		log.Printf("Error getting domains: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get domains",
			"error":   err.Error(),
		})
	}

	//prepare data user untuk masuk ke redis dan respon ke frontend
	dataUser := fiber.Map{
		"id":         user.Id,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.Phone,
		"domain":     domains,
	}

	//simpan data ke redis
	jsonData, _ := jsoniter.Marshal(dataUser)
	userKeyRedis := fmt.Sprintf(UserKeyPattern, user.Id)
	if err := SaveToRedis(userKeyRedis, string(jsonData)); err != nil {
		log.Printf("warning: failed to save data domain to redis: %v", err)
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"data":    dataUser,
		"token":   token,
	})
}

func Logout(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"message": "OK",
	})
}

func refreshTokenHandler(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Refresh token required",
		})
	}

	// verify refresh token
	user, err := util.VerifyRefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	//generate new access token
	accessToken, accessExpires, err := util.GenerateAccessToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate new token",
		})
	}

	//set new access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    accessToken,
		Expires:  accessExpires,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})
	return c.JSON(fiber.Map{
		"message":    "Token refreshed successfully",
		"expires_at": accessExpires.Format(time.RFC3339),
	})
}
