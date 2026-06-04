package request

import (
	"context"
	"net/http"

	jwtToken "github.com/vucongthanh92/go-base-utils/token"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

const UserContextKey = "usercontext"

type UserContext struct {
	Id    int64 `mapstructure:"_id"`
	OrgId int64 `mapstructure:"_orgId"`
}

func GetUserContext(c *gin.Context) UserContext {
	return c.MustGet(UserContextKey).(UserContext)
}

func SetUserContext(c *gin.Context, userContext *UserContext) {
	c.Set(UserContextKey, *userContext)
}

func MustGetUser(c *gin.Context) (user UserContext) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		panic("missing token")
	}
	token := authToken[len("Bearer "):]
	claims, err := jwtToken.ParseTokenUnverify(token)
	if err != nil {
		panic("invalid token")
	}
	if err := mapstructure.WeakDecode(claims, &user); err != nil {
		panic("can't decode token")
	}
	return
}

// function set and get header from context
type headerKeyType string

func SetHeaderToContext(c *gin.Context, key string) context.Context {
	var (
		headersVal = c.Request.Header
		headerKey  = headerKeyType(key)
	)
	return context.WithValue(c.Request.Context(), headerKey, headersVal)
}

func GetHeaderFromContext(ctx context.Context, key string) map[string][]string {
	var (
		resp      = make(map[string][]string)
		headerKey = headerKeyType(key)
	)

	if headers, existed := ctx.Value(headerKey).(http.Header); existed {
		if len(headers) == 0 {
			return resp
		}

		for key, value := range headers {
			resp[key] = value
		}
	}

	return resp
}
