package user

import (
	"context"
	"testing"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type userQueryRepoStub struct {
	searchReq models.SearchUsersRequest
	users     []entities.User
}

func (s *userQueryRepoStub) GetUserByIdOrEmail(context.Context, models.GetUserByIdOrEmailRequest) (*entities.User, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s *userQueryRepoStub) GetUsersByIDs(context.Context, []uint64) ([]entities.User, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s *userQueryRepoStub) SearchUsers(_ context.Context, req models.SearchUsersRequest) ([]entities.User, *errHandler.ErrorBuilder) {
	s.searchReq = req
	return s.users, nil
}

func (s *userQueryRepoStub) CheckExistingEmailOrPhone(context.Context, string, string) (bool, *errHandler.ErrorBuilder) {
	return false, nil
}

func TestSearchUsersMapsLimitedProfileFields(t *testing.T) {
	repo := &userQueryRepoStub{
		users: []entities.User{{
			ID:          22,
			DisplayName: "Lan Tran",
			PhoneNumber: "0901234567",
			Email:       "lan@example.com",
			AvatarURL:   "https://example.com/avatar.png",
		}},
	}
	service := InitUserUsecase(repo)

	response, resultErr := service.SearchUsers(context.Background(), models.SearchUsersRequest{
		SearchKey:     "  lan  ",
		ExcludeUserID: 20,
	})
	if resultErr != nil {
		t.Fatalf("SearchUsers returned error: %#v", resultErr)
	}
	if repo.searchReq.SearchKey != "lan" || repo.searchReq.ExcludeUserID != 20 {
		t.Fatalf("request was not normalized/preserved: %#v", repo.searchReq)
	}
	if len(response) != 1 {
		t.Fatalf("response length = %d, want 1", len(response))
	}
	if response[0].UserID != 22 || response[0].DisplayName != "Lan Tran" ||
		response[0].PhoneNumber != "0901234567" || response[0].Email != "lan@example.com" ||
		response[0].Avatar != "https://example.com/avatar.png" {
		t.Fatalf("unexpected response: %#v", response[0])
	}
}

func TestSearchUsersRejectsBlankSearchKey(t *testing.T) {
	service := InitUserUsecase(&userQueryRepoStub{})

	_, resultErr := service.SearchUsers(context.Background(), models.SearchUsersRequest{SearchKey: " "})
	if resultErr == nil {
		t.Fatal("expected blank search key to be rejected")
	}
}
