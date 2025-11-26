package util

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/models"
)

var SecretKeyToken string = os.Getenv("SECRETKEY")
var RefreshSecretKey string = os.Getenv("REFRESHSCRETKEY")

func GenerateAccessToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(15 * time.Minute) //short live 15 minute

	claims := jwt.MapClaims{
		"user_id": user.Id,
		"email":   user.Email,
		"role":    user.AccessRight,
		"exp":     expiresAt.Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString([]byte(SecretKeyToken))

	return tokenString, expiresAt, err
}

func GenerateRefreshToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 hari

	claims := jwt.MapClaims{
		"user_id": user.Id,
		"exp":     expiresAt.Unix(),
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString([]byte(RefreshSecretKey))

	return tokenString, expiresAt, err
}

func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func VerifyRefreshToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(RefreshSecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	var UserModel models.User

	result := database.DB.Where("id = ?", userID).First(&UserModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return &UserModel, nil
}
