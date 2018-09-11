package main

import (
	"context"
	"fmt"
	"net"
	"time"

	types "github.com/jessfraz/bpfd/api/grpc"
	"google.golang.org/grpc"
)

func getClient(ctx context.Context, address string) (types.APIClient, error) {
	// TODO: have more dial options
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	dialOpts = append(dialOpts,
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		},
		))

	conn, err := grpc.Dial(address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating grpc connection to %s failed: %v", address, err)
	}

	return types.NewAPIClient(conn), nil
}
