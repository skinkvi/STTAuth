package jwtT

import (
	"STTAuth/internal/domain/models"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestNewToken(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "test_user_email@example.com",
	}

	app := models.App{
		ID:     1,
		Secret: "test_app_secret",
	}

	duration := time.Hour * 24

	tokenString, err := NewToken(user, app, duration)
	assert.NoError(t, err)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(app.Secret), nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(user.ID), claims[UIDKey])
	assert.Equal(t, user.Email, claims[EmailKey])
	assert.NotZero(t, claims[ExpKey])
	assert.Equal(t, float64(app.ID), claims[AppIDKey])
}
