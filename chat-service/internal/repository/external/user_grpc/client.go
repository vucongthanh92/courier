package user_grpc

import (
	"context"
	"fmt"

	"github.com/vucongthanh92/courier/chat-service/config"
	errorhandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	jwkpb "github.com/vucongthanh92/courier/shared/grpc/user-service/jwk/gen"
	userprofilepb "github.com/vucongthanh92/courier/shared/grpc/user-service/user_profile/gen"
	userstatuspb "github.com/vucongthanh92/courier/shared/grpc/user-service/user_status/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserGrpcClient interface {
	GetPublicKey(ctx context.Context, kid string) (string, string, string, *errorhandler.ErrorBuilder)
	CheckUsersStatus(ctx context.Context, userIDs []uint64) ([]uint64, bool, *errorhandler.ErrorBuilder)
	BatchGetUserProfiles(ctx context.Context, userIDs []uint64) ([]models.UserProfileSummaryResponse, *errorhandler.ErrorBuilder)
}

type grpcClient struct {
	conn       *grpc.ClientConn
	jwkSvc     jwkpb.JWKServiceClient
	profileSvc userprofilepb.UserProfileServiceClient
	statusSvc  userstatuspb.UserStatusServiceClient
}

func NewGrpcClient(appCfg *config.AppConfig) UserGrpcClient {
	conn, err := grpc.NewClient(appCfg.Client.UserService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}

	return &grpcClient{
		conn:       conn,
		jwkSvc:     jwkpb.NewJWKServiceClient(conn),
		profileSvc: userprofilepb.NewUserProfileServiceClient(conn),
		statusSvc:  userstatuspb.NewUserStatusServiceClient(conn),
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

func (c *grpcClient) BatchGetUserProfiles(ctx context.Context, userIDs []uint64) ([]models.UserProfileSummaryResponse, *errorhandler.ErrorBuilder) {
	resp, err := c.profileSvc.BatchGetUserProfiles(ctx, &userprofilepb.BatchGetUserProfilesRequest{UserIds: userIDs})
	if err != nil {
		return nil, errorhandler.InitErrorBuilder(ctx).SetStatus(500).SetLogError(fmt.Errorf("batch get user profiles: %w", err))
	}

	profiles := make([]models.UserProfileSummaryResponse, 0, len(resp.Users))
	for _, user := range resp.Users {
		profiles = append(profiles, models.UserProfileSummaryResponse{
			UserID:      user.UserId,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarUrl,
			Status:      user.Status,
		})
	}
	return profiles, nil
}
