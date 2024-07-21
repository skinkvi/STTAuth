package tests

import (
	"STTAuth/tests/suite"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/skinkvi/protosSTT/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type validatorKey string

const (
	emptyAppID   = 0
	appID        = 1
	appSecret    = "test-secret"
	deltaSeconds = 1

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	loginTime := time.Now()

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	// check if exp of token is in correct range, ttl get from st.Cfg.TokenTTL
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

// тест на истечение срока токена и пропустит ли с ним
func TestRegisterLogin_Login_TokenExpired(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	originalTokenTTL := st.Cfg.TokenTTL

	st.Cfg.TokenTTL = time.Second

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	time.Sleep(2 * time.Second)

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NotNil(t, err)

	// Convert the error to a gRPC status error
	statusErr, ok := status.FromError(err)
	require.True(t, ok)

	assert.Nil(t, respLogin)

	// Check that the gRPC status code is Unauthenticated
	require.Equal(t, codes.Unauthenticated, statusErr.Code())

	// Check that the gRPC status message is "token expired or invalid"
	require.Equal(t, "token expired or invalid", statusErr.Message())

	st.Cfg.TokenTTL = originalTokenTTL
}

func TestRegisterLogin_Login_BruteForceAttack(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)

	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
			Email:    email,
			Password: "wrong-password",
			AppId:    appID,
		})
		require.Error(t, err)

		// Check that the error is gRPC error with the correct code and message
		grpcErr, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, grpcErr.Code())
		require.Contains(t, grpcErr.Message(), "invalid email or password")
	}

	// Try to log in with the correct password
	_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})

	// Check that the login attempt is blocked or delayed
	require.Error(t, err)
	require.NotEqual(t, "invalid email or password", err.Error())
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
