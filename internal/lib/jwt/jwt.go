package jwtT

import (
	"STTAuth/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	UIDKey   = "uid"
	EmailKey = "email"
	ExpKey   = "exp"
	AppIDKey = "app_id"
)

// Эта модель имеет риск быть логированной а в ней мы передаем секрет так что
// TODO: Нужно что то сделать с тем как прятать секрет что бы не спалить его в логах
func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims[UIDKey] = user.ID
	claims[EmailKey] = user.Email
	claims[ExpKey] = time.Now().Add(duration).Unix()
	claims[AppIDKey] = app.ID

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
