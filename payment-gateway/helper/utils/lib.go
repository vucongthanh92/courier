package utils

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	mrand "math/rand"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	httpreq "github.com/vucongthanh92/go-base-utils/http/request"
	utils "github.com/vucongthanh92/go-base-utils/http/request"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

// func Reverse reverses a string while properly handling UTF-8 encoded characters.
func Reverse(s string) (string, error) {
	if !utf8.ValidString(s) {
		return s, errors.New("input is not valid UTF-8")
	}
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r), nil
}

// func GetHashMD5 generates an MD5 hash of the input string and returns it as a hexadecimal string.
func GetHashMD5(input string) string {
	sum := md5.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}

// GetToken generates a SHA-256 hash of the concatenation of a secret and the provided ID,
// returning it as an uppercase hexadecimal string.
func GetToken(id string) string {
	secret := os.Getenv("PASSWORD_SECRET")
	hashMethod := sha256.New()
	hashMethod.Write([]byte(secret + id))
	hash := hashMethod.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(hash))
}

// func LowerInitial takes a slice of strings and returns a new slice where the first character of each string is converted to lowercase.
func LowerInitial(fields []string) (results []string) {
	for _, str := range fields {
		for j, val := range str {
			results = append(results, string(unicode.ToLower(val))+str[j+1:])
			break
		}
	}
	return results
}

// func Contains checks if a slice of comparable elements contains a specific element, returning true if found and false otherwise.
func Contains[T comparable](s []T, str T) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// func RemoveDuplicate removes duplicate elements from a slice of comparable types and returns a new slice with unique elements.
func RemoveDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// func FillTemplate replaces placeholders in a JSON template string with corresponding values from a provided map, returning the filled JSON string.
func FillTemplate(templateJSON string, data map[string]string) string {
	filledJSON := templateJSON
	for key, value := range data {
		placeholder := "${" + key + "}"
		filledJSON = strings.ReplaceAll(filledJSON, placeholder, value)
	}
	return filledJSON
}

// func SliceToMap converts a slice of any type into a map, using a provided function to determine the key for each element.
func SliceToMap[T any, K comparable](arr []T, f func(item T) (K, T)) map[K]T {
	result := make(map[K]T)
	for _, item := range arr {
		key, mapped := f(item)
		result[key] = mapped
	}
	return result
}

// func IterateSlice iterates over a slice of any type, applying a provided function to each element along with its index.
func IterateSlice[T any](params []T, f func(i int, item T)) {
	for i, item := range params {
		f(i, item)
	}
}

// func ConvertStrToStruct converts a JSON string into a struct of any type, returning the struct and any error encountered during unmarshalling.
func ConvertStrToStruct[T any](param string) (T, error) {
	var resp T
	err := json.Unmarshal([]byte(param), &resp)
	return resp, err
}

// func IterateMap iterates over a map of any key and value types, applying a provided function to each key-value pair.
func IterateMap[T comparable, K any](params map[T]K, f func(i T, item K)) {
	for i, item := range params {
		f(i, item)
	}
}

// func ConvertStructToStr converts a struct of any type into a JSON string, returning the string and any error encountered during marshalling.
func ConvertStructToStr[T any](param T) (string, error) {
	buff, err := json.Marshal(param)
	if err != nil {
		return "", err
	}
	return string(buff), nil
}

// func CheckDuplicateTypeByField checks for duplicate values in a slice of structs based on a specified field name.
func CheckDuplicateTypeByField[T any](options []T, field string) (err error, isDuplicate bool) {
	allKeys := make(map[interface{}]bool)
	for _, item := range options {
		val := reflect.ValueOf(item)
		fieldValue := val.FieldByName(field)
		if !fieldValue.IsValid() {
			return fmt.Errorf("field %s does not exist in struct", field), false
		}
		key := fieldValue.Interface()
		if _, exists := allKeys[key]; exists {
			return nil, true
		}
		allKeys[key] = true
	}
	return nil, false
}

// func GetHeaderFromKey retrieves a specific header value from the context using the provided key.
func GetHeaderFromKey(ctx context.Context, key, field string) (resp string) {
	headers := httpreq.GetHeaderFromContext(ctx, key)
	if values := headers[field]; len(values) > 0 {
		return values[0]
	}
	return ""
}

func SafeGo(f func()) {
	go func() {
		defer HandlePanic()
		f()
	}()
}

func HandlePanic() {
	if r := recover(); r != nil {
		logger.Error("Recovered from panic: ", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
	}
}

// HashPwdBySha256 hashes the password using
func HashPwdBySha256(email, password string) string {
	secret := viper.GetString("authenticate.passwordHashSecret")
	hashMethod := sha256.New()
	hashMethod.Write([]byte(secret + email + password))
	hash := hashMethod.Sum(nil)
	result := strings.ToUpper(hex.EncodeToString(hash))
	return result
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

	headers := httpreq.GetHeaderFromContext(ctx, "headers")
	if ua := headers["X-Request-Id"]; len(ua) > 0 {
		return ua[0]
	}

	return ""
}

// RandString generates a random string of the specified length using crypto/rand for better randomness.
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	mrand.Seed(time.Now().UnixNano())
}

func RandString(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, n)
	max := big.NewInt(int64(len(letters)))

	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err == nil {
			buf[i] = letters[v.Int64()]
			continue
		}
		// Fallback: non-crypto random
		buf[i] = letters[mrand.Intn(len(letters))]
	}
	return string(buf)
}

// StrPtr returns a pointer to the given string, or nil if the string is empty.
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Uint64Ptr returns a pointer to the given uint64, or nil if the value is 0.
func Uint64Ptr(s uint64) *uint64 {
	if s == 0 {
		return nil
	}
	return &s
}

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// SetHeaderByKey sets the header from the gin context to a new context with the specified key.
func SetHeaderByKey(c *gin.Context, key string) context.Context {
	return utils.SetHeaderToContext(c, key)
}

// ParseUserID converts the sub claim to uint64 safely.
func ParseUserID(sub any) uint64 {
	switch v := sub.(type) {
	case string:
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	case float64:
		return uint64(v)
	case json.Number:
		id, _ := v.Int64()
		return uint64(id)
	default:
		return 0
	}
}

// GetHeaderFromKey retrieves a specific header value from the context using the provided key.
func GenerateConversationDirectKey(memberUserIDs []uint64) string {
	if len(memberUserIDs) == 0 {
		return ""
	}

	ids := append([]uint64(nil), memberUserIDs...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(id, 10))
	}

	canonical := strings.Join(parts, ":")
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func GenerateSystemConversationDirectKey(userID uint64, name string) string {
	if userID == 0 || strings.TrimSpace(name) == "" {
		return ""
	}
	canonical := fmt.Sprintf("system:%d:%s", userID, strings.TrimSpace(name))
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// func NormalizeMemberIDs normalizes a slice of member IDs by removing duplicates and ensuring all IDs are greater than 0.
func NormalizeMemberIDs(ids []uint64) ([]uint64, error) {

	if len(ids) == 0 {
		return nil, errors.New("member_user_ids is required")
	}

	seen := make(map[uint64]struct{}, len(ids))
	out := make([]uint64, 0, len(ids))
	for _, id := range ids {
		_, isExist := seen[id]
		switch {
		case id <= 0:
			return nil, errors.New("member user id must be greater than 0")
		case isExist:
			continue
		default:
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})

	return out, nil
}
