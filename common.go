package common

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"strings"

	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var TASK_QUEUE string = "stock-trading"

type Params struct {
	Namespace     string
	GrpcEndpoint  string
	ApiKey        string
	ServerName    string
	EnableMetrics bool
}

func ParseParams(args []string) (Params, error) {
	set := flag.NewFlagSet("worker-scatter-gather", flag.ExitOnError)

	namespace := set.String("namespace", "Default", "Namespace for the server")
	grpcEndpoint := set.String("grpcEndpoint", "us-east-1.aws.api.temporal.io:7233", "Namespace gRPC endpoint")
	apiKey := set.String("apikey", "", "Data plane API key")

	if err := set.Parse(args); err != nil {
		return Params{}, fmt.Errorf("failed parsing args: %w", err)
	} else if *apiKey == "" {
		return Params{}, fmt.Errorf("-apiKey is required")
	}

	serverName := strings.Split(*grpcEndpoint, ":")[0]

	return Params{
		Namespace:    *namespace,
		GrpcEndpoint: *grpcEndpoint,
		ApiKey:       *apiKey,
		ServerName:   serverName,
	}, nil
}

func Connect(params *Params) (client.Client, error) {
	clientOptions := client.Options{
		HostPort:  params.GrpcEndpoint,
		Namespace: params.Namespace,
		Credentials: client.NewAPIKeyDynamicCredentials(
			func(context.Context) (string, error) {
				return params.ApiKey, nil
			},
		),
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         params.ServerName,
			},
			DialOptions: []grpc.DialOption{
				grpc.WithUnaryInterceptor(
					func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						return invoker(
							metadata.AppendToOutgoingContext(ctx, "temporal-namespace", params.Namespace),
							method,
							req,
							reply,
							cc,
							opts...,
						)
					},
				),
			},
		},
	}

	return client.Dial(clientOptions)
}
