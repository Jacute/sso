package suite

import (
	"context"
	"net"
	"os"
	"sso/internal/config"
	"strconv"
	"testing"

	ssov1 "github.com/jacute/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Config     *config.Config
	AuthClient ssov1.AuthClient
}

const (
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	v := os.Getenv("CONFIG_PATH")
	if v == "" {
		t.Fatal("CONFIG_PATH environment variable not set")
	}

	cfg := config.MustLoadByPath(v)

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	var cc *grpc.ClientConn

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
		cc.Close()
	})
	cc, err := grpc.NewClient(
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("gRPC server connection error: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Config:     cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
