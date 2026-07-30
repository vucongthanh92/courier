package message

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type conversationQueryStub struct {
	conversation *entities.Conversation
}

func (s conversationQueryStub) GetDirectConversationByKey(context.Context, string) (*entities.Conversation, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s conversationQueryStub) GetConversationByID(context.Context, uint64) (*entities.Conversation, *errHandler.ErrorBuilder) {
	return s.conversation, nil
}

func (s conversationQueryStub) ListConversationsByMember(context.Context, *models.ListConversationsRequest) ([]models.ConversationListResponse, *errHandler.ErrorBuilder) {
	return nil, nil
}

type memberQueryStub struct {
	member  *entities.ConversationMember
	members []entities.ConversationMember
}

func (s memberQueryStub) ListConversationMembers(context.Context, uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder) {
	return s.members, nil
}

func (s memberQueryStub) GetConversationMember(context.Context, uint64, uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	return s.member, nil
}

type messageQueryStub struct {
	byClientID      []*entities.Message
	byClientIDCalls int
	byID            *entities.Message
	listMessages    []entities.Message
	listRequest     models.ListMessagesRequest
}

func (s *messageQueryStub) GetMessageByClientMessageID(context.Context, uint64, string) (*entities.Message, *errHandler.ErrorBuilder) {
	if s.byClientIDCalls >= len(s.byClientID) {
		s.byClientIDCalls++
		return nil, nil
	}
	result := s.byClientID[s.byClientIDCalls]
	s.byClientIDCalls++
	return result, nil
}

func (s *messageQueryStub) ListMessages(_ context.Context, _ uint64, req models.ListMessagesRequest) ([]entities.Message, *errHandler.ErrorBuilder) {
	s.listRequest = req
	if len(s.listMessages) == 0 {
		return nil, nil
	}
	return s.listMessages, nil
}

func (s *messageQueryStub) GetMessageByID(context.Context, uint64) (*entities.Message, *errHandler.ErrorBuilder) {
	return s.byID, nil
}

type messageCommandStub struct {
	created *entities.Message
	err     *errHandler.ErrorBuilder
}

func (s *messageCommandStub) CreateMessage(_ context.Context, entity *entities.Message) *errHandler.ErrorBuilder {
	s.created = entity
	return s.err
}

type messageListCacheStub struct {
	page            *models.CachedMessageListPage
	setPage         *models.CachedMessageListPage
	invalidated     uint64
	getCalls        int
	setCalls        int
	invalidateCalls int
}

type realtimePublisherStub struct {
	event *models.MessageCreatedEvent
	calls int
}

func (s *realtimePublisherStub) PublishMessageCreated(_ context.Context, event models.MessageCreatedEvent) error {
	s.calls++
	s.event = &event
	return nil
}

func (s *messageListCacheStub) GetLatest(context.Context, uint64, int) (*models.CachedMessageListPage, error) {
	s.getCalls++
	return s.page, nil
}

func (s *messageListCacheStub) SetLatest(_ context.Context, _ uint64, _ int, page models.CachedMessageListPage, _ time.Duration) error {
	s.setCalls++
	s.setPage = &page
	return nil
}

func (s *messageListCacheStub) InvalidateLatest(_ context.Context, conversationID uint64) error {
	s.invalidateCalls++
	s.invalidated = conversationID
	return nil
}

func TestCreateMessageSuccess(t *testing.T) {
	query := &messageQueryStub{}
	command := &messageCommandStub{}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
		query,
		command,
		&messageListCacheStub{},
		nil,
	)
	clientID := " client-1 "
	request := &models.SendMessageRequest{
		Type:             "text",
		Body:             " hello ",
		ClientMessageID:  &clientID,
		Metadata:         map[string]any{},
		SenderID:         20,
		ConversationID:   10,
		ReplyToMessageID: nil,
	}

	response, created, resultErr := service.CreateMessage(context.Background(), request)

	if resultErr != nil {
		t.Fatalf("CreateMessage() error = %#v", resultErr)
	}
	if !created {
		t.Fatal("CreateMessage() created = false, want true")
	}
	if response.Body != "hello" {
		t.Fatalf("response body = %q, want trimmed body", response.Body)
	}
	if response.SenderID != 20 || response.ConversationID != 10 {
		t.Fatalf("unexpected response identity: %#v", response)
	}
	if command.created == nil || command.created.ID == 0 {
		t.Fatal("message command did not receive initialized entity")
	}
	if command.created.ClientMessageID == nil || *command.created.ClientMessageID != "client-1" {
		t.Fatalf("client_message_id was not normalized: %#v", command.created.ClientMessageID)
	}
}

func TestCreateMessageRejectsInactiveMember(t *testing.T) {
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "left"}},
		&messageQueryStub{},
		&messageCommandStub{},
		nil,
		nil,
	)

	_, _, resultErr := service.CreateMessage(context.Background(), validRequest())

	assertError(t, resultErr, 403, "not_active_conversation_member")
}

func TestCreateMessageReturnsIdempotentMessage(t *testing.T) {
	clientID := "client-1"
	existing := &entities.Message{
		ID:              99,
		ConversationID:  10,
		SenderID:        20,
		Type:            "text",
		Body:            "already created",
		ClientMessageID: &clientID,
		Metadata:        map[string]any{},
	}
	command := &messageCommandStub{}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
		&messageQueryStub{byClientID: []*entities.Message{existing}},
		command,
		&messageListCacheStub{},
		nil,
	)
	request := validRequest()
	request.ClientMessageID = &clientID

	response, created, resultErr := service.CreateMessage(context.Background(), request)

	if resultErr != nil {
		t.Fatalf("CreateMessage() error = %#v", resultErr)
	}
	if created {
		t.Fatal("idempotent response marked as newly created")
	}
	if response.ID != existing.ID {
		t.Fatalf("response ID = %d, want %d", response.ID, existing.ID)
	}
	if command.created != nil {
		t.Fatal("idempotent request unexpectedly inserted a message")
	}
}

func TestCreateMessageRecoversConcurrentIdempotencyConflict(t *testing.T) {
	clientID := "client-1"
	existing := &entities.Message{
		ID:              99,
		ConversationID:  10,
		SenderID:        20,
		Type:            "text",
		Body:            "created concurrently",
		ClientMessageID: &clientID,
		Metadata:        map[string]any{},
	}
	commandErr := errHandler.InitErrorBuilder(context.Background()).
		SetStatus(500).
		SetLogError(errors.New("duplicate key value violates unique constraint (SQLSTATE 23505)"))
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
		&messageQueryStub{byClientID: []*entities.Message{nil, existing}},
		&messageCommandStub{err: commandErr},
		nil,
		nil,
	)
	request := validRequest()
	request.ClientMessageID = &clientID

	response, created, resultErr := service.CreateMessage(context.Background(), request)

	if resultErr != nil || created || response.ID != existing.ID {
		t.Fatalf("unexpected conflict recovery: response=%#v created=%v error=%#v", response, created, resultErr)
	}
}

func TestCreateMessageRejectsReplyFromAnotherConversation(t *testing.T) {
	replyID := uint64(44)
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
		&messageQueryStub{byID: &entities.Message{ID: replyID, ConversationID: 11}},
		&messageCommandStub{},
		nil,
		nil,
	)
	request := validRequest()
	request.ReplyToMessageID = &replyID

	_, _, resultErr := service.CreateMessage(context.Background(), request)

	assertError(t, resultErr, 400, "reply_message_conversation_mismatch")
}

func TestCreateMessageValidation(t *testing.T) {
	tests := []struct {
		name string
		edit func(*models.SendMessageRequest)
		code string
	}{
		{name: "unsupported type", edit: func(r *models.SendMessageRequest) { r.Type = "system" }, code: "invalid_message_type"},
		{name: "blank body", edit: func(r *models.SendMessageRequest) { r.Body = " \n " }, code: "empty_message_body"},
		{name: "body too long", edit: func(r *models.SendMessageRequest) { r.Body = strings.Repeat("a", 4001) }, code: "message_body_too_long"},
		{name: "missing sender", edit: func(r *models.SendMessageRequest) { r.SenderID = 0 }, code: "unauthorized"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validRequest()
			test.edit(request)
			service := InitMessageUsecase(
				conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
				memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
				&messageQueryStub{},
				&messageCommandStub{},
				nil,
				nil,
			)

			_, _, resultErr := service.CreateMessage(context.Background(), request)

			if resultErr == nil || len(resultErr.Errors) == 0 || resultErr.Errors[0].Code != test.code {
				t.Fatalf("error = %#v, want code %q", resultErr, test.code)
			}
		})
	}
}

func TestCreateMessageInvalidatesLatestMessageCache(t *testing.T) {
	cache := &messageListCacheStub{}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
		&messageQueryStub{},
		&messageCommandStub{},
		cache,
		nil,
	)

	_, created, resultErr := service.CreateMessage(context.Background(), validRequest())

	if resultErr != nil || !created {
		t.Fatalf("CreateMessage() created=%v error=%#v", created, resultErr)
	}
	if cache.invalidateCalls != 1 || cache.invalidated != 10 {
		t.Fatalf("cache invalidation = calls:%d conversation:%d", cache.invalidateCalls, cache.invalidated)
	}
}

func TestCreateMessagePublishesRealtimeEventToActiveMembers(t *testing.T) {
	publisher := &realtimePublisherStub{}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{
			member: &entities.ConversationMember{Status: "active"},
			members: []entities.ConversationMember{
				{ID: 1, ConversationID: 10, UserID: 20, Status: "active"},
				{ID: 2, ConversationID: 10, UserID: 21, Status: "active"},
				{ID: 3, ConversationID: 10, UserID: 22, Status: "left"},
			},
		},
		&messageQueryStub{},
		&messageCommandStub{},
		nil,
		publisher,
	)

	response, created, resultErr := service.CreateMessage(context.Background(), validRequest())

	if resultErr != nil || !created {
		t.Fatalf("CreateMessage() created=%v error=%#v", created, resultErr)
	}
	if publisher.calls != 1 || publisher.event == nil {
		t.Fatalf("publisher was not called: %#v", publisher)
	}
	if publisher.event.Type != models.RealtimeEventMessageCreated ||
		publisher.event.ConversationID != 10 ||
		publisher.event.Message.ID != response.ID {
		t.Fatalf("unexpected realtime event: %#v", publisher.event)
	}
	if len(publisher.event.RecipientUserIDs) != 2 ||
		publisher.event.RecipientUserIDs[0] != 20 ||
		publisher.event.RecipientUserIDs[1] != 21 {
		t.Fatalf("unexpected recipients: %#v", publisher.event.RecipientUserIDs)
	}
}

func TestListMessagesReturnsMembersAndCachesLatestPage(t *testing.T) {
	query := &messageQueryStub{
		listMessages: []entities.Message{
			{ID: 100, ConversationID: 10, SenderID: 20, Type: "text", Body: "newer", Metadata: map[string]any{}},
			{ID: 99, ConversationID: 10, SenderID: 21, Type: "text", Body: "older", Metadata: map[string]any{}},
			{ID: 98, ConversationID: 10, SenderID: 22, Type: "text", Body: "extra", Metadata: map[string]any{}},
		},
	}
	cache := &messageListCacheStub{}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{
			member: &entities.ConversationMember{Status: "active"},
			members: []entities.ConversationMember{
				{ID: 1, ConversationID: 10, UserID: 20, Role: "owner", Status: "active"},
				{ID: 2, ConversationID: 10, UserID: 21, Role: "member", Status: "active"},
			},
		},
		query,
		&messageCommandStub{},
		cache,
		nil,
	)
	request := &models.ListMessagesRequest{ConversationID: 10, RequesterID: 20, Limit: 2}

	response, resultErr := service.ListMessages(context.Background(), request)

	if resultErr != nil {
		t.Fatalf("ListMessages() error = %#v", resultErr)
	}
	if len(response.Messages) != 2 || !response.Pagination.HasMore {
		t.Fatalf("unexpected pagination response: %#v", response)
	}
	if response.Pagination.NextBeforeMessageID == nil || *response.Pagination.NextBeforeMessageID != 99 {
		t.Fatalf("next_before_message_id = %#v, want 99", response.Pagination.NextBeforeMessageID)
	}
	if len(response.Members) != 2 || response.Members[0].UserID != 20 {
		t.Fatalf("members were not mapped: %#v", response.Members)
	}
	if query.listRequest.Limit != 3 {
		t.Fatalf("query limit = %d, want limit+1", query.listRequest.Limit)
	}
	if cache.getCalls != 1 || cache.setCalls != 1 || cache.setPage == nil {
		t.Fatalf("cache not used as expected: %#v", cache)
	}
}

func TestListMessagesUsesCachedLatestPage(t *testing.T) {
	cache := &messageListCacheStub{
		page: &models.CachedMessageListPage{
			Messages: []models.MessageResponse{
				{ID: 100, ConversationID: 10, SenderID: 20, Type: "text", Body: "cached", Metadata: map[string]any{}},
			},
			Pagination: models.MessagePaginationResponse{Limit: 20, HasMore: false},
		},
	}
	query := &messageQueryStub{}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{
			member:  &entities.ConversationMember{Status: "active"},
			members: []entities.ConversationMember{{ID: 1, ConversationID: 10, UserID: 20, Role: "owner", Status: "active"}},
		},
		query,
		&messageCommandStub{},
		cache,
		nil,
	)

	response, resultErr := service.ListMessages(context.Background(), &models.ListMessagesRequest{
		ConversationID: 10,
		RequesterID:    20,
	})

	if resultErr != nil {
		t.Fatalf("ListMessages() error = %#v", resultErr)
	}
	if len(response.Messages) != 1 || response.Messages[0].Body != "cached" {
		t.Fatalf("cached response not returned: %#v", response.Messages)
	}
	if query.listRequest.Limit != 0 {
		t.Fatalf("DB query was called despite cache hit: %#v", query.listRequest)
	}
	if len(response.Members) != 1 {
		t.Fatalf("members should still be loaded with cached messages: %#v", response.Members)
	}
}

func TestListMessagesSkipsCacheForCursorPage(t *testing.T) {
	beforeID := uint64(100)
	cache := &messageListCacheStub{
		page: &models.CachedMessageListPage{
			Messages: []models.MessageResponse{{ID: 1, Body: "should not use"}},
		},
	}
	query := &messageQueryStub{
		listMessages: []entities.Message{{ID: 99, ConversationID: 10, SenderID: 20, Type: "text", Body: "from db", Metadata: map[string]any{}}},
	}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
		query,
		&messageCommandStub{},
		cache,
		nil,
	)

	response, resultErr := service.ListMessages(context.Background(), &models.ListMessagesRequest{
		ConversationID:  10,
		RequesterID:     20,
		Limit:           20,
		BeforeMessageID: &beforeID,
	})

	if resultErr != nil {
		t.Fatalf("ListMessages() error = %#v", resultErr)
	}
	if response.Messages[0].Body != "from db" {
		t.Fatalf("cursor page should bypass latest cache: %#v", response.Messages)
	}
	if cache.getCalls != 0 || cache.setCalls != 0 {
		t.Fatalf("cache should not be used for cursor pages: %#v", cache)
	}
}

func validRequest() *models.SendMessageRequest {
	return &models.SendMessageRequest{
		Type:           "text",
		Body:           "hello",
		Metadata:       map[string]any{},
		SenderID:       20,
		ConversationID: 10,
	}
}

func assertError(t *testing.T, resultErr *errHandler.ErrorBuilder, status int, code string) {
	t.Helper()
	if resultErr == nil {
		t.Fatalf("expected error %q, got nil", code)
	}
	if resultErr.Status != status {
		t.Fatalf("status = %d, want %d", resultErr.Status, status)
	}
	if len(resultErr.Errors) == 0 || resultErr.Errors[0].Code != code {
		t.Fatalf("errors = %#v, want code %q", resultErr.Errors, code)
	}
}
