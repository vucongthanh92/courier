package message

import (
	"context"
	"errors"
	"strings"
	"testing"

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

type memberQueryStub struct {
	member *entities.ConversationMember
}

func (s memberQueryStub) ListConversationMembers(context.Context, uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s memberQueryStub) GetConversationMember(context.Context, uint64, uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	return s.member, nil
}

type messageQueryStub struct {
	byClientID      []*entities.Message
	byClientIDCalls int
	byID            *entities.Message
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

func (s *messageQueryStub) ListMessages(context.Context, uint64, models.ListMessagesRequest) ([]entities.Message, *errHandler.ErrorBuilder) {
	return nil, nil
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

func TestCreateMessageSuccess(t *testing.T) {
	query := &messageQueryStub{}
	command := &messageCommandStub{}
	service := InitMessageUsecase(
		conversationQueryStub{conversation: &entities.Conversation{ID: 10}},
		memberQueryStub{member: &entities.ConversationMember{Status: "active"}},
		query,
		command,
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
			)

			_, _, resultErr := service.CreateMessage(context.Background(), request)

			if resultErr == nil || len(resultErr.Errors) == 0 || resultErr.Errors[0].Code != test.code {
				t.Fatalf("error = %#v, want code %q", resultErr, test.code)
			}
		})
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
