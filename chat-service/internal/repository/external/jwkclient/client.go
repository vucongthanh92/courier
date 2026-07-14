package jwkclient

import (
	"context"
	"fmt"
	"time"

	errorhandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	jwkpb "github.com/vucongthanh92/courier/shared/grpc/user-service/jwk/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	GetPublicKey(ctx context.Context, kid string) (string, string, string, *errorhandler.ErrorBuilder)
}

type client struct {
	conn *grpc.ClientConn
	svc  jwkpb.JWKServiceClient
}

func New(address string) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return &client{
		conn: conn,
		svc:  jwkpb.NewJWKServiceClient(conn),
	}, nil
}

func (c *client) GetPublicKey(ctx context.Context, kid string) (string, string, string, *errorhandler.ErrorBuilder) {
	resp, err := c.svc.GetPublicKey(ctx, &jwkpb.GetPublicKeyRequest{Kid: kid})
	if err != nil {
		return "", "", "", errorhandler.InitErrorBuilder(ctx).SetStatus(500).SetLogError(fmt.Errorf("get public key: %w", err))
	}
	return resp.Kid, resp.PublicPem, resp.Alg, nil
}
