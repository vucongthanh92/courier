package user_grpc

import (
	"context"
	"fmt"

	"github.com/vucongthanh92/courier/chat-service/config"
	errorhandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	jwkpb "github.com/vucongthanh92/courier/shared/grpc/user-service/jwk/gen"
	userstatuspb "github.com/vucongthanh92/courier/shared/grpc/user-service/user_status/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserGrpcClient interface {
	GetPublicKey(ctx context.Context, kid string) (string, string, string, *errorhandler.ErrorBuilder)
	CheckUsersStatus(ctx context.Context, userIDs []uint64) ([]uint64, bool, *errorhandler.ErrorBuilder)
}

type grpcClient struct {
	conn      *grpc.ClientConn
	jwkSvc    jwkpb.JWKServiceClient
	statusSvc userstatuspb.UserStatusServiceClient
}

func NewGrpcClient(appCfg *config.AppConfig) UserGrpcClient {
	conn, err := grpc.NewClient(appCfg.Client.UserService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}

	return &grpcClient{
		conn:      conn,
		jwkSvc:    jwkpb.NewJWKServiceClient(conn),
		statusSvc: userstatuspb.NewUserStatusServiceClient(conn),
	}
}

func (c *grpcClient) GetPublicKey(ctx context.Context, kid string) (string, string, string, *errorhandler.ErrorBuilder) {
	resp, err := c.jwkSvc.GetPublicKey(ctx, &jwkpb.GetPublicKeyRequest{Kid: kid})
	if err != nil {
		return "", "", "", errorhandler.InitErrorBuilder(ctx).SetStatus(500).SetLogError(fmt.Errorf("get public key: %w", err))
	}
	return resp.Kid, resp.PublicPem, resp.Alg, nil
}

func (c *grpcClient) CheckUsersStatus(ctx context.Context, userIDs []uint64) ([]uint64, bool, *errorhandler.ErrorBuilder) {
	resp, err := c.statusSvc.CheckUsersStatus(ctx, &userstatuspb.CheckUsersStatusRequest{UserIds: userIDs})
	if err != nil {
		return nil, false, errorhandler.InitErrorBuilder(ctx).SetStatus(500).SetLogError(fmt.Errorf("check users status: %w", err))
	}
	return resp.InvalidUserIds, resp.AllVerified, nil
}
