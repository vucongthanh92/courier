package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
)

type rateLimiterStub struct {
	result interfaces.MessageRateLimitResult
	err    error
}

func (s rateLimiterStub) Allow(context.Context, uint64, uint64) (interfaces.MessageRateLimitResult, error) {
	return s.result, s.err
}

func TestMessageRateLimitMiddlewareRejectsExceededLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST(
		"/api/v1/conversations/:id/messages/create",
		injectClaims("20"),
		MessageRateLimitMiddleware(rateLimiterStub{
			result: interfaces.MessageRateLimitResult{
				Allowed:    false,
				Limit:      5,
				Remaining:  0,
				RetryAfter: 2,
				ResetAt:    100,
			},
		}),
		func(c *gin.Context) { c.Status(http.StatusCreated) },
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/10/messages/create", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
	if response.Header().Get("Retry-After") != "2" {
		t.Fatalf("Retry-After = %q, want 2", response.Header().Get("Retry-After"))
	}
	if response.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("X-RateLimit-Remaining = %q, want 0", response.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestMessageRateLimitMiddlewareFailsOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST(
		"/api/v1/conversations/:id/messages/create",
		injectClaims("20"),
		MessageRateLimitMiddleware(rateLimiterStub{err: errors.New("redis unavailable")}),
		func(c *gin.Context) { c.Status(http.StatusCreated) },
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/10/messages/create", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want fail-open status %d", response.Code, http.StatusCreated)
	}
}

func TestMessageRateLimitMiddlewareRejectsInvalidConversationID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST(
		"/api/v1/conversations/:id/messages/create",
		injectClaims("20"),
		MessageRateLimitMiddleware(rateLimiterStub{}),
		func(c *gin.Context) { c.Status(http.StatusCreated) },
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/not-a-number/messages/create", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func injectClaims(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("authClaims", jwt.MapClaims{"sub": userID})
		c.Next()
	}
}
