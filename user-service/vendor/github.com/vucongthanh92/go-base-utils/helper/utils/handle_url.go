package utils

import (
	"context"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vucongthanh92/go-base-utils/http/request"
	httpreq "github.com/vucongthanh92/go-base-utils/http/request"
)

func SanitizedOAuthRedirectURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsedURL.RawFragment == "" {
		return parsedURL.String()
	}

	fragment := url.Values{}
	for key, values := range ParseFragmentValues(parsedURL.RawFragment) {
		for _, value := range values {
			fragment.Add(key, value)
		}
	}

	MaskQueryValue(fragment, "oauth_result")
	parsedURL.RawFragment = fragment.Encode()
	return parsedURL.String()
}

func ParseFragmentValues(fragment string) url.Values {
	values, err := url.ParseQuery(fragment)
	if err != nil {
		return url.Values{"fragment": []string{"<unparseable>"}}
	}
	return values
}

func MaskQueryValue(values url.Values, key string) {
	if values.Get(key) == "" {
		return
	}
	values.Set(key, "<redacted>")
}

// SetHeaderByKey sets the header from the gin context to a new context with the specified key.
func SetHeaderByKey(c *gin.Context, key string) context.Context {
	ctx := request.SetHeaderToContext(c, key)
	return context.WithValue(ctx, clientIPContextKey{}, c.ClientIP())
}

// GetUserAgent retrieves the User-Agent from the context.
func GetUserAgent(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if ginCtx, ok := ctx.(*gin.Context); ok {
		if ginCtx.Request != nil {
			return ginCtx.Request.UserAgent()
		}
		return ""
	}

	headers := httpreq.GetHeaderFromContext(ctx, "headers")
	if ua := headers["User-Agent"]; len(ua) > 0 {
		return ua[0]
	}

	return ""
}

// GetClientIP retrieves the client's IP address from the context, checking common proxy headers.
// It checks the following headers in order: X-Forwarded-For, X-Real-IP, True-Client-IP.
func GetClientIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if ginCtx, ok := ctx.(*gin.Context); ok {
		return strings.TrimSpace(ginCtx.ClientIP())
	}

	if clientIP, ok := ctx.Value(clientIPContextKey{}).(string); ok && clientIP != "" {
		return strings.TrimSpace(clientIP)
	}

	// Common proxy headers, ordered by trust/preference
	if xff := GetHeaderFromKey(ctx, "headers", "X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can be a list: client, proxy1, proxy2...
		if idx := strings.Index(xff, ","); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xrip := GetHeaderFromKey(ctx, "headers", "X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	if tcip := GetHeaderFromKey(ctx, "headers", "True-Client-IP"); tcip != "" {
		return strings.TrimSpace(tcip)
	}

	return ""
}

func GetHeaderFromKey(ctx context.Context, key, field string) (resp string) {
	if key == "" || field == "" {
		return resp
	}

	headerMap := httpreq.GetHeaderFromContext(ctx, key)
	if val, existed := headerMap[field]; existed {
		resp = val[0]
	}

	return resp
}
