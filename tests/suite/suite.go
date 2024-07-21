package suite

import (
	"STTAuth/internal/config"
	"context"
	"net"
	"strconv"
	"testing"

	ssov1 "github.com/skinkvi/protosSTT/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

const (
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()   // текущая ф-ия это хелпер
	t.Parallel() // паралельные тесты

	// Если использовать какой нибудь Gitlab actions то нужно будет использовать переменную окружения
	cfg := config.MustLoadByPath("../config/local.yaml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	// grpc client test
	//
	//Мы используем не безопастное соединение (ну потому что простенький пет проект) в тестах возиться не хотим
	cc, err := grpc.DialContext(context.Background(), grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connetion falied: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
