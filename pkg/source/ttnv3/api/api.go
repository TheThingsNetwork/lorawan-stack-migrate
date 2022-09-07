package api

import (
	"context"
	"os"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"google.golang.org/grpc"
)

var (
	connMu sync.Mutex
	conns  = make(map[string]*grpc.ClientConn)
)

func Dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	connMu.Lock()
	defer connMu.Unlock()
	logger := log.FromContext(ctx).WithField("target", target)
	if conn, ok := conns[target]; ok {
		logger.Debug("Using existing gRPC connection")
		return conn, nil
	}
	logger.Debug("Connecting to gRPC server...")
	startTime := time.Now()
	conn, err := dialContext(ctx, target,
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(rpcmetadata.MD{
			AuthType:      "Bearer",
			AuthValue:     os.Getenv("TTNV3_APP_ACCESS_KEY"),
			AllowInsecure: true,
		}),
	)
	logger.WithField(
		"duration", time.Since(startTime).Round(time.Microsecond*100),
	).Debug("Connected to gRPC server")
	if err != nil {
		return nil, err
	}
	conns[target] = conn
	return conn, nil
}

func dialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(rpcclient.DefaultDialOptions(ctx), opts...)
	return grpc.DialContext(ctx, target, opts...)
}

func CloseConnections() error {
	for _, c := range conns {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}
