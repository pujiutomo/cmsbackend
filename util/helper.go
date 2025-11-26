package util

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var SecretKey string = os.Getenv("secret")

func GenerateJwt(issuer string) (string, error) {
	//create clams
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		//ExpiresAt: time.Now().Add(time.Hour*24).Unix(),
		Issuer: issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseJwt(cookie string) (string, error) {
	token, err := jwt.ParseWithClaims(cookie, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil || token.Valid {
		return "", err
	}
	claims := token.Claims.(*jwt.RegisteredClaims)
	return claims.Issuer, nil
}

// Helper convert functions
func ConvertToUint(id interface{}) (uint, error) {
	switch v := id.(type) {
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("id cannot be negative")
		}
		return uint(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("id cannot be negative")
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("id cannot be negative")
		}
		return uint(v), nil
	case uint:
		return v, nil
	case string:
		if v == "" {
			return 0, fmt.Errorf("id cannot be empty")
		}
		idInt, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid id format: %s", v)
		}
		return uint(idInt), nil
	default:
		return 0, fmt.Errorf("unsupported id type: %T", id)
	}
}

// modul help util
type multiModul map[string]interface{}

func ModulDesc(idModul string) multiModul {
	data := make(multiModul)
	switch idModul {
	case "package_tour":
		data["package_tour"] = map[string]interface{}{
			"name": "Package Tour",
			"icon": "-",
		}

	case "rent_car":
		data["rent_car"] = map[string]interface{}{
			"name": "Rent Car",
			"icon": "-",
		}

	case "blog":
		data["blog"] = map[string]interface{}{
			"name": "Blog",
			"icon": "-",
		}

	case "attractions":
		data["attractions"] = map[string]interface{}{
			"name": "Attractions",
			"icon": "-",
		}

	case "gallery":
		data["gallery"] = map[string]interface{}{
			"name": "Gallery",
			"icon": "-",
		}

	case "office":
		data["office"] = map[string]interface{}{
			"name": "Office",
			"icon": "-",
		}
	}
	return data
}
