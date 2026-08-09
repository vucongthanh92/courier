package user

import (
	"context"
	"net/http"
	"strings"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type userUsecase struct {
	userQueryRepo interfaces.UserQueryRepoI
}

func InitUserUsecase(userQueryRepo interfaces.UserQueryRepoI) interfaces.UserServiceI {
	return &userUsecase{userQueryRepo: userQueryRepo}
}

func (s *userUsecase) SearchUsers(ctx context.Context, req models.SearchUsersRequest) ([]models.SearchUserResponse, *errHandler.ErrorBuilder) {
	req.SearchKey = strings.TrimSpace(req.SearchKey)
	if req.SearchKey == "" {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "invalid_search_key", Field: "search_key", Message: "search_key is required"})
	}

	users, repoErr := s.userQueryRepo.SearchUsers(ctx, req)
	if repoErr != nil {
		return nil, repoErr
	}

	response := make([]models.SearchUserResponse, 0, len(users))
	for _, item := range users {
		response = append(response, models.SearchUserResponse{
			UserID:      item.ID,
			DisplayName: item.DisplayName,
			PhoneNumber: item.PhoneNumber,
			Email:       item.Email,
			Avatar:      item.AvatarURL,
		})
	}

	return response, nil
}
